# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 13

## Existing code searched
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `full_router_state_matrix_integration_test.go` 当前同时承载 module builder wiring 测试、report state 场景 glue 和一批共享 helper。
- `full_router_integration_test.go` 已基本只剩 module builder smoke + shared fixture，不适合继续按 HTTP 场景迁移。

## Decision
refactor_existing

## Reason
phase 13 改成文件职责拆分，而不是继续扩新的 HTTP 场景 package：

- 在 `internal/app` 下新增独立 test file 承接两个 module builder/wiring 测试
- `full_router_state_matrix_integration_test.go` 收敛成 report state glue 与共享 helper owner
- 本轮不抽 `waitForReportStatus`、multipart/websocket helper 或其他共享 seed

这样可以继续降低单文件混合职责，同时避免为了几十行 wiring 测试额外造不合适的 `tests/system/http` package。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase13.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase13-plan.md`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_module_wiring_test.go`

## After implementation
- `TestTeacherRoutesAreServedByTeachingQuery` 与 `TestStudentPracticeReadRoutesAreServedByPracticeModule` 不再和 report state / helper 混在同一文件。
- `full_router_state_matrix_integration_test.go` 继续保留 report state 场景 glue 与共享 helper owner。
