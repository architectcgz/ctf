# Reuse Decision

## Change type
frontend refactor / feature-owned challenge writeup manage panel decomposition

## Existing code searched
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupManagePanel.vue
- code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `ChallengeWriteupEditorPage.vue`、`ClassReportExportDialog.vue`、`PlatformContestFormPanel.vue` 都已经按“父层保留唯一 workflow / shell owner，稳定展示区与样式下沉，raw-source 护栏切到聚合源码视角”的模式完成收口；`ChallengeWriteupManagePanel.vue` 也适合沿同一模式继续收口。
- `useChallengeWriteupManagement()` 已经把官方题解读取、投稿目录分页、删除动作与提示收成单点 workflow owner；当前主问题不是 workflow 未收口，而是 panel SFC 继续承载 header、summary、directory row、pagination 和大段样式。
- `ChallengeWriteupManagePanel` 当前可以稳定切出 header、summary strip、directory section、directory row，不需要改变 route view、detail workspace 或 writeup API owner。

## Decision
refactor_existing

## Reason
`ChallengeWriteupManagePanel.vue` 当前约 `590` 行，父组件同时混放：

- panel owner：`openWriteup`、`handleDelete`、`actionMenuOpen`
- workflow wiring：`useChallengeWriteupManagement({ challengeId })`
- 多个稳定展示区：页头、summary strip、directory section、directory row / action menu
- 大体量 panel 样式

最小正确改动不是继续把目录结构塞进 composable，也不是保持单文件承载所有 section，而是：

- 保持 `ChallengeWriteupManagePanel.vue` 继续作为 workflow wiring、writeup open/delete 事件 owner 和 action menu state owner
- 新增 `ChallengeWriteupManageHeader.vue` 承接页头与“编写题解”主动作
- 新增 `ChallengeWriteupSummaryStrip.vue` 承接官方题解 / 学员题解 summary strip
- 新增 `ChallengeWriteupDirectorySection.vue` 承接 directory heading、loading/empty、pagination 与 row 列表组合
- 新增 `ChallengeWriteupDirectoryRow.vue` 承接单条目录 row 与官方题解 action menu
- 新增 `challengeWriteupManagePanel.css` 承接 panel shell 与 directory 样式
- 同步把 raw-source 护栏改成聚合源码视角

本轮不调整 `useChallengeWriteupManagement()` 的 workflow owner，不改 `PlatformChallengeDetailWorkspace` 组合方式，不改 `ChallengeWriteupEditorPage.vue` 或 `ChallengeWriteupViewPage.vue`。

## Files to modify
- .harness/reuse-decisions/challenge-writeup-manage-panel-decomposition.md
- docs/plan/impl-plan/2026-05-28-challenge-writeup-manage-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-challenge-writeup-manage-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupManagePanel.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupManageHeader.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupSummaryStrip.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupDirectorySection.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupDirectoryRow.vue
- code/frontend/src/features/challenge-writeup-editor/ui/challengeWriteupManagePanel.css
- code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts

## After implementation
- `ChallengeWriteupManagePanel.vue` 会回到“workflow wiring + open/delete/menu owner”这一层，不再继续内联 header / summary / directory row 与整段样式。
- 平台题目详情中的题解目录行为保持不变，但 feature 内部 owner 会更清晰，后续若继续瘦身，只需在目录 section 局部继续处理。
- backlog 里的 feature 内残余大组件债会补上一条这次收口记录，便于后续继续评估 `AWDRoundInspector.vue`。
