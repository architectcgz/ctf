# Reuse Decision

## Change type
backend test refactor / oversized integration test split

## Existing code searched
- `code/backend/internal/module`
- `code/backend/internal/app/composition`
- `code/backend/internal/model`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## Similar implementations found
- `full_router_state_matrix_integration_test.go` 已经在上一刀证明：保留同包 helper owner，只移动顶层 `Test...` 到按领域分组的新文件，是当前 `internal/app` 集成测试最稳妥的拆法。
- `full_router_integration_test.go` 也有同样的天然分层：前半段是顶层测试，后半段从 `newFullRouterTestEnv` 开始都是共享 env 和 route helper。

## Decision
refactor_existing

## Reason
这次继续复用上一刀已经验证过的拆分模式：

- 继续使用 `package app`
- 继续复用原文件中的 `fullRouterTestEnv`、`newFullRouterTestEnv`、route helper 和 fixture seed
- 不改变测试语义、HTTP 断言、seed 数据和路由 owner
- 只把顶层测试按 access / admin / teacher-authoring 三组拆到独立文件

相比重新抽 helper 或重排整个 `internal/app` 测试体系，这条路径改动最小，也能直接降低后续继续扩展 `full_router_integration_test.go` 的维护成本。

## Files to modify
- `.harness/reuse-decisions/backend-internal-app-full-router-integration-test-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-full-router-integration-test-split-plan.md`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_access_integration_test.go`
- `code/backend/internal/app/full_router_admin_integration_test.go`
- `code/backend/internal/app/full_router_teacher_authoring_integration_test.go`

## After implementation
- `full_router_integration_test.go` 只保留 shared env / helper 和 `TestRouterBuildUsesCompositionModules`。
- access / admin / teacher-authoring 三类顶层集成测试分别拥有独立文件。
- 本轮不继续拆 `practice_flow_integration_test.go`、`router_test.go`。
