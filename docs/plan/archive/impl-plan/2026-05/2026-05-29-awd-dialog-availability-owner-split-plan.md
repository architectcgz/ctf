> 状态：Current
> 事实源：`useAwdOperationsDialogState.ts` 当前 owner、AWD operations panel extraction 护栏、frontend tech debt backlog
> 替代：无

# AWD Dialog Availability Owner Split Plan

## 目标

- 把 `useAwdOperationsDialogState.ts` 里的 availability / hint 派生逻辑拆到独立 owner。
- 明确 AWD operations dialog cluster 内“state workflow owner”和“availability presentation owner”的边界。

## 非目标

- 本轮不改 dialog payload contract。
- 本轮不改 `AWDOperationsDialogHub.vue`。
- 本轮不重塑 `usePlatformContestAwd()` 或 `AWDOperationsPanel.vue` 的上层 workflow。

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `useAwdOperationsDialogState.ts` 当前已经把 save success close 等 workflow 收口得比较清楚，但 `canRecord*` 和 `*Hint` 这组派生规则不属于 dialog local state。
- 这组逻辑未来更可能沿着“规则更多、文案更多、依赖更多状态”的方向增长，因此应提前独立成 availability owner，而不是继续挤在 state owner 里。

## 设计边界

### `useAwdOperationsDialogState.ts` 本轮继续负责

- dialog open/close state
- `nextRoundNumber`
- save success -> close dialog
- override close guard

### `useAwdOperationsDialogAvailability.ts` 本轮负责

- `canRecordServiceChecks`
- `canRecordAttackLogs`
- `serviceCheckHint`
- `attackLogHint`

### `AWDOperationsPanel.vue` 本轮负责

- 组合 state owner 与 availability owner 的结果，继续向 `AWDRoundInspector` 透传

## 任务切片

### Slice 1：提取 availability owner

- 目标：
  - 新增 `useAwdOperationsDialogAvailability.ts`
  - 从 `useAwdOperationsDialogState.ts` 移除 `canRecord*` 与 `*Hint`
  - `AWDOperationsPanel.vue` 改为组合两个 composable
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- Review focus：
  - state owner 是否真的只剩 dialog local workflow
  - availability owner 是否不反向接管 open/close 或 mutation

### Slice 2：backlog / review 收口

- 目标：
  - 更新 backlog 说明
  - 补本轮 review 文档
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - 没有把拆分做成无意义的平移
  - 后续增长面是否已经落到更合适的 owner

## 实施记录

- [x] Slice 1：已新增 `useAwdOperationsDialogAvailability.ts`，并把 `canRecordServiceChecks`、`canRecordAttackLogs`、`serviceCheckHint`、`attackLogHint` 从 `useAwdOperationsDialogState.ts` 移出；`AWDOperationsPanel.vue` 已改为组合 availability owner 与 state owner。
- [x] Slice 2：已更新 `useAwdOperationsDialogState.test.ts`、新增 `useAwdOperationsDialogAvailability.test.ts`，并将 `awdOperationsPanelTabsExtraction.test.ts` 切到新的 source 边界视角；同时已更新 backlog 与本轮 review 文档。

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 如果以后 availability 规则开始依赖 readiness、权限或当前轮状态，本轮新 owner 仍然可能继续变大；但它至少不会再污染 dialog state owner。
- review 仍很可能只能做到同上下文 self-review，独立 reviewer gate 需要继续显式说明。
