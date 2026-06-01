# Reuse Decision

## Change type
frontend refactor / platform contest edit route page owner cleanup

## Existing code searched
- `code/frontend/src/pages/platform/contests/ContestEditRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestEditPage.ts`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestAnnouncementsPage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsPage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `PlatformContestManagePage.vue` 已承接 contest manage 的 page model、page shell 与 overlay owner，route page 退回薄壳。
- `PlatformContestAnnouncementsPage.vue` 与 `PlatformContestOperationsPage.vue` 已分别承接公告页、单场运维页的 page model 与 page shell，route page 只保留 feature public API 薄壳。
- `ContestEditRoutePage.vue` 当前也只接收一个 `contestId` route prop，但仍直接组合 `useContestEditPage()`、topbar、stage tabs、workspace panel 与 redirect shell，因此适合沿同一模式收口。

## Decision
refactor_existing

## Reason
这轮最小正确改动不是去碰 `useContestEditPage()` 内部的 AWD workbench、save redirect 或 query-tab owner，而是新增 `PlatformContestEditPage.vue`，把：

- `useContestEditPage(toRef(props, 'contestId'))`
- `AppRouteRedirect`
- `ContestEditTopbarPanel`
- `ContestWorkbenchStageTabs`
- `ContestEditWorkspacePanel`
- 本地 loading shell 与 workspace 布局

统一收回 `features/platform/contests/ui` 内部。`ContestEditRoutePage.vue` 继续只保留 route prop 输入并渲染 feature page，从而把 route-level owner 压到最小，同时保持 `contestId` prop contract 不变。

## Files to modify
- `.harness/reuse-decisions/platform-contest-edit-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-platform-contest-edit-page-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-platform-contest-edit-page-owner-cleanup-review.md`
- `code/frontend/src/features/platform/contests/ui/PlatformContestEditPage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/pages/platform/contests/ContestEditRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 竞赛编辑 route page 会退回 “route prop -> feature page” 的薄壳。
- `platform/contests` 内会新增明确的 edit page-level owner，这组 route owner 收口会进一步接近完成态。
- `useContestEditPage.ts` 的 AWD workbench、保存、redirect 与 query-tab owner 保持不变。
