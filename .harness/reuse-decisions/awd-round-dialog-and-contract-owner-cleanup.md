# Reuse Decision

## Change type
frontend refactor / feature-owned AWD round dialog decomposition and operations dialog contract owner cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- docs/reviews/frontend/2026-05-29-awd-operations-dialog-cluster-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 刚完成的 `AWDRoundCreateDialog.vue`、`AWDAttackLogDialog.vue` 与 `AWDServiceCheckDialog.vue` 已按“父 dialog 保留唯一 form / validation / submit owner，稳定 section、footer 和 CSS 下沉，raw-source 护栏切到聚合源码”的模式完成收口。
- `useAwdOperationsDialogState.ts` 当前已经是 `contest-awd-admin` dialog cluster 的共享 workflow owner，但 payload contract、override state shape 和保存成功后关闭 dialog 语义仍分散写在多个文件内联声明。

## Decision
refactor_existing

## Reason
当前剩余债已经集中在一层：

- `AWDOperationsDialogHub.vue`、`useAwdOperationsDialogState.ts` 与 operations panel tabs 护栏仍重复感知：
  - create round / service check / attack log payload shape
  - override dialog state shape
  - “保存成功后关闭对应 dialog” 这组 workflow 语义

最小正确改动不是把 `useAwdOperationsDialogState.ts` 再继续硬拆成很多小 composable，而是：

- 保持三组 dialog 自身继续作为各自的 form / validation / submit owner
- 继续复用已落地的 `awdOperationsDialogContracts.ts`
- 在 `useAwdOperationsDialogState.ts` 内用局部 helper 收口“保存成功后关闭对应 dialog”的 shared workflow 语义
- 更新 operations tabs 护栏与 state 单测，确认 deeper owner 已集中到 `dialog hub + dialog state`

## Files to modify
- .harness/reuse-decisions/awd-round-dialog-and-contract-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-awd-round-dialog-and-contract-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-awd-round-dialog-and-contract-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts

## After implementation
- AWD operations dialog cluster 的 create payload 与 override state contract，会统一收口到 feature 内单点 type owner，不再在 `dialog hub`、`dialog state` 与各 dialog 文件中继续各自手写一遍。
- `useAwdOperationsDialogState.ts` 会继续保留为 dialog cluster 的 workflow owner，但“保存成功后关闭对应 dialog” 这组 shared workflow 语义不再重复写三次。
- backlog 中当前这条 contest / AWD feature 内 residual，会继续把更深层重点转到真正仍在增长的 workflow / contract owner，而不再是这组三个 dialog 的表单壳体。
