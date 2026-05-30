# Reuse Decision

## Change type
frontend refactor / feature-owned AWD challenge config panel decomposition

## Existing code searched
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue
- code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts
- code/frontend/src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts
- code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/plan/impl-plan/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `ChallengeWriteupManagePanel.vue`、`AWDRoundInspector.vue`、`ContestChallengeOrchestrationPanel.vue` 都已经按“父层保留唯一 workflow / presentation owner，稳定 header / summary / directory section / row / CSS 下沉，raw-source 护栏切到聚合源码”的模式完成收口。
- `AWDChallengeConfigPanel.vue` 当前的真实问题也一致：目录 owner 已经在正确 feature 内，但父组件仍同时混放 header、summary strip、directory table、row presentation 和大段 scoped CSS。

## Decision
refactor_existing

## Reason
`AWDChallengeConfigPanel.vue` 当前约 `562` 行，主要由：

- AWD challenge directory 的排序与 summary 派生
- `useAwdCheckResultPresentation()` 提供的校验结果 presentation helper 接线
- 稳定 header / summary 视图
- directory table 和 row 模板
- 整段 scoped CSS

组成。最小正确改动不是再引入新的跨 feature 依赖，也不是把 presentation helper 散到 row 组件，而是：

- 保持 `AWDChallengeConfigPanel.vue` 继续作为唯一 props / emits contract、排序与 AWD 校验 presentation owner
- 新增 `AWDChallengeConfigHeader.vue` 承接 overline / title / description 与 summary strip
- 新增 `AWDChallengeConfigDirectorySection.vue` 承接 empty state、table shell 和 row 列表
- 新增 `AWDChallengeConfigDirectoryRow.vue` 承接单行 challenge identity / checker / score / validation / actions 展示
- 新增 `awdChallengeConfigPanel.css` 承接 panel、table、row 和 validation 的样式
- 把 raw-source 护栏改成聚合源码视角，继续覆盖 primitive / token / dark surface 对齐

## Files to modify
- .harness/reuse-decisions/awd-challenge-config-panel-decomposition.md
- docs/plan/impl-plan/2026-05-29-awd-challenge-config-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-29-awd-challenge-config-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigHeader.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue
- code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue
- code/frontend/src/features/platform-contests/ui/awdChallengeConfigPanel.css
- code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanelExtraction.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts

## After implementation
- `AWDChallengeConfigPanel.vue` 会回到“排序 + 校验 presentation helper + directory view model + outward edit emit”这一层，不再继续直接承接 header、table row 和整段样式。
- AWD 配置目录的用户可见行为保持不变：预览链接、编辑按钮、checker 标签、规则摘要和就绪验证提示都不变。
- backlog 中这条 contest / AWD feature 内残余大 surface 会继续收口，`AWDChallengeConfigPanel.vue` 不再是同级别高优先候选，后续 focus 会进一步转向其它仍未拆的 feature surface。
