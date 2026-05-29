# Reuse Decision

## Change type
frontend refactor / feature-owned AWD instance orchestration panel decomposition

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanelExtraction.test.ts
- code/frontend/src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- docs/plan/impl-plan/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-review.md
- docs/reviews/frontend/2026-05-28-awd-operations-panel-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `AWDChallengeConfigPanel.vue`、`ContestChallengeOrchestrationPanel.vue` 和 `AWDOperationsPanel.vue` 都已经按“父层保留唯一 view model / workflow owner，稳定 header、section、row 和 CSS 下沉，raw-source 护栏切到聚合源码”的模式完成收口。
- `AWDInstanceOrchestrationPanel.vue` 当前的债也一致：feature owner 已经正确，但父层仍同时混放 summary、empty state、matrix table、cell action、row action 和整段 scoped CSS。

## Decision
refactor_existing

## Reason
`AWDInstanceOrchestrationPanel.vue` 当前约 `464` 行，主要由：

- instance map / visible services / running summary 这组实例编排视图模型
- header summary 与全局动作按钮
- empty state
- matrix table、team row、service cell
- 大段 scoped CSS

组成。最小正确改动不是把逻辑推回 `AWDOperationsPanel.vue`，也不是引入新的 feature model，而是：

- 保持 `AWDInstanceOrchestrationPanel.vue` 继续作为唯一 props / emits contract、实例编排 view model 和 row/cell action owner
- 新增 `AWDInstanceOrchestrationHeader.vue` 承接标题、running summary、refresh / start-all actions
- 新增 `AWDInstanceOrchestrationMatrix.vue` 承接 empty state、table shell 与行列表
- 新增 `AWDInstanceOrchestrationRow.vue` 承接单队伍行和服务单元格展示
- 新增 `awdInstanceOrchestrationPanel.css` 承接 panel、matrix、cell 和状态样式
- 新增 extraction 护栏，并把 primitive / token / surface 对齐测试切到聚合源码视角

## Files to modify
- .harness/reuse-decisions/awd-instance-orchestration-panel-decomposition.md
- docs/plan/impl-plan/2026-05-29-awd-instance-orchestration-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-29-awd-instance-orchestration-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationHeader.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationMatrix.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationRow.vue
- code/frontend/src/features/contest-awd-admin/ui/awdInstanceOrchestrationPanel.css
- code/frontend/src/features/contest-awd-admin/ui/awdInstanceOrchestration.types.ts
- code/frontend/src/components/platform/__tests__/AWDInstanceOrchestrationPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts

## After implementation
- `AWDInstanceOrchestrationPanel.vue` 会回到“instance orchestration view model + outward emits owner”这一层，不再继续直接承接 header、matrix row 和整段样式。
- 实例编排的用户可见行为保持不变：刷新、启动全部、启动本队、启动单元格、状态标签和访问链接都不变。
- backlog 中当前这条 contest / AWD feature 内残余大 surface，会把 `AWDInstanceOrchestrationPanel.vue` 从高优先 residual 候选里收掉，后续再继续看其它仍混写 matrix / row / CSS 的 feature panel。
