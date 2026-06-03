# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 6

## Existing code searched
- `code/backend/internal/app/full_router_teacher_authoring_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouteraccess/access_matrix.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouteraccess/access_matrix.go` 已经证明 full router 子场景可以先迁断言 owner，再通过 callback 复用 `internal/app` 的 fixture。
- `tests/system/http/fullrouteradmin/admin_flow.go` 已经证明可以把 HTTP 场景 owner 迁出，同时把 DB seed 留在 `internal/app`。

## Decision
refactor_existing

## Reason
`full_router_teacher_authoring` 这组测试的最小安全切片仍然是先迁断言 owner：

- 在 `code/backend/tests/system/http/fullrouterteacherauthoring/` 建立可导入的场景断言 package
- `internal/app/full_router_teacher_authoring_integration_test.go` 保留数据库校验、文件系统 seed 和少量 glue code
- 先不抽 `newFullRouterTestEnv`、authoring challenge seed builder 或 full router fixture owner

这样能继续缩小 `internal/app` 的系统测试 owner，同时避免把这一刀扩成整套 fixture / testkit 迁移。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase6.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase6-plan.md`
- `code/backend/internal/app/full_router_teacher_authoring_integration_test.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/*.go`
- `code/backend/tests/README.md`

## After implementation
- `full_router_teacher_authoring` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `internal/app/full_router_teacher_authoring_integration_test.go` 收成 seed、数据库断言和兼容入口。
- full router fixture owner 仍暂留 `internal/app/full_router_integration_test.go`，后续继续按切片迁移。
