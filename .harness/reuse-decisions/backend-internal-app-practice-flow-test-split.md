# Reuse Decision

## Change type
backend test refactor / oversized integration test split

## Existing code searched
- `code/backend/internal/module`
- `code/backend/internal/app/composition`
- `code/backend/internal/model`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/router_test.go`

## Similar implementations found
- `full_router_state_matrix_integration_test.go` 和 `full_router_integration_test.go` 已经收口成“原文件保留 shared env/helper owner，新文件承载更窄的顶层测试”模式。
- `practice_flow_integration_test.go` 和前两刀不同：这里只有两个顶层测试，但其中一个已经承担了多个流程阶段，因此需要继续复用“shared env owner + 顶层测试按职责分解”的方向，只是这次要先把巨型测试切成多个场景测试。

## Decision
refactor_existing

## Reason
本轮不引入新的测试框架或新的 env 体系，而是在现有 `practice_flow` 测试上做最小结构收敛：

- 保留 `newPracticeFlowTestEnv`、登录 helper、JSON helper 和断言 helper
- 把大测试里的练习链路沉到共享 scenario runner
- 把顶层断言拆成 lifecycle/access、submission/progress、observability 三组测试
- 保留未发布题目的负向用例，但移出 giant test owner

相比只移动一个小测试或完全重写 practice flow fixture，这条路径能真正解决 1000+ 行文件问题，同时保持语义不变。

## Files to modify
- `.harness/reuse-decisions/backend-internal-app-practice-flow-test-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-practice-flow-test-split-plan.md`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/practice_flow_scenario_test.go`
- `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
- `code/backend/internal/app/practice_flow_observability_integration_test.go`

## After implementation
- `practice_flow_integration_test.go` 只保留 shared env、fixture 和 helper。
- `practice_flow_scenario_test.go` 承载共享的已发布练习链路 scenario runner 和结果结构。
- 练习流程的用户链路与观测链路拆成多个更短的顶层测试。
- 本轮不继续拆 `router_test.go`。
