# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 1

## Existing code searched
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
- `code/backend/internal/app/practice_flow_observability_integration_test.go`
- `code/backend/internal/app/test_schema_test.go`
- `code/backend/internal/module/contest/testsupport/db.go`
- `code/backend/internal/testutil`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`

## Similar implementations found
- `internal/app/test_schema_test.go` 已经提供了 `sqlite` schema template 与动态端口块保留逻辑，但它们目前被锁在 `package app` 内。
- `internal/module/contest/testsupport/db.go` 说明模块级测试已经在用独立 testsupport owner，只是还没有统一的 system test owner。
- 最近的 `practice_flow` / `full_router` 文件拆分已经把测试职责按文件收口，但系统场景仍停留在 `internal/app`。

## Decision
refactor_existing

## Reason
本轮不直接删除或整体迁走 `internal/app` 下的 system tests，而是先做可复用 testkit 抽取：

- 在 `code/backend/internal/testutil` 下建立新的 system test helper owner
- 把 `practice_flow` 场景运行所需的 sqlite schema、动态端口、HTTP env、scenario runner 和断言类型迁到新包
- 让现有 `internal/app/practice_flow_*` 测试先改用新包
- 同时建立 `code/backend/tests/` 的新入口说明，为后续物理迁移到 `tests/system/http` 铺路

这样能先解除 `package app` 私有 helper 对 system tests 的绑定，让下一刀目录迁移变成机械搬运，而不是边搬边重写。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase1.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase1-plan.md`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
- `code/backend/internal/app/practice_flow_observability_integration_test.go`
- `code/backend/internal/testutil/systemapp/*.go`
- `code/backend/tests/README.md`

## After implementation
- `practice_flow` 场景测试不再依赖 `package app` 私有 env/helper owner。
- `internal/testutil/systemapp` 成为 system testkit 的首个落点。
- `code/backend/tests/README.md` 说明新的系统测试目录职责与迁移方向。
- 本轮不删除 `internal/app` 里的 system test 文件，也不迁 `full_router`。
