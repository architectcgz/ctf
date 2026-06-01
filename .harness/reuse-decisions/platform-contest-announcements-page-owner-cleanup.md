# Reuse Decision

## Change type
frontend refactor / platform contest announcements route page owner cleanup

## Existing code searched
- `code/frontend/src/pages/platform/contests/ContestAnnouncementsRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestAnnouncementsTopbarPanel.vue`
- `code/frontend/src/features/platform/contests/ui/ContestAnnouncementsWorkspacePanel.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsHubPage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `PlatformContestManagePage.vue` 已经承接 contest manage 的 page model、page shell 和 overlay owner，route page 退回薄壳。
- `PlatformContestOperationsHubPage.vue` 也已经承接 operations hub 的 page model 与 panel 组合，route page 只保留 feature public API 薄壳。
- `ContestAnnouncementsRoutePage.vue` 当前同样直接组合 `useContestAnnouncementsPage()`、topbar panel 与 workspace panel，只是额外透传一个 `contestId` route prop，因此适合沿同一模式继续收口。

## Decision
refactor_existing

## Reason
当前最小正确改动不是去碰公告发布/删除 workflow，而是新增 `PlatformContestAnnouncementsPage.vue`，把：

- `useContestAnnouncementsPage(toRef(props, 'contestId'))`
- `ContestAnnouncementsTopbarPanel`
- `ContestAnnouncementsWorkspacePanel`
- 本地 loading / error shell

一起收回 `features/platform/contests/ui` 内部。`ContestAnnouncementsRoutePage.vue` 继续只保留 route prop 输入并渲染 feature page，从而把 route-level owner 压到最小，同时不改变 `contestId` prop contract。

## Files to modify
- `.harness/reuse-decisions/platform-contest-announcements-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-platform-contest-announcements-page-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-platform-contest-announcements-page-owner-cleanup-review.md`
- `code/frontend/src/features/platform/contests/ui/PlatformContestAnnouncementsPage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/pages/platform/contests/ContestAnnouncementsRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 单场公告管理 route page 会退回 “route prop -> feature page” 的薄壳。
- `platform/contests` 内会新增明确的 announcements page-level owner，后续如果继续清 `ContestEditRoutePage.vue` 等路由页，可以继续复用同一模式。
- `useContestAnnouncementsPage.ts`、公告发布/删除 workflow 和现有 route target contract 保持不变。
