# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/index.ts
- code/frontend/src/features/awd-inspector/index.ts
- code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestEditTopbarPanel.vue` 已迁入 `features/platform-contests/ui`，contest edit 的 route shell UI 已开始按 `platform-contests` owner 收口。
- `ContestChallengeOrchestrationPanel.vue` 已迁入 `features/contest-workbench/ui`，`ContestEditWorkspacePanel.vue` 当前通过 feature public API + feature 内部 UI 组合 contest edit workspace。
- `features/awd-inspector` 当前只暴露 model，不持有独立 UI surface。

## Decision
refactor_existing

## Reason
`AWDChallengeConfigPanel.vue` 当前只服务 `ContestEditWorkspacePanel.vue` 的 AWD 配置阶段，是 contest edit workspace 里的单一展示分区；继续把它留在 `components/platform/contest/*`，会让同一条 route shell 仍保留一块历史 UI owner，同时冻结一条 `components -> @/features/awd-inspector` allowlist。最小正确改动是：

- 把 `AWDChallengeConfigPanel.vue` 迁入 `features/platform-contests/ui`
- `ContestEditWorkspacePanel.vue` 改用 feature 内部相对 import
- 删除 feature UI 中显式 `RouterLink` import，避免 router access 回流到 feature UI
- 更新 raw-source / surface / theme token / architecture 边界测试与 backlog

这样可以继续清空 contest edit 线上的 legacy component surface，同时在 touched surface 内收掉这条 allowlist。

## Files to modify
- .harness/reuse-decisions/awd-challenge-config-panel-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `platform-contests` 会持有 contest edit AWD 配置阶段的目录展示 panel。
- `ContestEditWorkspacePanel.vue` 不再回头引用旧 `components/platform/contest/AWDChallengeConfigPanel.vue`。
- `componentFeatureImportAllowlist` 里这条 `AWDChallengeConfigPanel.vue -> @/features/awd-inspector` 会在 touched surface 内收口。
