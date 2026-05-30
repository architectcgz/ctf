# Reuse Decision

## Change type
frontend refactor / contest AWD runtime policy owner cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdContestStateFlags.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts
- code/frontend/src/features/awd-inspector/ui/AWDRoundHeaderPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `useAwdOperationsDialogState.ts` 刚刚已经收口成 dialog controller owner，但它仍依赖来自 UI layer 的 `runtimeStageReady`，说明 runtime rule owner 还没有完全下沉。
- `useAwdContestStateFlags.ts` 当前已经承接 `hasSelectedContest` 与 `shouldAutoRefresh`，是最接近 runtime policy owner 的现有入口。
- `useAwdRoundOperations.ts` 当前仍本地重复维护 `canOperateRound` 与 `shouldRunCurrentRound` 判定，适合继续并回 `useAwdContestStateFlags.ts` 这一条 runtime policy 线，而不是再新增分散 helper。

## Decision
refactor_existing

## Reason
当前 runtime 规则分散在三层：

- `useAwdOperationsPanelViewState.ts`：根据 contest status 推导 `runtimeStageReady`
- `useAwdOperationsDialogState.ts`：用 `runtimeStageReady` 决定 dialog 能否打开
- `useAwdRoundOperations.ts`：本地再算 `canOperateRound` 与 `shouldRunCurrentRound`
- `useAwdContestStateFlags.ts`：只承接了 `hasSelectedContest` 与 `shouldAutoRefresh`

这会导致 runtime 规则没有单点 owner。最小正确改动是：

- 扩展 `useAwdContestStateFlags.ts`，让它成为统一 runtime policy owner
- 把 `runtimeStageReady`、`canOperateSelectedRound`、`shouldUseCurrentRoundCheck` 与 `shouldAutoRefresh` 收到这一个 model composable
- `useAwdRoundOperations.ts` 改为消费 runtime policy，而不是继续本地重复推导
- `AWDOperationsPanel.vue` 改为从 model 读取 `runtimeStageReady`，再传给 panel view-state / dialog controller
- `useAwdOperationsPanelViewState.ts` 回到纯 shell/view-state owner，不再继续持有 runtime rule

本轮不调整：

- `useAwdReadinessDecision.ts` 的 override 规则
- `useAwdOperationsDialogAvailability.ts` 的队伍/题目提示语义
- `AWDRoundHeaderPanel.vue` 的按钮文案与布局

## Files to modify
- .harness/reuse-decisions/awd-runtime-policy-owner-cleanup.md
- docs/plan/impl-plan/2026-05-30-awd-runtime-policy-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-30-awd-runtime-policy-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/model/useAwdContestStateFlags.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdContestStateFlags.test.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts

## After implementation
- contest AWD runtime 规则会回到 model 层单点 owner。
- panel shell 只消费 runtime policy，不再自己定义运行态。
- dialog controller 与 round mutation 会共用同一套 runtime policy，而不是各算一套 gate。
