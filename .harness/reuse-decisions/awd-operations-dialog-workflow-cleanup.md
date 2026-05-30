# Reuse Decision

## Change type
frontend refactor / contest AWD admin dialog workflow owner cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogContracts.ts
- code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogOptions.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdRoundDetailState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ChallengeWriteupManagePanel`、`AWDRoundInspector` 与 `NotificationDrawer` 近期都已经收口到“父壳保留唯一 workflow / state owner，稳定 section 与样式下沉，局部 view-state 再落到 feature 内 composable”的模式。
- `ContestChallengeEditorDialog.vue` 当前仍保留 form / validation / submit owner，但其字段 section 已下沉；AWD 三个 dialog 的现状与它相似，适合继续把 form workflow 从 SFC 本体下沉到 feature 内 composable。
- `useAwdOperationsDialogState.ts` 已经是 dialog open/close 与 mutation close guard owner；本轮不需要把 dialog workflow 拉回 page model，而是继续把它收成更明确的 controller contract。
- `useAwdRoundOperations.ts` 与 `useAwdRoundDetailState.ts` 都属于 round list/detail 的远端数据 workflow owner，负责轮次读取、选择、刷新与服务/攻击明细同步；它们不是 dialog 本地 draft / validation / payload build owner，因此本轮不会把 dialog form state 继续塞回这些 model hook，而是新增 feature 内局部 form composable 承接纯本地对话框工作流。

## Decision
refactor_existing

## Reason
`AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 虽然已经不是超大模板，但仍同时混放：

- dialog shell / emits contract
- form draft 默认值
- open 时 reset
- validation / parse / payload build
- submit guard

同时 `AWDOperationsPanel.vue` 到 `AWDOperationsDialogHub.vue` 之间仍靠一组展开的 `open/saving/update/save` props 和 handlers 传递 dialog 状态，`useAwdOperationsDialogState.ts` 也还是“3 组并列 ref + 3 组并列 handler”。

最小正确收口方式是：

- 为 3 个 dialog 各自新增 feature 内 form workflow composable，承接 draft、open reset、validation、payload build
- `AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 回到 modal shell + section composition + save emit
- `useAwdOperationsDialogState.ts` 收成 runtime dialog controller owner，不再暴露一长串平铺的 open/update/handleCreate handler
- `AWDOperationsDialogHub.vue` 与 `AWDOperationsPanel.vue` 改用更明确的 grouped dialog binding contract，而不是继续把同一条 workflow 展开成多组散 props

本轮不调整：

- `usePlatformContestAwd` 的远端 mutation owner
- `AWDReadinessOverrideDialog` 的内部实现
- `AWDOperationsPreRuntimeStage` / `AWDOperationsRuntimeStage` 的舞台壳层

## Files to modify
- .harness/reuse-decisions/awd-operations-dialog-workflow-cleanup.md
- docs/plan/impl-plan/2026-05-30-awd-operations-dialog-workflow-cleanup-plan.md
- docs/reviews/frontend/2026-05-30-awd-operations-dialog-workflow-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogContracts.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdRoundCreateDialogForm.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdServiceCheckDialogForm.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdAttackLogDialogForm.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogForms.test.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts

## After implementation
- 三个 AWD dialog 的 SFC 会只保留 modal shell、section 组合、saving guard 和 emit contract。
- form draft / validation / payload build 会在 feature 内形成清晰的 per-dialog owner，不再继续黏在 SFC 顶部。
- `AWDOperationsPanel` 会只负责 dialog hub 组合与 runtime workflow wiring，不再继续维护展开式 dialog props/handlers。
