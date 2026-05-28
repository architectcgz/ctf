# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/ContestEditTopbarPanel.vue
- code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue
- code/frontend/src/components/platform/contest/ContestAwdChallengeSelectorSection.vue
- code/frontend/src/components/platform/contest/ContestChallengeSettingsSection.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/contest-workbench/ui/index.ts
- code/frontend/src/views/platform/ContestEdit.vue
- code/frontend/src/views/platform/__tests__/contestEditTopbarExtraction.test.ts
- code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `contest-announcements` 与 `contest-workbench` 前两轮已经按“feature public API + feature-owned ui + route shell 组合”收掉历史路径。
- `ContestEditWorkspacePanel.vue` 已经迁入 `features/platform-contests/ui`，`ContestChallengeOrchestrationPanel.vue` 已经迁入 `features/contest-workbench/ui`。
- `ContestEditTopbarPanel.vue` 现在只服务 `ContestEdit.vue` 这条 route shell，`ContestChallengeEditorDialog.vue` 与其两个 section 只服务 `contest-workbench` 的题目编排面板。

## Decision
refactor_existing

## Reason
本轮 touched surface 已经把 contest edit 的 workspace 和 workbench 子面板迁走，继续把 topbar 与 challenge editor dialog 留在 `components/platform/contest` 会让同一条工作台仍存在两套 owner 落点。最小正确改动是：

- `ContestEditTopbarPanel.vue` 迁入 `features/platform-contests/ui`
- `ContestChallengeEditorDialog.vue`、`ContestAwdChallengeSelectorSection.vue`、`ContestChallengeSettingsSection.vue` 迁入 `features/contest-workbench/ui`
- `ContestEdit.vue` 与 `ContestChallengeOrchestrationPanel.vue` 改为通过 feature public API 组合

这样能把 contest edit 线上剩余的单一 feature UI 壳继续从旧组件目录里清出去。

## Files to modify
- .harness/reuse-decisions/contest-edit-surface-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-27-contest-edit-surface-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-27-contest-edit-surface-feature-ui-normalization-review.md
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue
- code/frontend/src/features/contest-workbench/ui/index.ts
- code/frontend/src/features/contest-workbench/ui/ContestChallengeEditorDialog.vue
- code/frontend/src/features/contest-workbench/ui/ContestAwdChallengeSelectorSection.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeSettingsSection.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/views/platform/ContestEdit.vue
- code/frontend/src/components.d.ts
- code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts
- code/frontend/src/views/platform/__tests__/contestEditTopbarExtraction.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `platform-contests` 会持有 contest edit topbar 这类 route shell UI。
- `contest-workbench` 会持有 challenge editor dialog 及其 selector/settings sections。
- `ContestEdit.vue` 和 `ContestChallengeOrchestrationPanel.vue` 不再回头引用旧 `components/platform/contest/*` 路径。
