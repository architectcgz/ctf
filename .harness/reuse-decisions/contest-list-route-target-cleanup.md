# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/contest-detail/model/useContestListPage.ts
- code/frontend/src/views/contests/ContestList.vue
- code/frontend/src/views/contests/__tests__/ContestList.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/router/routes/studentRoutes.ts

## Similar implementations found
- `features/teacher-dashboard/model/teacherDashboardRoutes.ts`
- `features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts`
- `features/awd-review-workspace/model/awdReviewIndexRoutes.ts`

## Decision
refactor_existing

## Reason
`useContestListPage.ts` 当前只额外持有一条薄导航：进入竞赛详情。它已经不是 route/query owner，也不承担额外 workflow，仅仅因为一个 `router.push()` 留在 `featureRouterImportAllowlist` 里。

当前仓库对这类债的主模式已经很明确：

- page model 保留数据加载、筛选、分页和展示格式化 owner
- 单独新增 route target helper
- route view / widget 直接通过 `AppRouteLink` 消费目标路由

因此这轮最小正确改动是把 contest list 详情入口收口成显式 route target contract，而不是继续让 `useContestListPage.ts` 直接 import `vue-router`。

## Files to modify
- .harness/reuse-decisions/contest-list-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-list-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-list-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/contest-detail/model/useContestListPage.ts
- code/frontend/src/features/contest-detail/model/contestListRoutes.ts
- code/frontend/src/features/contest-detail/model/index.ts
- code/frontend/src/features/contest-detail/index.ts
- code/frontend/src/views/contests/ContestList.vue
- code/frontend/src/views/contests/__tests__/ContestList.test.ts

## After implementation
- `useContestListPage.ts` 不再 import `vue-router`
- contest list 详情入口改为显式 route target contract
- `ContestList.vue` 直接通过 `AppRouteLink` 消费竞赛详情路由
- `featureRouterImportAllowlist` 收掉 `useContestListPage.ts`
