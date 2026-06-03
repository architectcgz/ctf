# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 3

## Existing code searched
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/full_router_*`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/testutil/systemapp/*.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/internal/testutil`

## Similar implementations found
- `internal/testutil/systemapp` 已承接 `practice_flow` 的 env、scenario runner、request helper 和断言 helper，`internal/app` 中对应桥接层大多已经无调用者。
- `full_router_*` 和 `router_test.go` 仍只依赖少数登录、header、config 和 image seed helper，不需要继续背着整套 practice flow bridge。

## Decision
refactor_existing

## Reason
phase 1 和 phase 2 完成后，`internal/app/practice_flow_integration_test.go` 中大量 wrapper 已经没有调用者：

- 场景 env / JSON decode / request helper 已由 `tests/system/http/practiceflow` 和 `internal/testutil/systemapp` 承接
- `internal/app` 里真正还在复用的只剩 `newPracticeFlowTestConfig`、登录 header helper 和 `createFlowImage`

因此本轮直接收掉死桥接层，让 `internal/app` 的残留 owner 只保留仍被 `full_router` 复用的最小集合。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase3.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase3-plan.md`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`

## After implementation
- `internal/app/practice_flow_integration_test.go` 只保留仍有调用者的兼容 helper。
- `practice_flow` 场景桥接不再残留在 `internal/app`。
- 下一步若继续迁移，可以直接面向 `full_router` 或其他剩余复用 helper 做独立切片。
