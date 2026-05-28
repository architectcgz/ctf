# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDRoundInspector.vue
- code/frontend/src/components/platform/contest/AWDTrafficPanel.vue
- code/frontend/src/components/platform/contest/AWDServiceStatusPanel.vue
- code/frontend/src/components/platform/contest/AWDScoreboardSummaryPanel.vue
- code/frontend/src/components/platform/contest/AWDRoundHeaderPanel.vue
- code/frontend/src/components/platform/contest/AWDAttackLogPanel.vue
- code/frontend/src/components/platform/contest/AWDServiceAlertBanner.vue
- code/frontend/src/components/platform/contest/awdInspector.types.ts
- code/frontend/src/features/awd-inspector/index.ts
- code/frontend/src/features/awd-inspector/model/index.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/ContestOperations.vue
- code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts
- code/frontend/src/views/platform/__tests__/ContestOperations.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `AWDOperationsPanel.vue` 已迁入 `features/contest-awd-admin/ui`，说明 AWD 运维主面板 owner 已经与 `contest-awd-admin` 对齐。
- `features/awd-inspector` 当前已经承接 `useAwdInspectorCoreState`、`useAwdInspectorDerivedData`、`useAwdInspectorExports`、`useAwdInspectorFormatting`、`useAwdTrafficPanel` 等完整 inspector model owner。
- `architectureAllowlist.ts` 当前仍保留 `AWDRoundInspector.vue`、`AWDTrafficPanel.vue`、`awdInspector.types.ts -> @/features/awd-inspector` 的历史例外，表明 UI owner 早已确认在 `awd-inspector`。

## Decision
refactor_existing

## Reason
`AWDRoundInspector` 及其直接依赖的 round header / traffic / service status / scoreboard / attack log / alert banner / shared types，语义上都属于 `awd-inspector` 单一 feature 的 UI cluster，不应继续滞留在旧 `components/platform/contest/*` 路径。最小正确改动是：

- 新增 `features/awd-inspector/ui/index.ts`
- 把 `AWDRoundInspector.vue`、`AWDTrafficPanel.vue`、`AWDServiceStatusPanel.vue`、`AWDScoreboardSummaryPanel.vue`、`AWDRoundHeaderPanel.vue`、`AWDAttackLogPanel.vue`、`AWDServiceAlertBanner.vue`、`awdInspector.types.ts` 一起迁入 `features/awd-inspector/ui`
- `features/awd-inspector/index.ts` 暴露这组 UI public API
- 更新 `AWDOperationsPanel.vue`、`ContestOperations.vue`、相关测试、`components.d.ts` 与 `architectureAllowlist.ts`

本轮不迁 `AWDReadiness*`、`AWDInstanceOrchestrationPanel.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 等仍更贴近 `contest-awd-admin` 的 runtime / dialog cluster。

## Files to modify
- .harness/reuse-decisions/awd-inspector-ui-cluster-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-inspector-ui-cluster-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-inspector-ui-cluster-feature-ui-normalization-review.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/features/awd-inspector/index.ts
- code/frontend/src/features/awd-inspector/ui/index.ts
- code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.vue
- code/frontend/src/features/awd-inspector/ui/AWDTrafficPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDScoreboardSummaryPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDRoundHeaderPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDAttackLogPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDServiceAlertBanner.vue
- code/frontend/src/features/awd-inspector/ui/awdInspector.types.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/ContestOperations.vue
- code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts
- code/frontend/src/views/platform/__tests__/ContestOperations.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `AWDRoundInspector` cluster 会归 `features/awd-inspector/ui` 持有。
- `AWDOperationsPanel.vue` 与 `ContestOperations.vue` 会通过 `@/features/awd-inspector` public API 组合 inspector UI。
- `architectureAllowlist.ts` 中这 3 条 `awd-inspector` 历史例外会收掉。
