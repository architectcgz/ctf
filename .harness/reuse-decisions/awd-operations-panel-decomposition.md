# Reuse Decision

## Change type
frontend refactor / feature-owned awd operations panel decomposition

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDContestSelectorField.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePendingState.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.vue
- code/frontend/src/features/awd-readiness/ui/AWDReadinessSummary.vue
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `PlatformContestFormPanel.vue` 与 `ContestChallengeOrchestrationPanel.vue` 都已经按“父层保留唯一 workflow owner，稳定展示分区下沉”完成拆分，本轮 `AWDOperationsPanel.vue` 应沿同一模式推进。
- `AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue`、`AWDInstanceOrchestrationPanel.vue`、`AWDRoundInspector.vue`、`AWDReadinessSummary.vue` 已经各自拥有清楚的 capability owner，本轮不应把这些下游 capability 再重新并回父文件。
- `usePlatformContestAwd()` 已经是 AWD 运维面板唯一的 runtime workflow owner，本轮更合适的是把 pre-runtime / runtime 的壳体组合下沉，而不是改动 composable contract。
- 当前父文件在阶段壳体拆分后，剩余最肥的模板 cluster 是 `AWDRoundCreateDialog` / `AWDServiceCheckDialog` / `AWDAttackLogDialog` / `AWDReadinessOverrideDialog`；这一段也适合继续用“父层保留 dialog state owner，子层只做展示桥接”的同一模式收口。

## Decision
refactor_existing

## Reason
`AWDOperationsPanel.vue` 当前约 `643` 行，真正应该保留在父层的逻辑主要是：

- `selectedContest`、`runtimeStageReady`、`activePanel` 这类顶层视图状态
- `usePlatformContestAwd()` 的 model wiring
- dialog open state 与 create / save / override 事件桥接
- tab keyboard navigation owner

剩余大部分模板属于两个稳定阶段的组合壳：

- pre-runtime：tabs + header + pending/readiness + instance orchestration
- runtime：tabs + readiness strip + round inspector / instance orchestration

最小正确改动不是改 `usePlatformContestAwd()`，也不是再新增 store，而是：

- 保持 `AWDOperationsPanel.vue` 继续做唯一 runtime workflow owner、tab owner 和 dialog owner
- 新增 `AWDOperationsTabs.vue` 承接 tabs 视图
- 新增 `AWDOperationsPreRuntimeStage.vue` 承接未开赛阶段壳体
- 新增 `AWDOperationsRuntimeStage.vue` 承接运行中阶段壳体
- 新增 `AWDOperationsDialogHub.vue` 承接 live region 与 dialog cluster 展示桥接

本轮不调整 `AWDRoundInspector.vue`、`AWDReadinessSummary.vue`、`AWDInstanceOrchestrationPanel.vue` 的对外 contract，不改变 `ContestOperations.vue` 的消费方式。

## Files to modify
- .harness/reuse-decisions/awd-operations-panel-decomposition.md
- docs/plan/impl-plan/2026-05-28-awd-operations-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-awd-operations-panel-decomposition-review.md
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsTabs.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsRuntimeStage.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `AWDOperationsPanel.vue` 会回到 “contest selection guard + tab owner + runtime workflow owner + dialog owner” 这一层，不再继续内联两个阶段的大块壳体模板和 dialog cluster。
- `contest-awd-admin/ui` 会形成更明确的 operations stage cluster，后续若继续处理 `AWDInstanceOrchestrationPanel.vue` 或 readiness/runtime 布局债，可以在正确 owner 下继续分刀。
- 当前 backlog 里 `AWDOperationsPanel.vue` 这一项会从“超大组件待拆”转成已收口或至少显著收口。
