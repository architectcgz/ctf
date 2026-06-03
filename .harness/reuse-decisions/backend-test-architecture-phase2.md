# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 2

## Existing code searched
- `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
- `code/backend/internal/app/practice_flow_observability_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/testutil/systemapp/*.go`
- `code/backend/tests/README.md`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/testutil`
  - `code/backend/tests`

## Similar implementations found
- `internal/testutil/systemapp` 已经承接了 `practice_flow` 的 env、scenario runner 和断言 helper，可以作为场景断言迁移的直接依赖。
- `internal/app/practice_flow_lifecycle_integration_test.go` 与 `practice_flow_observability_integration_test.go` 现在仍持有完整测试断言逻辑，但它们的依赖已经都能从共享 testkit 获得。
- `code/backend/tests/README.md` 已经声明 `tests/system/http` 是后续黑盒 HTTP 系统测试的归属目录。

## Decision
refactor_existing

## Reason
本轮不删除 `internal/app` 下的现有测试文件，而是先把真实场景断言 owner 迁到 `code/backend/tests/system/http`：

- 在 `code/backend/tests/system/http/practiceflow/` 建立可导入的场景断言 package
- 让 `internal/app/practice_flow_*` 只保留 `Test... -> practiceflow.Verify...` 的兼容壳
- 继续复用 `internal/testutil/systemapp`，不重复造 env / request helper

这样可以继续缩小 `internal/app` 的测试 owner，同时避免因为直接删文件而触发高风险删除流程。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase2.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase2-plan.md`
- `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
- `code/backend/internal/app/practice_flow_observability_integration_test.go`
- `code/backend/tests/README.md`
- `code/backend/tests/system/http/practiceflow/*.go`

## After implementation
- `practice_flow` 的测试断言与场景 owner 不再放在 `internal/app`。
- `internal/app/practice_flow_*` 收成纯兼容入口。
- `tests/system/http` 会有第一批真实落点，后续同类 system tests 可以继续按这个模式迁移。
