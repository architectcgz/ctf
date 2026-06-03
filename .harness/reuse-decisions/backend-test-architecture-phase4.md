# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 4

## Existing code searched
- `code/backend/internal/app/full_router_access_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/tests/system/http/practiceflow/practice_flow.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/practiceflow/practice_flow.go` 已经证明“把场景断言 owner 迁出 `internal/app`，但先保留包内 env/helper”这条路径可行。
- `full_router_access_integration_test.go` 自己不持有 env owner，主要依赖 `full_router_integration_test.go` 提供的 fixture / route helper。

## Decision
refactor_existing

## Reason
对 `full_router_access` 来说，最小安全切片不是先抽整套 env，而是先迁断言 owner：

- 在 `code/backend/tests/system/http/fullrouteraccess/` 建立可导入的断言 package
- `internal/app/full_router_access_integration_test.go` 改成只负责 glue code，把现有 env / request / route helper 作为回调传入
- 先避免复制 `newFullRouterTestEnv`、seed 和 route materialize 逻辑，等 access 场景迁完后再决定是否抽 full router fixture owner

这样能继续缩小 `internal/app` 的场景测试 owner，同时避免本轮 scope 膨胀成 “重做整套 full router 测试基建”。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase4.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase4-plan.md`
- `code/backend/internal/app/full_router_access_integration_test.go`
- `code/backend/tests/system/http/fullrouteraccess/*.go`

## After implementation
- `full_router_access` 的核心断言 owner 不再放在 `internal/app`。
- `internal/app/full_router_access_integration_test.go` 收成兼容入口。
- full router fixture owner 仍暂留 `internal/app/full_router_integration_test.go`，作为后续单独切片处理。
