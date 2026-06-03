# Reuse Decision

## Change type
backend test refactor / integration test schema consolidation

## Existing code searched

- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/module/runtime/entity/*.go`
- `code/backend/internal/module/runtime/infrastructure/*.go`

## Similar implementations found

- `full_router_integration_test.go` 已经有 sqlite schema template 复用逻辑，但表清单只在该文件内部维护。
- `practice_flow_integration_test.go` 和 `router_test.go` 也各自维护独立 sqlite 初始化逻辑，重复面已经出现。
- runtime 侧默认节点建表逻辑在 `composition/runtime_module.go`，说明 runtime node 缺表本身已经是已知测试适配点。

## Decision
refactor_existing

## Reason

当前最小正确改动不是继续在 `full_router` 或 `practice_flow` 里各补一张表，而是把 `internal/app` 这组测试共享的 sqlite schema owner 收口到一处：

- 当前 `internal/app` 多套测试都在重复维护建表清单，runtime / practice 新增表时很容易只补一处。
- `TestFullRouter_AuthorizedSmokeMatrix` 等失败已经证明，分散维护会直接把 runtime 路径打成 500，而不是单纯测试数据问题。
- 抽成共享 helper 后，runtime 相关表只需要维护一份；后续如果要补 Postgres-backed 测试路径，也有统一入口可扩展。

## Files to modify

- `.harness/reuse-decisions/backend-internal-app-test-schema-consolidation.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-test-schema-consolidation-plan.md`
- `code/backend/internal/app/test_schema_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/router_test.go`

## After implementation

- `internal/app` 的 sqlite 测试 schema 表清单收口到共享 helper。
- `full_router`、`practice_flow`、`router` 测试不再各自维护独立表集。
- runtime 相关缺表至少覆盖 `network_allocations`、`runtime_nodes` 及当前 `internal/app` 测试已触达的 runtime 基础表。
