# 2026-06-03 backend test architecture phase 3 plan

## Objective

- 清掉 `internal/app` 中已经没有调用者的 `practice_flow` 桥接层。
- 把 `practice_flow_integration_test.go` 收成真正的最小兼容 helper owner。
- 为后续处理 `full_router` 复用 helper 留出更清楚的 touched surface。

## Non-goals

- 本轮不迁移 `full_router_*` 测试。
- 本轮不删除 `internal/app/practice_flow_*` 文件。
- 本轮不改 `tests/system/http/practiceflow` 的场景断言行为。
- 本轮不触碰生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/full_router_*`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/testutil/systemapp/*.go`
- `.harness/reuse-decisions/backend-test-architecture-phase3.md`

## Current problem

- phase 2 之后，`internal/app` 里还残留着大量 `practice_flow` bridge，但其中大部分已经没有调用者。
- 这些死桥接层会误导后续迁移判断，让 `internal/app` 看起来还在持有场景 owner。
- 如果不先清掉，下一轮继续拆测试时会反复扫到已经无效的中间层。

## Working design

### Target structure

- `practice_flow_integration_test.go` 只保留：
  - `newPracticeFlowTestConfig`
  - `createFlowImage`
  - `loginForSession`
  - `sessionHeaders`
  - `loginForToken`
  - `bearerHeaders`
- `practice_flow_scenario_test.go` 收成兼容说明文件，不再保留无调用者的 result wrapper。

### Boundary choice

- 不删文件，只收掉死代码，避免进入文件删除确认流程。
- 保留仍被 `full_router` 和 `router_test.go` 复用的 helper，避免把 scope 扩到新的测试体系。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 3 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase3`

### Slice 2：收掉无调用者的 practice flow bridge

- Goal：删去 `internal/app` 中不再被引用的 env / request / decode / scenario bridge，仅保留仍在使用的 helper。
- Touched files：
  - `code/backend/internal/app/practice_flow_integration_test.go`
  - `code/backend/internal/app/practice_flow_scenario_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter|TestFullRouter' -count=1`

### Slice 3：workflow gate

- Goal：确认本轮瘦身没有破坏 architecture / reuse / workflow 约束。
- Validation：
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- `internal/app` 中残留的 practice flow 兼容 helper

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - 误删仍被 `full_router` 或其他 `internal/app` 测试引用的 helper
  - 清空 `practice_flow_scenario_test.go` 后遗漏隐藏引用

## Review fit check

- Owner 更清晰：`internal/app` 只保留还有调用者的 helper，不再假装持有场景 owner。
- Reuse 更清晰：已迁移的场景逻辑继续留在 `internal/testutil/systemapp` 与 `tests/system/http/practiceflow`。
- 结构收敛：这刀是对前两阶段的收尾，不引入新抽象，也不扩大到 `full_router`。

## Rollback / recovery

- 纯测试 bridge 收缩，可直接回退到当前版本。
