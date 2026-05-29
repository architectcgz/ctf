# Reuse Decision

## Change type
frontend refactor / feature-owned AWD traffic panel decomposition

## Existing code searched
- code/frontend/src/features/awd-inspector/ui/AWDTrafficPanel.vue
- code/frontend/src/features/awd-inspector/model/useAwdTrafficPanel.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/reviews/frontend/2026-05-28-awd-round-inspector-decomposition-review.md
- docs/reviews/frontend/2026-05-29-awd-service-status-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `AWDServiceStatusPanel.vue`、`AWDInstanceOrchestrationPanel.vue` 和 `AWDChallengeConfigPanel.vue` 都已经按“父层保留唯一 view model / workflow owner，稳定 summary、section、table 和 CSS 下沉，raw-source 护栏切到聚合源码”的模式完成收口。
- `AWDTrafficPanel.vue` 当前的债也一致：feature owner 已经正确，`useAwdTrafficPanel()` 也已经承担局部交互 owner，但父层仍同时混放 metric band、intelligence grid、event drill-down 和整段 scoped CSS。

## Decision
refactor_existing

## Reason
`AWDTrafficPanel.vue` 当前约 `450` 行，主要由：

- `useAwdTrafficPanel()` workflow wiring
- metric summary band
- intelligence grid（热点实体分析 + 12-bucket trend）
- event drill-down filters / table / pagination
- 大段 scoped CSS

组成。最小正确改动不是改 `useAwdTrafficPanel()`，也不是把逻辑推回 `AWDRoundInspector.vue`，而是：

- 保持 `AWDTrafficPanel.vue` 继续作为唯一 props / emits contract、`useAwdTrafficPanel()` owner 与 label/option presentation owner
- 新增 `AWDTrafficSummaryBand.vue` 承接 metric band
- 新增 `AWDTrafficIntelligenceGrid.vue` 承接热点实体分析与趋势区
- 新增 `AWDTrafficEventTable.vue` 承接 filters、table 与 pagination
- 新增 `awdTrafficPanel.css` 与 `awdTrafficPanel.types.ts` 承接样式和局部展示类型
- 新增 extraction 护栏，并把 primitive / theme / alignment 测试切到聚合源码视角

## Files to modify
- .harness/reuse-decisions/awd-traffic-panel-decomposition.md
- docs/plan/impl-plan/2026-05-29-awd-traffic-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-29-awd-traffic-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/awd-inspector/ui/AWDTrafficPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDTrafficSummaryBand.vue
- code/frontend/src/features/awd-inspector/ui/AWDTrafficIntelligenceGrid.vue
- code/frontend/src/features/awd-inspector/ui/AWDTrafficEventTable.vue
- code/frontend/src/features/awd-inspector/ui/awdTrafficPanel.css
- code/frontend/src/features/awd-inspector/ui/awdTrafficPanel.types.ts
- code/frontend/src/components/platform/__tests__/AWDTrafficPanelExtraction.test.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts

## After implementation
- `AWDTrafficPanel.vue` 会回到“traffic props / emits contract + `useAwdTrafficPanel()` owner + presentation owner”这一层，不再继续直接承接 summary、intelligence、event table 和整段样式。
- 流量态势的用户可见行为保持不变：热点统计、趋势条、filter、search、table、pagination 和 reset 行为都不变。
- backlog 中当前这条 contest / AWD feature 内残余大 surface，会把 `AWDTrafficPanel.vue` 从高优先 residual 候选里收掉，后续继续看 `AWDAttackLogDialog.vue`、`AWDServiceCheckDialog.vue` 这类仍混写 form / section / CSS 的 surface。
