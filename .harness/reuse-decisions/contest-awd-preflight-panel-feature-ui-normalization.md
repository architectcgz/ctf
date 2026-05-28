# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/ContestAwdPreflightPanel.vue
- code/frontend/src/components/platform/contest/AWDReadinessChecklist.vue
- code/frontend/src/components/platform/contest/AWDReadinessDecisionHUD.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `PlatformContestTable.vue`、`PlatformContestFormPanel.vue`、`PlatformContestFormDialog.vue` 已迁入 `features/platform-contests/ui`，说明 `platform-contests` 线上的目录 / 表单 / 编辑壳体 UI 已按 feature owner 收口。
- `ContestEditWorkspacePanel.vue` 当前已经在 `features/platform-contests/ui`，并继续通过 feature 内部相对 import 组合 `AWDChallengeConfigPanel.vue`、`PlatformContestFormPanel.vue`。
- `AWDReadinessChecklist.vue` 同时被 `AWDReadinessSummary.vue` 复用，`AWDReadinessDecisionHUD.vue` 目前只被 `ContestAwdPreflightPanel.vue` 使用，但它们都还是更偏 AWD readiness primitive，而不是 route shell owner。

## Decision
refactor_existing

## Reason
`ContestAwdPreflightPanel.vue` 当前只被 `ContestEditWorkspacePanel.vue` 消费，是 contest edit surface 里的单一 feature 展示 UI。继续把它留在旧 `components/platform/contest/*` 路径，会让 `platform-contests` 线上的 AWD 赛前检查面板继续成为历史残片。最小正确改动是：

- 把 `ContestAwdPreflightPanel.vue` 迁入 `features/platform-contests/ui`
- `ContestEditWorkspacePanel.vue` 改成 feature 内部相对 import
- 保留它对 `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue` 的显式依赖，不在这一刀顺手迁移 readiness primitive
- 更新 `components.d.ts`、raw-source 测试和 backlog 记录

这样可以继续缩小 `platform-contests` 线上的 legacy component surface，同时不把 readiness primitive 和 route shell owner 混在同一刀里。

## Files to modify
- .harness/reuse-decisions/contest-awd-preflight-panel-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-contest-awd-preflight-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-contest-awd-preflight-panel-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `ContestAwdPreflightPanel.vue` 会归 `features/platform-contests/ui` 持有。
- `ContestEditWorkspacePanel.vue` 与相关 raw-source 测试不再引用旧 `components/platform/contest/ContestAwdPreflightPanel.vue` 路径。
- `platform-contests` 线上的 contest edit / awd preflight 展示残片会继续缩小。
