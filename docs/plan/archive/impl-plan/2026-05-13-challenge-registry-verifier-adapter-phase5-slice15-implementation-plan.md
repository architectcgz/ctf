# Challenge Registry Verifier Adapter Phase 5 Slice 15 Implementation Plan

## Objective

继续 phase 5 收窄 `challenge` application concrete allowlist，把 registry manifest 校验的 HTTP adapter 下沉到模块 infrastructure：

- 去掉 `challenge/application/commands/registry_client.go -> net/http`
- 保持 image build service 对 `challengeports.RegistryVerifier` 的使用方式和 manifest 校验行为不变

## Non-goals

- 不处理 `auth/application/commands/cas_service.go -> net/http`
- 不改 `challenge/application/commands/image_build_service.go` 的业务编排、仓储事务或镜像构建状态机
- 不新增 registry 领域模型、事件或新的跨模块 contract

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/challenge/application/commands/registry_client.go`
- `code/backend/internal/module/challenge/application/commands/registry_client_test.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Current Baseline

- `challenge/application/commands/registry_client.go` 当前直接持有 `*http.Client`
- 它本质上是 `challengeports.RegistryVerifier` 的 HTTP 实现，但物理位置还在 application 包下
- `challenge/runtime/module.go` 直接从 `challenge/application/commands` 构建该 verifier
- allowlist 当前保留：
  - `challenge/application/commands/registry_client.go -> net/http`

## Chosen Direction

把 registry manifest 校验收口为 `challenge` 自己的 infrastructure adapter：

1. 保留 `challengeports.RegistryVerifier` 作为 application 依赖的窄能力口
2. 把现有 `RegistryClient` 和测试迁到 `challenge/infrastructure`
3. `challenge/runtime/module.go` 改为从 infrastructure 构建 verifier，再注入 image build service
4. 删除对应 allowlist，并同步 phase5 当前事实

## Ownership Boundary

- `challenge/application/commands/image_build_service.go`
  - 负责：决定何时校验 external image manifest、如何处理 digest / size / build 状态
  - 不负责：知道 HTTP request、accept header、认证头或 registry URL 拼装细节
- `challenge/infrastructure/registry_client.go`
  - 负责：实现 registry manifest 的 HTTP 校验、认证头组装和 digest 提取
  - 不负责：决定 image build service 的业务分支、重试策略或镜像状态写入
- `challenge/runtime/module.go`
  - 负责：装配 registry verifier adapter 并传给 image build service
  - 不负责：把 `net/http` 依赖继续留在 application surface

## Change Surface

- Add: `.harness/reuse-decisions/challenge-registry-verifier-adapter-phase5-slice15.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-registry-verifier-adapter-phase5-slice15-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- Add: `code/backend/internal/module/challenge/infrastructure/registry_client_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Delete: `code/backend/internal/module/challenge/application/commands/registry_client.go`
- Delete: `code/backend/internal/module/challenge/application/commands/registry_client_test.go`

## Task Slices

- [ ] Slice 1: 迁移 registry verifier adapter
  - 目标：`challenge` application surface 不再直接持有 `net/http` concrete，运行时装配和测试行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'RegistryClient' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/challenge/application/commands -run 'ImageBuildService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/challenge/runtime -run 'BuildWires.*ImageBuildService' -count=1 -timeout 5m`
  - Review focus：HTTP 细节是否完全落回 infrastructure；runtime 是否仍能给 image build service 注入正确 verifier

- [ ] Slice 2: 删除 allowlist 并同步文档
  - 目标：删掉 `challenge/application/commands/registry_client.go -> net/http`，phase5 当前事实同步更新
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除本次实际收口的 allowlist；文档正确反映 challenge 仍剩余的 concrete 依赖面

## Risks

- 如果迁移过程中改动了 registry URL 或认证头拼装，external image 校验可能误判失败
- 如果 runtime 仍保留旧 application 构造路径，allowlist 不会真正收口
- 如果测试路径没有同步迁移，可能会漏掉 verifier 的 HTTP contract 回归

## Verification Plan

1. `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'RegistryClient' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/challenge/application/commands -run 'ImageBuildService' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/challenge/runtime -run 'BuildWires.*ImageBuildService' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`
8. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：image build service 继续只依赖 `challengeports.RegistryVerifier`，HTTP adapter 明确落回 infrastructure
- reuse point 明确：复用当前 `challenge` 模块已存在的 `ports + infrastructure + runtime wiring` 结构，不引入新的宽抽象
- 这刀同时解决行为与结构：保留 registry manifest 校验行为，同时删除 application surface 的 `net/http` concrete 例外
