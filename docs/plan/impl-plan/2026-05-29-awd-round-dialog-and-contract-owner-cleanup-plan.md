> 状态：Current
> 事实源：`AWDRoundCreateDialog.vue`、`AWDOperationsDialogHub.vue`、`useAwdOperationsDialogState.ts` 当前 owner，AWD operations dialog cluster 护栏与 frontend tech debt backlog
> 替代：无

# AWD Round Dialog And Contract Owner Cleanup Plan

## 目标

- 为 `contest-awd-admin` 的 operations dialog cluster 补齐共享 payload / override state contract owner，减少 hub / state / dialog 间重复内联类型。
- 在不扩大到整个 AWD operations feature 重构的前提下，收口 “save success -> close dialog” 这组 shared workflow 语义。

## 非目标

- 本轮不再继续拆 `useAwdOperationsDialogState.ts` 成多个 composable。
- 本轮不改 `AWDOperationsPanel.vue`、`AWDOperationsPreRuntimeStage.vue`、`AWDOperationsRuntimeStage.vue` 的结构边界。
- 本轮不改 `usePlatformContestAwd()` 的远端 mutation contract。

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDRoundCreateDialog.vue` 与 attack/service 两个 dialog 的表单壳体已经在上一刀收口；当前剩余真实 residual，不在“再拆一个子文件”本身，而在 AWD operations dialog cluster 仍没有明确的共享 payload / override state contract owner，导致同一套 shape 在 hub、workflow state 和 tabs 护栏里重复书写。
- 最小正确边界是：三组 dialog 继续保留各自 props / emits、form、validation、submit；dialog cluster 的共享 contract 与 close-after-save 语义收口到 `dialog hub + dialog state` 这一层的单点 owner。

## 设计边界

### `awdOperationsDialogContracts.ts` 本轮负责

- create round payload
- create service check payload
- create attack log payload
- override dialog state shape

### `useAwdOperationsDialogState.ts` 本轮继续负责

- dialog open state
- next round number
- close-after-save workflow owner
- override close guard

### `AWDOperationsDialogHub.vue` 本轮继续负责

- dialog cluster composition surface
- 通过共享 contract 转发 `props / emits`

## 任务切片

### Slice 1：收口 dialog cluster 的共享 contract 与 close-after-save 语义

- 目标：
  - 更新 `AWDOperationsDialogHub.vue`、`useAwdOperationsDialogState.ts` 使用共享 payload / override state contract
  - 用局部 helper 收口 save success 后关闭对应 dialog 的 shared workflow
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- Review focus：
  - payload / override state 是否已存在单点 type owner
  - workflow owner 是否仍集中在 `useAwdOperationsDialogState.ts`，没有分散回 panel 或 hub

### Slice 2：backlog / review 收口

- 目标：
  - 更新 backlog 中当前 residual 的优先目标说明
  - 补本轮 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上 round dialog 与 deeper contract owner 债是否一起收口
  - 没有把 contract owner 清理只做成类型平移而没有减少重复

## 实施记录

- [x] Slice 1：已让 `AWDOperationsDialogHub.vue`、`useAwdOperationsDialogState.ts` 统一转到共享 dialog payload / override state contract；`useAwdOperationsDialogState.ts` 同时已用局部 helper 收口 save success 后关闭对应 dialog 的 shared workflow 语义。
- [x] Slice 2：已更新 backlog 与本轮 review 文档，记录 `AWDOperationsDialogHub.vue` 从约 `113` 行降到 `87` 行，`useAwdOperationsDialogState.ts` 从本轮实现前的 `178` 行降到 `150` 行，并明确 residual 重点已不再停留在 dialog cluster 的重复 payload contract。

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useAwdOperationsDialogState.ts` 即使完成这轮 contract owner 收口，仍会继续保留 open state；如果后续 availability / hint workflow 继续增长，再考虑是否把 availability presentation 和 mutation workflow 分成两个 composable。
- 这轮不改 `AWDOperationsPanel.vue` 上层 wiring；如果后续 runtime stage 和 dialog cluster 间再出现跨 surface 的 owner 漂移，不应把它和当前 round dialog/contract 收口混在一起。
- review 很可能仍只能做到同上下文 self-review，独立 reviewer gate 需要在交付说明里继续显式说明。
