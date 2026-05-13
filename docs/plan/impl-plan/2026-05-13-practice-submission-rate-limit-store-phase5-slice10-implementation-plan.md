# Practice Submission Rate Limit Store Phase 5 Slice 10 Implementation Plan

## Objective

继续 phase 5 收窄 `practice` application concrete allowlist，把 flag submit 限流的 Redis 依赖下沉到模块 port / infrastructure adapter：

- 去掉 `practice/application/commands/service.go -> github.com/redis/go-redis/v9`
- 保持 flag submit 窗口计数、首次提交设置窗口和超频拒绝提交的现有行为不变

## Non-goals

- 不改 `practice/application/commands/submission_service.go` 的 flag 校验、manual review 或 solve grace 业务规则
- 不处理 `assessment`、`ops` 等其他模块剩余 Redis concrete allowlist
- 不改 rate-limit 配置字段、默认 prefix 或 HTTP 错误码

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/runtime/module.go`

## Current Baseline

- `practice/application/commands/service.go` 当前直接持有 `*redis.Client`
- `practice/application/commands/submission_service.go` 直接拼接 rate-limit key，并直接 `Incr` / `Expire`
- allowlist 当前保留：
  - `practice/application/commands/service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把 practice flag submit 限流状态面表达成 `practice` 自己的 submission rate-limit store：

1. 在 `practice/ports` 新增 `PracticeFlagSubmitRateLimitStore`
2. `practice/application/commands/Service` 只依赖该 store，保留“何时做 flag submit 限流、何时返回 `ErrSubmitTooFrequent`”的业务编排语义
3. `practice/infrastructure` 提供 Redis adapter，内部封装 prefix fallback、key 形状、`Incr` 和首次 `Expire` 细节
4. `practice/runtime/module.go` 统一构建该 store，并注入 practice command service

## Ownership Boundary

- `practice/application/commands/service.go`、`submission_service.go`
  - 负责：决定何时消耗一次 flag submit 尝试、何时因为超频拒绝提交
  - 不负责：知道 Redis key、`Incr`、首次 `Expire` 或 prefix fallback 细节
- `practice/infrastructure/submission_rate_limit_store.go`
  - 负责：实现 flag submit 计数和窗口初始化的 Redis 状态读写
  - 不负责：承接 flag 校验、manual review 或提交成功后的业务分支
- `practice/runtime/module.go`
  - 负责：装配 submission rate-limit store 并注入 practice command service
  - 不负责：把 Redis client 继续暴露回 practice application surface

## Change Surface

- Add: `.harness/reuse-decisions/practice-submission-rate-limit-store-phase5-slice10.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-submission-rate-limit-store-phase5-slice10-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/submission_rate_limit_store.go`
- Add: `code/backend/internal/module/practice/ports/submission_rate_limit_context_contract_test.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 提取 practice submission rate-limit store port 与 Redis adapter
  - 目标：practice command service 不再持有 `*redis.Client`，flag submit rate-limit 行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'SubmitFlag|Service' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：service/submission 是否已经不再知道 Redis concrete；首个提交设置窗口、超频拒绝和 context 透传是否保持不变

- [x] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除 `practice/application/commands/service.go` 的 Redis allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除 `practice/application/commands/service.go` 这条例外；文档准确描述 practice Redis concrete allowlist 的当前收口范围

## Risks

- 如果首次计数没有设置窗口 TTL，submit 限流会永久增长
- 如果 key 形状或 prefix fallback 变化，现有限流窗口行为会漂移
- 如果 runtime 继续把旧 Redis client 注回 practice command service，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'SubmitFlag|Service' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：practice command service 继续持有 flag submit 限流的业务编排语义，Redis 计数状态细节落回 infrastructure
- reuse point 明确：复用 contest slice7 的 submission rate-limit store 模式，但保持 practice key 形状和 use case 语义独立
- 这刀同时解决行为与结构：保留 practice submit 限流现有行为，同时删掉 practice command service 的 concrete Redis import
