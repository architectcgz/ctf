# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/platform-contests/model/useContestEditPage.ts
- code/frontend/src/views/platform/ContestEdit.vue
- code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue
- code/frontend/src/views/platform/__tests__/ContestEdit.test.ts
- code/frontend/src/router/routes/platformRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/platform-contests/model/contestManageRoutes.ts`
- `features/platform-contests/model/contestOperationsHubRoutes.ts`
- `features/platform-contests/model/useContestAnnouncementsPage.ts`
- `views/platform/ContestAnnouncements.vue`

## Decision
refactor_existing

## Reason
`useContestEditPage.ts` 当前混着两类 owner：

- 应继续保留在 page model 的：竞赛详情加载、AWD workbench 数据、保存 workflow、状态派生
- 不该继续依赖 `vue-router` 的：返回竞赛目录、打开公告页、进入 AWD 配置页、从赛前检查定位配置页

这条和前面几轮不同的一点是，它还包含“保存成功后返回目录”的 mutation 后跳转。这个跳转仍属于 page workflow，但不需要 page model 直接拿 `router.push()`。

因此这轮最小正确改动是：

- route param 改成 route props
- 薄导航改成显式 route target contract
- 保存成功后的跳转由独立 navigation transport 承接，而不是继续把 router 留在 page model

## Files to modify
- .harness/reuse-decisions/contest-edit-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-edit-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-edit-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/router/routes/platformRoutes.ts
- code/frontend/src/features/platform-contests/model/contestManageRoutes.ts
- code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts
- code/frontend/src/features/platform-contests/model/useContestEditPage.ts
- code/frontend/src/views/platform/ContestEdit.vue
- code/frontend/src/views/platform/__tests__/ContestEdit.test.ts
- code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue
- code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue
- code/frontend/src/features/awd-readiness/ui/AWDReadinessChecklist.vue
- code/frontend/src/components/navigation/AppRouteRedirect.vue

## After implementation
- `useContestEditPage.ts` 不再 import `vue-router`
- `ContestEdit` route 显式把 `contestId` 作为 props 下传
- 竞赛目录 / 公告页 / AWD 配置页入口都改成 route target contract
- 保存成功后的目录跳转由独立 navigation transport 承接
- `featureRouterImportAllowlist` 收掉 `useContestEditPage.ts`
