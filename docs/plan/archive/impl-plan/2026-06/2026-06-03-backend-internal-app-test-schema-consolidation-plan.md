# 2026-06-03 backend internal app test schema consolidation plan

## Objective

- 把 `code/backend/internal/app` 下分散的 sqlite 测试建表逻辑收口到共享 helper。
- 修复 `full_router` / `practice_flow` 这组 runtime 相关测试因为缺少 runtime schema 而出现的既有 500。
- 保持现有测试行为、fixture 语义和 sqlite 隔离方式不变。

## Non-goals

- 本轮不把 `internal/app` 全量集成测试迁到 Postgres。
- 本轮不改业务 handler、service 或 runtime 仓储实现。
- 本轮不处理所有 `internal/app` 失败，只聚焦共享 schema owner 和由缺表直接导致的失败。

## Inputs

- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/module/runtime/entity/network_allocation.go`
- `code/backend/internal/module/runtime/entity/runtime_node.go`

## Current problem

- `full_router`、`practice_flow`、`router` 测试当前分别维护 sqlite 初始化逻辑。
- `full_router` 和 `practice_flow` 的 schema 表集不一致，runtime / practice 新增持久化表后容易漏补。
- 当前已复现的失败包括 `network_allocations`、`runtime_nodes`、`instances` 缺表触发的 500；这些失败先暴露的是测试 schema owner 分散，不是业务主逻辑。

## Working design

### Target structure

- 在 `internal/app` 新增共享测试 helper，例如：
  - 统一声明 `internal/app` 测试需要的 schema models
  - 统一提供 sqlite 打开 / migrate / template clone 能力
- `full_router_integration_test.go` 继续保留 template clone 以控制启动成本，但模板来源改为共享 schema owner。
- `practice_flow_integration_test.go` 和 `router_test.go` 改用同一套 helper，而不是手写 `AutoMigrate(...)`。

### Boundary choice

- helper 只负责测试数据库 schema 和 sqlite 打开方式，不接管业务 seed。
- schema owner 放在 `internal/app`，因为当前复用面都在 app 级集成测试，而不是某个单独 module。
- runtime 缺表要直接并入共享 schema，而不是在单个失败测试里临时补 `AutoMigrate`。

## Task slices

### Slice 1：intake 与 red case

- Goal：补 reuse decision / plan，并用现有失败测试确认缺表问题。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-internal-app-test-schema-consolidation`
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`

### Slice 2：收口共享 schema helper

- Goal：新增 `internal/app` 共享测试 schema/helper，并让 `full_router`、`practice_flow`、`router` 改用它。
- Touched files：
  - `code/backend/internal/app/test_schema_test.go`
  - `code/backend/internal/app/full_router_integration_test.go`
  - `code/backend/internal/app/practice_flow_integration_test.go`
  - `code/backend/internal/app/router_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestNewRouterRegistersStudentChallengeRoutes' -count=1`

### Slice 3：最小充分回归

- Goal：确认共享 schema 收口后，router / full router / practice flow 关键面不回退。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestFullRouter_AuthorizedSmokeMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
  - `bash scripts/check-consistency.sh`

## Expected change surface

- `internal/app` 测试基础设施
- sqlite schema model owner
- 受影响测试文件的建库入口

## Data / API / compatibility impact

- 无生产数据、API 或配置变更。
- 风险主要在：
  - 抽 helper 时漏掉现有某组测试依赖的表
  - template clone 和直接 migrate 的行为不一致
  - `router_test.go` 这类仅构建 router 的测试被不必要地耦合到更重的 schema

## Review fit check

- Owner 清晰：`internal/app` 的测试 schema 由单一 helper 维护。
- Reuse 清晰：`full_router` 已有 template 复用逻辑，本轮只把 schema owner 上提。
- 结构收敛：既然本轮已经因为缺表触碰测试基建，就顺手完成同一 surface 的 owner 收口，不再接受“某个测试再补一张表”的继续分散。

## Rollback / recovery

- 纯测试代码重构，可直接回退共享 helper 与引用点。
