# 2026-06-03 backend test architecture phase 1 plan

## Objective

- 为后端测试架构重做建立第一阶段落点：先把 `practice_flow` system tests 从 `package app` 私有 helper 中解耦。
- 在 `code/backend/internal/testutil/` 下建立可复用的 system testkit owner。
- 建立 `code/backend/tests/` 的目录入口，为后续把 system tests 迁出 `internal/app` 铺路。

## Non-goals

- 本轮不迁移 `full_router_*` system tests。
- 本轮不删除 `internal/app/practice_flow_*` 文件。
- 本轮不把 sqlite fixture 全量改成 `ctf-postgres`。
- 本轮不改生产业务逻辑或 runtime 行为。

## Inputs

- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/backend/05-key-flows.md`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/test_schema_test.go`
- `code/backend/internal/module/contest/testsupport/db.go`
- `.harness/reuse-decisions/backend-test-architecture-phase1.md`

## Current problem

- `practice_flow` system tests 当前仍然绑定在 `package app`，依赖本地未导出的 env、sqlite schema helper、HTTP helper 和 persistence row test types。
- 这种绑定意味着测试文件即使已经拆短，也无法直接迁到 `code/backend/tests/system/http` 一类的新目录。
- 如果直接物理迁文件，会先撞上 Go 包边界和 helper owner 问题，结果只能边迁边重写。

## Working design

### Target structure

- 在 `code/backend/internal/testutil/systemapp/` 新增：
  - sqlite schema / dynamic port helper
  - `PracticeFlowEnv`
  - `RunPublishedPracticeFlowScenario`
  - practice flow response / scenario result types
  - request / decode / login helpers
- `internal/app/practice_flow_*` 改为 import 新 testkit，不再直接持有 env/helper owner。
- 新增 `code/backend/tests/README.md`，明确未来目录职责：
  - `tests/system/http`
  - `tests/runtime`
  - `tests/testkit`

### Boundary choice

- 先抽 helper，不先迁目录。这样可以保证当前测试入口和 CI 习惯不变，同时把下一刀的迁移成本压低。
- testkit 先只覆盖 `practice_flow`，不在本轮同时泛化到 `full_router`，避免抽象过度。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase1`

### Slice 2：抽取 practice flow system testkit

- Goal：把 `practice_flow` 现有 env / sqlite helper / scenario runner 迁到 `internal/testutil/systemapp`，并让 `internal/app/practice_flow_*` 改用新包。
- Touched files：
  - `code/backend/internal/app/practice_flow_integration_test.go`
  - `code/backend/internal/app/practice_flow_scenario_test.go`
  - `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
  - `code/backend/internal/app/practice_flow_observability_integration_test.go`
  - `code/backend/internal/testutil/systemapp/*.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_' -count=1`

### Slice 3：建立 tests 入口说明

- Goal：在 `code/backend/tests/` 建立新的测试目录入口说明，明确分层和后续迁移方向。
- Touched files：
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- `practice_flow` 测试 helper owner
- `internal/testutil` 测试基建
- backend tests 新目录入口说明

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变更。
- 风险主要在：
  - helper 抽取时漏掉当前 scenario 依赖
  - 导出类型后断言或 JSON 结构不再完全等价
  - 把还没有稳定复用面的逻辑过早泛化成“大而全 testkit”

## Review fit check

- Owner 清晰：`practice_flow` 的 system env / scenario owner 从 `internal/app` 转到 `internal/testutil/systemapp`。
- Reuse 清晰：只抽当前已经被多文件共享的 helper 和场景，不引入提前泛化的全局测试框架。
- 结构收敛：先解决“系统测试被 package app 私有 helper 锁死”的根因，为后续物理迁目录扫清边界。

## Rollback / recovery

- 纯测试代码与目录说明调整，可直接回退到 `internal/app` 私有 helper 版本。
