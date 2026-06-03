# Reuse Decision

## Change type
backend test refactor / oversized integration test split

## Existing code searched
- `code/backend/internal/module`
- `code/backend/internal/app/composition`
- `code/backend/internal/model`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-test-schema-consolidation-plan.md`

## Similar implementations found
- `internal/app` 已经存在按职责拆开的 `*_test.go` 文件，测试环境和 helper 通过同包共享，而不是把所有集成测试堆到一个文件里。
- 当前 `full_router_state_matrix_integration_test.go` 已经天然按 contest / teacher / admin / awd 几个领域分段，只是还没有落成独立文件。

## Decision
refactor_existing

## Reason
这次不是新增新的测试基建，而是把已经存在的顶层测试函数按领域移动到独立 `_test.go` 文件：

- 继续复用同一个 `package app`
- 继续复用现有 `newFullRouterTestEnv` 与底部 helper
- 不修改测试数据模型、路由断言和 fixture 语义
- 只把 owner 混杂的超大文件拆成更可维护的几个 review 单位

相比重新设计 helper 或重新归类所有 `internal/app` 测试，这条路径改动最小，也能直接降低后续继续扩展同一文件的成本。

## Files to modify
- `.harness/reuse-decisions/backend-internal-app-full-router-state-matrix-test-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-full-router-state-matrix-test-split-plan.md`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_contest_state_matrix_test.go`
- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/internal/app/full_router_awd_state_matrix_test.go`

## After implementation
- `full_router_state_matrix_integration_test.go` 只保留轻量 wiring 测试、`ReportPreview` 状态矩阵和共享 helper。
- contest / teacher / admin / awd 四类状态矩阵测试分别拥有独立文件。
- 本轮不继续拆 `full_router_integration_test.go`、`practice_flow_integration_test.go`、`router_test.go`。
