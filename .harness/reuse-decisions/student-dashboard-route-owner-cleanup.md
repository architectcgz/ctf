# Reuse Decision

## Change type
frontend refactor / student dashboard route owner cleanup

## Existing code searched
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts
- code/frontend/src/views/dashboard/DashboardView.vue
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts
- code/frontend/src/features/challenge-list/model/challengeListRoutes.ts
- code/frontend/src/composables/useRouteQueryTabs.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/skill-profile/model/skillProfileRoutes.ts`
- `features/challenge-list/model/challengeListRoutes.ts`
- `composables/routeQueryTransport.ts`
- `views/dashboard/__tests__/DashboardView.test.ts`

## Decision
refactor_existing

## Reason
`useStudentDashboardPage.ts` 现在同时混着三类 route owner：

- `panel` query tab sync
- 题库 / 分类 / 难度 / 画像 / 题目详情 5 条薄导航
- teacher/admin 角色重定向

这条如果继续把 `vue-router` 留在 page model 里，`featureRouterImportAllowlist` 不会下降；但如果为了消掉 allowlist 把这些按钮都硬改成 route link，又会把当前 dashboard panel 的 action contract 放大成一次更重的 UI 迁移。更合适的收口方式是：

- `panel` query tab 继续留在 student dashboard page owner，由 `useRouteQueryTabs()` 内部持有 route transport
- 新增本地 route target builder，统一描述 5 条薄导航与角色重定向目标
- 把 `push / replace` 下沉成共享 navigation transport，让 page model 只保留 dashboard 自己的 workflow owner

这样既能真实减少 allowlist，又不会引入新的 route wrapper 中间态，也不需要把现有 dashboard panel 的 button contract 全部翻成 link surface。

## Files to modify
- .harness/reuse-decisions/student-dashboard-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-student-dashboard-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-student-dashboard-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/features/student-dashboard/model/studentDashboardRoutes.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts

## After implementation
- `useStudentDashboardPage.ts` 不再 import `vue-router`
- student dashboard 的 query tab 继续由 page owner 负责，但 router transport 不再散落在 feature page model 里
- dashboard 内薄导航与角色 redirect 改为显式 route target + navigation transport
- `featureRouterImportAllowlist` 再收掉 `features/student-dashboard/model/useStudentDashboardPage.ts -> vue-router`
