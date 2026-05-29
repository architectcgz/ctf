# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts
- code/frontend/src/views/platform/ContestAnnouncements.vue
- code/frontend/src/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue
- code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/platform-contests/model/contestManageRoutes.ts`
- `features/platform-contests/model/contestOperationsHubRoutes.ts`
- `features/contest-detail/model/contestListRoutes.ts`

## Decision
refactor_existing

## Reason
`useContestAnnouncementsPage.ts` 当前的主 owner 是单场公告页的竞赛详情加载、公告列表 workflow、发布和删除操作。额外保留的 `goBackToStudio()` 只是单条薄导航：返回竞赛编辑工作台。

这类行为不值得继续把 `useRouter()` 留在 page model 里。当前仓库对这类债的主模式已经明确：

- page model 保留数据、async workflow 和错误态 owner
- 单独新增或复用 route target helper
- view / feature UI 直接通过 `AppRouteLink` 消费目标路由

因此这轮最小正确改动是把 contest announcements 的返回入口收口成显式 route target contract，而不是继续让 `useContestAnnouncementsPage.ts` 直接 import `vue-router`。

## Files to modify
- .harness/reuse-decisions/contest-announcements-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-announcements-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-announcements-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts
- code/frontend/src/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue
- code/frontend/src/views/platform/ContestAnnouncements.vue
- code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts

## After implementation
- `useContestAnnouncementsPage.ts` 不再 import `vue-router`
- 公告页返回竞赛工作台入口改为显式 route target contract
- topbar 和错误空态直接通过 `AppRouteLink` 消费返回路由
- `featureRouterImportAllowlist` 收掉 `useContestAnnouncementsPage.ts`
