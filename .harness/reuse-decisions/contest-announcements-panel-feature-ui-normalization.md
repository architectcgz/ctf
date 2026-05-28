# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/views/platform/ContestAnnouncements.vue
- code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/components/platform/contest/ContestAnnouncementsTopbarPanel.vue
- code/frontend/src/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue
- code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts
- code/frontend/src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestEditTopbarPanel.vue`、`ContestEditWorkspacePanel.vue` 已迁入 `features/platform-contests/ui`，说明 contest 路由页的大颗粒 feature UI 已经按 owner 收口。
- `PlatformContestFormPanel.vue`、`PlatformContestTable.vue` 也已经在 `features/platform-contests/ui`，并通过 `features/platform-contests` public API 供 route shell 组合。
- `ContestAnnouncements.vue` 当前已经把页面 owner 收到 `useContestAnnouncementsPage()`，route shell 只负责 loading / error / panel 组合。

## Decision
refactor_existing

## Reason
`ContestAnnouncementsTopbarPanel.vue` 和 `ContestAnnouncementsWorkspacePanel.vue` 只被 `ContestAnnouncements.vue` 使用，且语义上属于 `platform-contests` 单一 feature 的 route-owned UI，不应继续滞留在旧 `components/platform/contest/*` 目录。最小正确改动是：

- 把两个 panel 迁入 `features/platform-contests/ui`
- `ContestAnnouncements.vue` 改为通过 `@/features/platform-contests` public API 组合它们
- 更新 `features/platform-contests/ui/index.ts`、`components.d.ts`、相关 raw-source 测试
- 在 backlog 记录这条 feature UI 收口进展

本轮不调整 `useContestAnnouncementsPage()`、`useContestAnnouncementManagement()` 的 owner，也不改公告读写行为。

## Files to modify
- .harness/reuse-decisions/contest-announcements-panel-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-contest-announcements-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-contest-announcements-panel-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestAnnouncementsWorkspacePanel.vue
- code/frontend/src/views/platform/ContestAnnouncements.vue
- code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts
- code/frontend/src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `ContestAnnouncementsTopbarPanel.vue` 与 `ContestAnnouncementsWorkspacePanel.vue` 会归 `features/platform-contests/ui` 持有。
- `ContestAnnouncements.vue` 与对应 raw-source 测试不再引用旧 `components/platform/contest/*` 路径。
- `platform-contests` 这条 feature owner 下的公告页面大颗粒 UI 会继续从 legacy components 收口到 feature public API。
