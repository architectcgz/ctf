> 状态：Current
> 事实源：`contest-awd-admin` dialog cluster 当前实现、AWD operations 提取护栏、前端技术债 backlog
> 替代：无

# AWD Operations Dialog Workflow Cleanup Plan

## 目标

- 把 `AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 从“modal shell + form workflow 混写”收口到显式 dialog shell owner
- 把 form draft / reset / validation / payload build 收到 feature 内 composable
- 把 `useAwdOperationsDialogState.ts` 收成 runtime dialog controller owner
- 把 `AWDOperationsDialogHub.vue` / `AWDOperationsPanel.vue` 之间的 dialog contract 改成 grouped binding，避免平铺的大串 props / handlers

## 非目标

- 本轮不修改 `usePlatformContestAwd` 的远端 mutation owner
- 本轮不修改 `AWDReadinessOverrideDialog` 的内部逻辑
- 本轮不继续拆 `AWDOperationsPreRuntimeStage.vue`、`AWDOperationsRuntimeStage.vue`

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts`
- `code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogContracts.ts`
- `code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogOptions.ts`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 3 个 dialog 的稳定 section 和 footer 已经下沉，但 form workflow 还停留在 SFC 本体。
- `AWDOperationsPanel.vue` 已经把 dialog 组合下沉到 `AWDOperationsDialogHub.vue`，但 panel -> hub 的 contract 仍然是展开式 open/saving/update/save 平铺传递。
- `useAwdOperationsDialogState.ts` 已经是唯一 dialog open/close 与 mutation close guard owner，适合继续收口成 controller，而不是再把逻辑拉回 page model。

## 设计边界

### Dialog form composable 本轮负责

- dialog open 时 reset draft
- field error 清理
- validation / parse
- payload build

### Dialog SFC 本轮继续负责

- modal shell
- section 组合
- saving guard
- `update:open` / `save` emit

### `useAwdOperationsDialogState.ts` 本轮负责

- 运行态 gate
- dialog open/close controller
- mutation success 后 close
- override dialog close guard

### `AWDOperationsPanel.vue` 本轮继续负责

- 远端 AWD workflow wiring
- dialog controller 与 dialog hub 组合
- inspector / instances / readiness stage 组合

## 任务切片

### Slice 1：收口三个 dialog 的 form workflow

- 新增：
  - `useAwdRoundCreateDialogForm.ts`
  - `useAwdServiceCheckDialogForm.ts`
  - `useAwdAttackLogDialogForm.ts`
- 更新：
  - `AWDRoundCreateDialog.vue`
  - `AWDServiceCheckDialog.vue`
  - `AWDAttackLogDialog.vue`
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts`

### Slice 2：收口 dialog controller 与 hub contract

- 更新：
  - `useAwdOperationsDialogState.ts`
  - `AWDOperationsDialogHub.vue`
  - `AWDOperationsPanel.vue`
  - `awdOperationsDialogContracts.ts`
  - `useAwdOperationsDialogState.test.ts`
  - `awdOperationsPanelTabsExtraction.test.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts`

### Slice 3：同步 backlog 与终态验证

- 更新：
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - review 文档
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDOperationsPanel` 虽然会减少 dialog wiring 噪音，但更深层的 runtime capability owner 仍主要留在 `usePlatformContestAwd`，本轮不触碰。
- 如果后续 dialog 继续增长，下一刀更适合再抽出 shared dialog form helper，而不是重新把校验拉回各自 SFC。
