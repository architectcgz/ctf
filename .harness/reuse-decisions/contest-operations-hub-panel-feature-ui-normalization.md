# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/views/platform/ContestOperationsHub.vue
- code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/components/platform/contest/ContestOperationsHubHeroPanel.vue
- code/frontend/src/components/platform/contest/ContestOperationsHubWorkspacePanel.vue
- code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts
- code/frontend/src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestAnnouncementsTopbarPanel.vue`、`ContestAnnouncementsWorkspacePanel.vue` 已收口到 `features/platform-contests/ui`，说明 `platform-contests` 线上的 page-sized panel 正在按 feature owner 回收。
- `ContestEditTopbarPanel.vue`、`ContestEditWorkspacePanel.vue` 也已经在 `features/platform-contests/ui`，并通过 `features/platform-contests` public API 供 route shell 组合。
- `ContestOperationsHub.vue` 当前已经把页面 workflow 收到 `useContestOperationsHubPage()`，route shell 只负责 hero/workspace 组合。

## Decision
refactor_existing

## Reason
`ContestOperationsHubHeroPanel.vue` 和 `ContestOperationsHubWorkspacePanel.vue` 只被 `ContestOperationsHub.vue` 使用，且语义上属于 `platform-contests` 单一 feature 的 route-owned UI，不应继续滞留在旧 `components/platform/contest/*`。最小正确改动是：

- 把两个 panel 迁入 `features/platform-contests/ui`
- `ContestOperationsHub.vue` 改为通过 `@/features/platform-contests` public API 组合它们
- 更新 `features/platform-contests/ui/index.ts`、`components.d.ts`、相关 raw-source 测试
- 在 backlog 记录这条 feature UI 收口进展

本轮不调整 `useContestOperationsHubPage()` 的分页、推荐赛事或导航 owner，也不改运维目录行为。

## Files to modify
- .harness/reuse-decisions/contest-operations-hub-panel-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-contest-operations-hub-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-contest-operations-hub-panel-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue
- code/frontend/src/views/platform/ContestOperationsHub.vue
- code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts
- code/frontend/src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `ContestOperationsHubHeroPanel.vue` 与 `ContestOperationsHubWorkspacePanel.vue` 会归 `features/platform-contests/ui` 持有。
- `ContestOperationsHub.vue` 与对应 raw-source 测试不再引用旧 `components/platform/contest/*` 路径。
- `platform-contests` 这条 feature owner 下的赛事运维目录大颗粒 UI 会继续从 legacy components 收口到 feature public API。
