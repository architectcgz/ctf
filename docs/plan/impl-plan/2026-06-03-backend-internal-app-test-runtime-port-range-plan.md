# 2026-06-03 backend internal app test runtime port range plan

## Objective

- 修复 `internal/app` 集成测试因为固定 host port range 与宿主机端口冲突而出现的 runtime 500。
- 保持 test runtime / sqlite 架构不变，只收口测试环境端口分配策略。

## Non-goals

- 本轮不修改生产 runtime 端口分配逻辑。
- 本轮不迁移测试到 Postgres，也不改业务 handler / service 契约。
- 本轮不处理与端口冲突无关的其余 `internal/app` 失败。

## Inputs

- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/test_schema_test.go`
- `code/backend/internal/app/composition/runtime_test_engine.go`

## Current problem

- `newPracticeFlowTestConfig(t)` 当前把 `cfg.Container.PortRangeStart/End` 固定写成 `30000-30100`。
- `runtime_test_engine.go` 在 test 环境会真实监听 `127.0.0.1:<hostPort>`。
- 当宿主机已有服务占用 `30000` 时，`TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge` / `TestFullRouter_AuthorizedSmokeMatrix` 会在实例创建链路上返回 500。

## Working design

### Target structure

- 在 `internal/app` 共享测试 helper 中增加一个动态端口块分配器：
  - 探测某段候选连续端口是否都能临时 `Listen`
  - 在同进程测试内记录已占用块，避免不同测试抢同一段
  - `t.Cleanup` 时归还块
- `newPracticeFlowTestConfig(t)` 改用该 helper 设置 `PortRangeStart/End`。

### Boundary choice

- 端口块分配 helper 放在 `internal/app` 共享测试基础设施里，因为 `practice_flow` / `full_router` / `router` 都经由同一测试配置入口。
- 不把这层逻辑塞进生产 `config` 或 runtime module，避免把测试环境特殊性泄漏进业务代码。

## Task slices

### Slice 1：red case

- Goal：先加一个失败测试，证明测试配置在 `30000` 被占用时仍会错误选择固定端口池。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewPracticeFlowTestConfigAvoidsOccupiedHostPortRange' -count=1`

### Slice 2：实现动态端口块分配

- Goal：新增共享 helper，并让 `newPracticeFlowTestConfig(t)` 改用它。
- Touched files：
  - `code/backend/internal/app/test_schema_test.go`
  - `code/backend/internal/app/practice_flow_integration_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewPracticeFlowTestConfigAvoidsOccupiedHostPortRange|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestFullRouter_AuthorizedSmokeMatrix' -count=1`

### Slice 3：最小充分回归

- Goal：确认共享 schema helper 和端口块分配一起工作，没有打坏基础 router 构建。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestInternalAppTestSchemaIncludesRuntimeTables|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestFullRouter_AuthorizedSmokeMatrix' -count=1`

## Expected change surface

- `internal/app` 测试基础设施
- `internal/app` 测试配置生成逻辑

## Data / API / compatibility impact

- 无生产行为变化。
- 风险主要在：
  - 动态端口块分配器本身仍可能选到被其他进程瞬时抢占的端口
  - 端口块过小，无法覆盖单测一次性创建的多个实例

## Rollback / recovery

- 纯测试代码修复，可直接回退测试端口块 helper 与配置入口变更。
