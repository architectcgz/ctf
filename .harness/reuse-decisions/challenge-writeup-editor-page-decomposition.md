# Reuse Decision

## Change type
frontend refactor / feature-owned challenge writeup editor page decomposition

## Existing code searched
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupEditorPage.vue
- code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupEditorPage.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `ClassReportExportDialog.vue`、`PlatformContestFormPanel.vue`、`ContestProjectorAttackMap.vue` 都已经按“父层保留唯一 workflow / shell owner，稳定展示区下沉，raw-source 护栏切到聚合源码视角”的模式完成收口；`ChallengeWriteupEditorPage.vue` 也适合沿同一模式继续收口。
- `useChallengeWriteupEditorPage()` 已经把加载、保存、删除、推荐切换、表单恢复与提示都收成单点 workflow owner；这说明当前主问题不在 workflow，而在 page SFC 仍继续承载多个稳定 section 与大段样式。
- 当前页面内部的 editor form、saved snapshot、challenge info rail 都是稳定展示区块，拆成子组件不会破坏现有 API 或路由壳组合方式。

## Decision
refactor_existing

## Reason
`ChallengeWriteupEditorPage.vue` 当前约 `670` 行，父组件同时混放：

- page shell owner：`embedded` / `back` contract、topbar、`PageHeader`、embedded heading
- workflow wiring：`useChallengeWriteupEditorPage(props.challengeId)`
- 多个稳定展示区块：editor form、snapshot、challenge rail
- 大体量 page 样式

最小正确改动不是继续把展示结构塞进 composable，也不是维持单文件承载所有 section，而是：

- 保持 `ChallengeWriteupEditorPage.vue` 继续作为 shell owner 与 workflow wiring owner
- 新增 `ChallengeWriteupEditorFormSection.vue` 承接编辑器表单、badge、visibility note 和 editor actions
- 新增 `ChallengeWriteupSnapshotSection.vue` 承接当前已保存版本 snapshot / empty state
- 新增 `ChallengeWriteupChallengeRail.vue` 承接 challenge meta rail
- 新增 `challengeWriteupEditorPage.css` 承接 page shell 与 section 样式
- 同步把 raw-source 护栏改成聚合源码视角

本轮不调整 `useChallengeWriteupEditorPage()` 的 workflow owner，不改 route page composition，不改 `ChallengeWriteupViewPage.vue` 和 `ChallengeWriteupManagePanel.vue`。

## Files to modify
- .harness/reuse-decisions/challenge-writeup-editor-page-decomposition.md
- docs/plan/impl-plan/2026-05-28-challenge-writeup-editor-page-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-challenge-writeup-editor-page-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupEditorPage.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupEditorFormSection.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupSnapshotSection.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupChallengeRail.vue
- code/frontend/src/features/challenge-writeup-editor/ui/challengeWriteupEditorPage.css
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts

## After implementation
- `ChallengeWriteupEditorPage.vue` 会回到“page shell owner + workflow wiring owner”这一层，不再继续内联 editor / snapshot / rail 和整段样式。
- 平台侧题解管理入口、嵌入态体验、保存/删除/推荐行为保持不变，但 feature 内部 owner 会更清晰。
- backlog 里的 feature 内大组件债会补上一条这次收口记录，便于后续继续评估 `ChallengeWriteupManagePanel.vue` 和 `AWDRoundInspector.vue`。
