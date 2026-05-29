# Reuse Decision

## Change type
frontend refactor / feature-owned AWD dialog availability and hint owner split

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- docs/reviews/frontend/2026-05-29-awd-round-dialog-and-contract-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `useAwdOperationsPanelViewState.ts` 已经承担了 AWD operations 的 tab / contest selection / runtime content 这类 panel-level availability 与展示态 owner。
- 当前 `useAwdOperationsDialogState.ts` 继续承担 dialog open/close、next round number、save success close，同时还夹带 `canRecord*` 和 `*Hint`，已经形成第二组独立派生职责。

## Decision
refactor_existing

## Reason
当前 `useAwdOperationsDialogState.ts` 同时承担：

- dialog open/close state
- next round number
- save success -> close dialog workflow
- service/attack availability 与 hint

其中最后一组已经是独立的 availability presentation owner。最小正确改动不是继续扩张 `useAwdOperationsDialogState.ts`，而是：

- 新增 `useAwdOperationsDialogAvailability.ts`
- 把 `canRecordServiceChecks`、`canRecordAttackLogs`、`serviceCheckHint`、`attackLogHint` 下沉到该 composable
- 让 `AWDOperationsPanel.vue` 从 availability owner 和 state owner 两处组合结果
- 更新相关测试与 raw-source 护栏，明确 dialog state owner 不再自己内联 hint derivation

## Files to modify
- .harness/reuse-decisions/awd-dialog-availability-owner-split.md
- docs/plan/impl-plan/2026-05-29-awd-dialog-availability-owner-split-plan.md
- docs/reviews/frontend/2026-05-29-awd-dialog-availability-owner-split-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogAvailability.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts

## After implementation
- `useAwdOperationsDialogState.ts` 只继续负责 dialog local state、next round number、save success close 与 override close guard。
- `useAwdOperationsDialogAvailability.ts` 成为“能不能录入 / 为什么不能录入”的唯一 owner。
- 后续如果 availability 规则继续增长，会长在新 composable，而不会再把 dialog state owner 拉回到 owner-mixed 状态。
