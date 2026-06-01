# Platform Contest Announcements Page Owner Cleanup 计划

## Objective

- 把 `ContestAnnouncementsRoutePage.vue` 收口成纯 route 薄壳。
- 在 `features/platform/contests/ui` 新增 `PlatformContestAnnouncementsPage.vue`，统一承接单场公告页的 page model 与 page shell owner。

## Non-goals

- 不改 `useContestAnnouncementsPage.ts` 的数据加载、发布公告、删除公告、toast 策略或 back route contract。
- 不继续拆 `ContestAnnouncementsTopbarPanel.vue` 或 `ContestAnnouncementsWorkspacePanel.vue`。
- 不顺手处理 `ContestEditRoutePage.vue` 或 `ContestOperationsRoutePage.vue`。

## Source Inputs

- `code/frontend/src/pages/platform/contests/ContestAnnouncementsRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestAnnouncementsTopbarPanel.vue`
- `code/frontend/src/features/platform/contests/ui/ContestAnnouncementsWorkspacePanel.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsHubPage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这页比 `ContestEditRoutePage.vue` 更适合作为下一刀，因为它只有一个 route prop，没有额外 overlay 或 stage workflow，owner 收口面清楚。
- 保留 `contestId` route prop 透传能让这轮只处理 route shell 和 feature page owner，不扩大到 route param 输入边界本身。

## Task Slices

### Slice 1: 新增 announcements feature page

- 目标：新增 `PlatformContestAnnouncementsPage.vue`，承接 `useContestAnnouncementsPage()`、topbar panel、loading/error shell 和 workspace panel 组合。
- 风险：表单字段双向绑定和 submit/delete 事件桥接必须保持现有行为。

### Slice 2: 收口 route page 与 public API

- 目标：让 `ContestAnnouncementsRoutePage.vue` 只接受 `contestId` prop 并渲染 `PlatformContestAnnouncementsPage`，通过 `@/features/platform/contests` 公共出口读取。
- 风险：如果 route page 仍直接引用 page model 或 panel，owner 收口不完整。

### Slice 3: 护栏与 backlog 同步

- 目标：更新 `ContestAnnouncements.test.ts` 的 source-boundary 断言和 backlog 进展，防止 route page 再次直接持有 page model。
- 风险：如果只改实现不改断言，这条路由页很容易回到 “薄壳 + page model” 混合状态。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-announcements-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-contest-announcements-page-owner-cleanup.md docs/plan/impl-plan/2026-06-01-platform-contest-announcements-page-owner-cleanup-plan.md docs/reviews/frontend/2026-06-01-platform-contest-announcements-page-owner-cleanup-review.md code/frontend/src/features/platform/contests/ui/PlatformContestAnnouncementsPage.vue code/frontend/src/features/platform/contests/ui/index.ts code/frontend/src/pages/platform/contests/ContestAnnouncementsRoutePage.vue code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `PlatformContestAnnouncementsPage.vue` 是否真正成为 announcements 的 page-level owner，而不是 route page 内容的机械搬家。
- `ContestAnnouncementsRoutePage.vue` 是否已经退回 “route prop -> feature page” 的薄壳。
- source-boundary 护栏是否足够防止 route page 再次直接拿 `useContestAnnouncementsPage()` 或 panel。
