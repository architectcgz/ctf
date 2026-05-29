# Reuse Decision

## Change type
frontend refactor / feature-owned AWD service status panel decomposition

## Existing code searched
- code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.vue
- code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue
- docs/reviews/frontend/2026-05-28-awd-round-inspector-decomposition-review.md
- docs/reviews/frontend/2026-05-29-awd-instance-orchestration-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `AWDInstanceOrchestrationPanel.vue`、`AWDChallengeConfigPanel.vue` 和 `AWDRoundInspector.vue` 都已经按“父层保留唯一 view model / workflow owner，稳定 toolbar、section、row 和 CSS 下沉，raw-source 护栏切到聚合源码”的模式完成收口。
- `AWDServiceStatusPanel.vue` 当前的债也一致：feature owner 已经正确，但父层仍同时混放 toolbar filters、matrix table、round performance summary 和整段 scoped CSS。

## Decision
refactor_existing

## Reason
`AWDServiceStatusPanel.vue` 当前约 `509` 行，主要由：

- 筛选条件和导出动作 toolbar
- 服务运行矩阵 table shell、队伍行和状态单元格
- 本轮得分与健康表现 summary table
- 大段 scoped CSS

组成。最小正确改动不是把逻辑推回 `AWDRoundInspector.vue`，也不是新增 feature model，而是：

- 保持 `AWDServiceStatusPanel.vue` 继续作为唯一 props / emits contract、筛选 forwarding 和 presentation model owner
- 新增 `AWDServiceStatusToolbar.vue` 承接标题、筛选器和导出动作
- 新增 `AWDServiceStatusMatrix.vue` 承接运行矩阵 table shell、空态和状态单元格展示
- 新增 `AWDServiceRoundPerformanceTable.vue` 承接 summary table
- 新增 `awdServiceStatusPanel.css` 与 `awdServiceStatusPanel.types.ts` 承接样式和局部 view model 类型
- 新增 extraction 护栏，并把 theme token 护栏切到聚合源码视角

## Files to modify
- .harness/reuse-decisions/awd-service-status-panel-decomposition.md
- docs/plan/impl-plan/2026-05-29-awd-service-status-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-29-awd-service-status-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.vue
- code/frontend/src/features/awd-inspector/ui/AWDServiceStatusToolbar.vue
- code/frontend/src/features/awd-inspector/ui/AWDServiceStatusMatrix.vue
- code/frontend/src/features/awd-inspector/ui/AWDServiceRoundPerformanceTable.vue
- code/frontend/src/features/awd-inspector/ui/awdServiceStatusPanel.css
- code/frontend/src/features/awd-inspector/ui/awdServiceStatusPanel.types.ts
- code/frontend/src/components/platform/__tests__/AWDServiceStatusPanelExtraction.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts

## After implementation
- `AWDServiceStatusPanel.vue` 会回到“service status props / emits contract + presentation owner”这一层，不再继续直接承接 toolbar、matrix、summary table 和整段样式。
- 服务状态矩阵的用户可见行为保持不变：筛选、导出、状态标签、元信息和 round performance summary 都不变。
- backlog 中当前这条 contest / AWD feature 内残余大 surface，会把 `AWDServiceStatusPanel.vue` 从高优先 residual 候选里收掉，后续继续看 `AWDTrafficPanel.vue`、`AWDAttackLogDialog.vue` 这类仍混写 section / CSS 的 surface。
