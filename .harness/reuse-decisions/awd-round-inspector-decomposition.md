# Reuse Decision

## Change type
frontend refactor / feature-owned AWD round inspector decomposition

## Existing code searched
- code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.vue
- code/frontend/src/features/awd-inspector/ui/awdInspector.types.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/views/platform/__tests__/ContestOperations.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/architecture/features/AWD巡检结果视图架构.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `ChallengeWriteupManagePanel.vue`、`ChallengeWriteupEditorPage.vue`、`ClassReportExportDialog.vue` 都已经按“父层保留唯一 workflow / state owner，稳定展示区和样式下沉，raw-source 护栏切到聚合源码视角”的模式完成收口；`AWDRoundInspector.vue` 适合沿同一模式继续收口。
- `AWDRoundInspector.vue` 之前已经把 header、service status、attack log、traffic、scoreboard 拆成独立 panel；当前主问题不再是 table/filter 还堆在父层，而是父层继续承载 stats HUD、sub-tab shell、loading/empty canvas 和 scoped CSS。
- `ContestOperations.vue` 现在只负责 route page 组合，并通过 `AWDOperationsPanel` 组合 `AWDRoundInspector`；本轮不需要改变上层 route / panel owner。

## Decision
refactor_existing

## Reason
`AWDRoundInspector.vue` 当前约 `545` 行，父组件同时混放：

- inspector owner：`initialTab`、`activeSubTab`、`open:contestEdit`
- workflow wiring：`useAwdInspectorCoreState`、`useAwdInspectorFormatting`、`useAwdInspectorDerivedData`、`useAwdInspectorExports`
- 稳定展示区：stats HUD、canvas tabs header、loading/empty shell、pane workspace
- 大段 scoped 样式

最小正确改动不是继续把更多展示壳塞回各个子 panel，也不是让父层继续混放 HUD / canvas / CSS，而是：

- 保持 `AWDRoundInspector.vue` 继续作为 props / emits / slots contract、composable wiring、`activeSubTab` owner、export / traffic forwarding owner
- 新增 `AWDInspectorStatsHud.vue` 承接四张 summary HUD 卡
- 新增 `AWDInspectorCanvasWorkspace.vue` 承接 sub-tabs、导出按钮、loading/empty shell 与四个 pane 的组合
- 新增 `awdRoundInspector.css` 承接 inspector shell、HUD、canvas、tab 与 loading 样式
- 同步把读取 `AWDRoundInspector.vue?raw` 的护栏改成聚合源码视角

本轮不调整：

- `AWDRoundHeaderPanel.vue`、`AWDServiceStatusPanel.vue`、`AWDAttackLogPanel.vue`、`AWDTrafficPanel.vue`、`AWDScoreboardSummaryPanel.vue` 的内部 owner
- `ContestOperations.vue`、`AWDOperationsPanel.vue` 的组合方式
- `useAwdInspector*` model 层的职责边界

## Files to modify
- .harness/reuse-decisions/awd-round-inspector-decomposition.md
- docs/plan/impl-plan/2026-05-28-awd-round-inspector-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-awd-round-inspector-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.vue
- code/frontend/src/features/awd-inspector/ui/AWDInspectorStatsHud.vue
- code/frontend/src/features/awd-inspector/ui/AWDInspectorCanvasWorkspace.vue
- code/frontend/src/features/awd-inspector/ui/awdRoundInspector.css
- code/frontend/src/features/awd-inspector/ui/awdInspector.types.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts

## After implementation
- `AWDRoundInspector.vue` 会回到“contract + active tab + inspector workflow wiring”这一层，不再继续内联 HUD、canvas shell 和整段样式。
- AWD 巡检页的用户可见行为保持不变：轮次切换、导出复盘包、matrix / scoreboard / attacks / traffic 切换、加载和空态都不变。
- backlog 里的 feature 内残余大组件债会把本轮收口记录进去，剩余重点会进一步收敛到 `AWDOperationsPanel.vue` 的更深层 workflow handler / dialog cluster 和其它仍在 feature 内混写的 surface。
