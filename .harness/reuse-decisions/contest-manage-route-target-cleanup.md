# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/platform-contests/model/useContestManagePage.ts
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue
- code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue
- code/frontend/src/views/platform/__tests__/ContestManage.test.ts
- code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/contest-detail/model/contestListRoutes.ts`
- `features/teacher-dashboard/model/teacherDashboardRoutes.ts`
- `features/awd-review-workspace/model/awdReviewIndexRoutes.ts`

## Decision
refactor_existing

## Reason
`useContestManagePage.ts` 当前除竞赛目录数据、抽屉状态和创建流程 owner 外，还额外保留了三条薄导航：

- 进入竞赛编辑页
- 进入竞赛运维台
- 进入公告完整管理页

这三条都不再是 route/query owner，也不承载额外 workflow，只是因为 `router.push()` 继续留在 `featureRouterImportAllowlist` 里。当前仓库对这类债的主模式已经明确：

- page model 保留数据、筛选、dialog 和 workflow owner
- 单独新增 route target helper
- view / feature UI 直接通过 `AppRouteLink` 消费目标路由

因此这轮最小正确改动是把 contest manage 的三条目录导航收口成显式 route target contract，而不是让 `ContestManage.vue` 或 `useContestManagePage.ts` 继续直接 import `vue-router`。

## Files to modify
- .harness/reuse-decisions/contest-manage-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-manage-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-manage-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-contests/model/useContestManagePage.ts
- code/frontend/src/features/platform-contests/model/contestManageRoutes.ts
- code/frontend/src/features/platform-contests/model/index.ts
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue
- code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/views/platform/__tests__/ContestManage.test.ts
- code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts

## After implementation
- `useContestManagePage.ts` 不再 import `vue-router`
- 竞赛管理页的编辑 / 运维 / 公告完整页入口改为显式 route target contract
- `PlatformContestTable.vue` 与 `ContestAnnouncementManageDrawer.vue` 直接通过 `AppRouteLink` 消费竞赛管理路由
- `featureRouterImportAllowlist` 收掉 `useContestManagePage.ts`
