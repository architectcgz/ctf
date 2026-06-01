> 状态：Current
> 事实源：`contest-awd-admin` runtime rule 调用链、AWD panel / dialog / round operation 测试
> 替代：无

# AWD Runtime Policy Owner Cleanup Plan

## 目标

- 把 contest AWD 的运行态规则收成单点 model owner
- 让 `runtimeStageReady`、`canOperateSelectedRound`、`shouldUseCurrentRoundCheck`、`shouldAutoRefresh` 不再分散在 UI / dialog / round mutation 各自推导
- 保持 AWD operations 的现有外部行为不变

## 非目标

- 本轮不调整 `useAwdReadinessDecision.ts` 的强制放行规则
- 本轮不调整 `useAwdOperationsDialogAvailability.ts` 的提示文案 owner
- 本轮不变更 inspector header 的按钮布局或交互文案

## 输入依据

- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdContestStateFlags.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.ts`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts`
- `code/frontend/src/features/awd-inspector/ui/AWDRoundHeaderPanel.vue`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`

## 当前结论

- `runtimeStageReady` 现在由 `useAwdOperationsPanelViewState.ts` 在 UI layer 推导，但 `useAwdOperationsDialogState.ts` 与 `useAwdRoundOperations.ts` 又都依赖同一条运行态语义。
- `useAwdContestStateFlags.ts` 已经是最接近 runtime policy owner 的现有入口，但目前只承接 `hasSelectedContest` 与 `shouldAutoRefresh`。
- `useAwdRoundOperations.ts` 本地还在重复推导 `canOperateRound` 与 `shouldRunCurrentRound`，说明 runtime rule owner 还没有完全收口。

## 设计边界

### `useAwdContestStateFlags.ts` 本轮负责

- `hasSelectedContest`
- `runtimeStageReady`
- `canOperateSelectedRound`
- `shouldUseCurrentRoundCheck`
- `shouldAutoRefresh`

### `useAwdRoundOperations.ts` 本轮继续负责

- round check / create round / service check / attack log mutation
- readiness blocked 时委托 override dialog
- toast 与 refresh side effect

但它不再负责定义 runtime gate 规则。

### `useAwdOperationsPanelViewState.ts` 本轮继续负责

- selected panel / tabs / runtime content 视图壳层
- keyboard navigation
- pre-runtime / runtime stage 可见区域推导

但它不再负责定义 runtime rule。

## 任务切片

### Slice 1：收口 runtime policy model owner

- 更新：
  - `useAwdContestStateFlags.ts`
  - `useAwdRoundOperations.ts`
  - `usePlatformContestAwd.ts`
- 新增：
  - `useAwdContestStateFlags.test.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/model/useAwdContestStateFlags.test.ts src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts`

### Slice 2：让 panel / dialog controller 消费 runtime policy

- 更新：
  - `AWDOperationsPanel.vue`
  - `useAwdOperationsPanelViewState.ts`
  - `useAwdOperationsPanelViewState.test.ts`
  - `useAwdOperationsDialogState.test.ts`
  - `AWDOperationsPanel.test.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts`

### Slice 3：同步 backlog / review 与终态验证

- 更新：
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-30-awd-runtime-policy-owner-cleanup-review.md`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/model/useAwdContestStateFlags.test.ts src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- readiness override 的更深层动作规则仍留在 `useAwdReadinessDecision.ts`，本轮不会继续把它与普通 runtime gate 混到一起。
- availability 提示仍只覆盖队伍/题目充分性；如果未来需要把权限、readiness 或 round 状态也纳入 availability，这会是下一条独立规则 owner 收口线。
