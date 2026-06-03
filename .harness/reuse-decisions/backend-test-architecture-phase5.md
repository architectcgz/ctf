# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 5

## Existing code searched
- `code/backend/internal/app/full_router_admin_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouteraccess/access_matrix.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouteraccess/access_matrix.go` 已经证明 full router 子场景可以先迁断言 owner，再通过 callback 复用 `internal/app` 的 fixture。
- `full_router_admin_integration_test.go` 自身不持有 full router env owner，主要是场景 seed + HTTP 断言。

## Decision
refactor_existing

## Reason
`full_router_admin` 的最小安全切片仍然是先迁断言 owner：

- 在 `code/backend/tests/system/http/fullrouteradmin/` 建立可导入的场景断言 package
- `internal/app/full_router_admin_integration_test.go` 保留场景 seed、数据库准备和 glue code
- 先不抽 `newFullRouterTestEnv`、DB seed helper 或 publish request 相关 fixture

这样能继续缩小 `internal/app` 的系统测试 owner，同时避免把本轮 scope 扩成 full router fixture 基建迁移。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase5.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase5-plan.md`
- `code/backend/internal/app/full_router_admin_integration_test.go`
- `code/backend/tests/system/http/fullrouteradmin/*.go`
- `code/backend/tests/README.md`

## After implementation
- `full_router_admin` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `internal/app/full_router_admin_integration_test.go` 收成场景 seed + 兼容入口。
- full router fixture owner 仍暂留 `internal/app/full_router_integration_test.go`，后续继续按切片迁移。
