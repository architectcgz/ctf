# Reuse Decision

## Change type
frontend refactor / feature router owner cleanup

## Existing code searched
- code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts

## Similar implementations found
- `useStudentDashboardPage.ts` 已经是 student dashboard route 的 page owner，当前天然持有 `useRoute()`、`useRouter()` 和 panel query sync。
- `useStudentDashboardData.ts` 当前主要负责仪表盘数据加载与展示派生，role redirect 混在这里属于 data owner 越权。

## Decision
refactor_existing

## Reason
`featureRouterImportAllowlist` 中，`features/student-dashboard/model/useStudentDashboardData.ts -> vue-router` 不是合理长期例外。这个文件的命名、返回面和调用面都说明它应该是 dashboard data owner，而不是 route-aware workflow owner。

最小正确改动是：

- 让 `useStudentDashboardData()` 不再接收 `Router`
- 把 teacher/admin 的 role redirect 变成显式 signal，由 `useStudentDashboardPage.ts` 统一处理
- 删除对应 allowlist 条目，并补 source guardrail，防止 router 再漂回 data 层

本轮不做：

- 不改 `useStudentDashboardPage.ts` 继续作为 route-aware page owner 的身份
- 不调整 student dashboard 的 panel registry、UI 结构和数据装配
- 不处理 `featureRouterImportAllowlist` 其它剩余条目

## Files to modify
- .harness/reuse-decisions/student-dashboard-data-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-student-dashboard-data-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-student-dashboard-data-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts

## After implementation
- `useStudentDashboardData.ts` 不再 import `vue-router`
- student dashboard 的 role redirect owner 明确回到 `useStudentDashboardPage.ts`
- `featureRouterImportAllowlist` 缩小一条
