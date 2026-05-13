# Auth CAS Ticket Validator Phase 5 Slice 16 Implementation Plan

## Objective

继续 phase 5 收窄 `auth` application concrete allowlist，把 CAS ticket 的 HTTP 校验下沉到模块 port / infrastructure adapter：

- 去掉 `auth/application/commands/cas_service.go -> net/http`
- 保持 CAS callback 的用户同步、自动开通、锁定解锁和 session 签发行为不变

## Non-goals

- 不改 `auth/application/queries/cas_service.go` 的登录跳转 URL 生成逻辑
- 不改 CAS callback 的 HTTP API 契约、cookie 写入或 audit 行为
- 不引入新的认证 provider、多协议抽象或跨模块 contract

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/auth/application/commands/cas_service.go`
- `code/backend/internal/module/auth/application/commands/cas_service_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/module/auth/runtime/module.go`

## Current Baseline

- `auth/application/commands/cas_service.go` 当前直接持有 `*http.Client`
- CAS validate URL 由 application 拼装，随后在同一文件里发起 HTTP request、解析 XML、校验 principal，再继续做用户同步和 session 签发
- `auth/runtime/module.go` 和测试 helper 都直接把 `nil` / concrete `http.Client` 传给 command service
- allowlist 当前保留：
  - `auth/application/commands/cas_service.go -> net/http`

## Chosen Direction

把 CAS ticket 校验表达成 auth 自己的 validator port：

1. 在 `auth/ports` 新增 `CASTicketValidator`、`CASPrincipal` 与 invalid ticket sentinel
2. `auth/application/commands/cas_service.go` 保留 validate URL 构造、错误码映射、用户同步和 session 签发，只通过 port 调用 validator
3. `auth/infrastructure` 提供 HTTP validator，统一承接 request、XML 解析、principal attribute 提取和用户名校验
4. `auth/runtime/module.go` 与测试 helper 统一构建 validator，再注入 CAS command service

## Ownership Boundary

- `auth/application/commands/cas_service.go`
  - 负责：CAS callback 的用例编排、validate URL 生成、错误码映射、用户同步和 session 签发
  - 不负责：知道 HTTP client、XML response、CAS response 解析或用户名校验细节
- `auth/infrastructure/cas_ticket_validator.go`
  - 负责：向 CAS validate endpoint 发请求、解析 XML、抽取 principal 并区分 invalid ticket
  - 不负责：用户同步、自动开通、登录失败码映射或 session 签发
- `auth/runtime/module.go`
  - 负责：装配 validator 并传给 CAS command service
  - 不负责：把 `net/http` concrete 留在 auth application surface

## Change Surface

- Add: `.harness/reuse-decisions/auth-cas-ticket-validator-phase5-slice16.md`
- Add: `docs/plan/impl-plan/2026-05-13-auth-cas-ticket-validator-phase5-slice16-implementation-plan.md`
- Add: `code/backend/internal/module/auth/ports/cas_ticket_validator.go`
- Add: `code/backend/internal/module/auth/ports/cas_ticket_validator_context_contract_test.go`
- Add: `code/backend/internal/module/auth/infrastructure/cas_ticket_validator.go`
- Add: `code/backend/internal/module/auth/infrastructure/cas_ticket_validator_test.go`
- Modify: `code/backend/internal/module/auth/application/commands/cas_service.go`
- Modify: `code/backend/internal/module/auth/application/commands/cas_service_test.go`
- Modify: `code/backend/internal/module/auth/api/http/handler.go`
- Modify: `code/backend/internal/module/auth/api/http/http_integration_test.go`
- Modify: `code/backend/internal/module/auth/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [ ] Slice 1: 提取 CAS ticket validator port 与 HTTP adapter
  - 目标：auth command service 不再直接持有 `net/http` concrete，CAS callback 行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/auth/infrastructure -run 'CASTicketValidator' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/auth/application/commands -run 'CASService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/auth/api/http -run 'CASCallback' -count=1 -timeout 5m`
  - Review focus：invalid ticket / service unavailable 的错误码映射是否保持一致；runtime 和 handler 默认 wiring 是否都切到 validator port

- [ ] Slice 2: 删除 allowlist 并同步文档
  - 目标：删掉 `auth/application/commands/cas_service.go -> net/http`，phase5 当前事实同步更新
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除本次 CAS 实际收口的 allowlist；文档正确反映 auth 仍剩余或已清空的 concrete 依赖范围

## Risks

- 如果 invalid ticket sentinel 映射不准，CAS callback 会从 业务错误退化成 503
- 如果 runtime / handler 的默认 wiring 没有同步，CAS callback 可能误报 `ErrCASNotImplemented`
- 如果 principal attribute 解析迁移时丢字段，自动开通和资料同步会发生行为漂移

## Verification Plan

1. `cd code/backend && go test ./internal/module/auth/infrastructure -run 'CASTicketValidator' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/auth/application/commands -run 'CASService' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/auth/api/http -run 'CASCallback' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`
8. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：CAS callback 业务编排留在 auth command service，网络校验和 XML 解析落回 infrastructure
- reuse point 明确：沿用 phase5 已验证的 `application -> ports -> infrastructure -> runtime wiring` 模式，不再追加临时 helper
- 这刀同时解决行为与结构：保留 CAS 登录行为，同时删掉 auth application surface 的 `net/http` concrete 例外
