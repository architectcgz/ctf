# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 8

## Existing code searched
- `code/backend/internal/app/full_router_contest_state_matrix_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterawdstate/awd_state.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/authoring_flow.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouterawdstate/awd_state.go` 已经证明可以把多场景 HTTP 状态矩阵 owner 迁出，同时把 DB/contest seed 留在 `internal/app`。
- `full_router_contest_state_matrix_test.go` 里的 contest 场景大多是 HTTP 状态流，适合继续沿用 driver + callback 的模式。

## Decision
refactor_existing

## Reason
`full_router_contest_state_matrix` 的最小安全切片仍然是先迁断言 owner：

- 在 `code/backend/tests/system/http/fullrouterconteststate/` 建立可导入的场景断言 package
- `internal/app/full_router_contest_state_matrix_test.go` 保留 contest/challenge/user seed、少量数据库操作和 glue code
- 先不抽 `newFullRouterTestEnv`、contest seed builder、scoreboard seed helper 或 report wait helper owner

这样能继续缩小 `internal/app` 的系统测试 owner，同时避免本轮 scope 扩成 contest 测试基建重做。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase8.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase8-plan.md`
- `code/backend/internal/app/full_router_contest_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouterconteststate/*.go`
- `code/backend/tests/README.md`

## After implementation
- `full_router_contest_state_matrix` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `internal/app/full_router_contest_state_matrix_test.go` 收成 contest/challenge/user seed、少量数据库操作和兼容入口。
- full router fixture owner 仍暂留 `internal/app/full_router_integration_test.go`，后续继续按切片迁移。
