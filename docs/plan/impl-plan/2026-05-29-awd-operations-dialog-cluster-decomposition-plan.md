> 状态：Current
> 事实源：`AWDAttackLogDialog.vue`、`AWDServiceCheckDialog.vue` 当前 owner、AWD operations extraction / primitive / duplicate-action / backoffice dialog 护栏、frontend tech debt backlog
> 替代：无

# AWD Operations Dialog Cluster Decomposition Plan

## 目标

- 把 `AWDRoundCreateDialog.vue`、`AWDAttackLogDialog.vue` 与 `AWDServiceCheckDialog.vue` 从“父层同时混放 form owner / stable section / footer / scoped CSS”的状态继续收口成唯一 dialog workflow owner。
- 为 `contest-awd-admin` 补齐创建轮次、攻击日志与服务检查这组三个 dialog 的稳定 section、共享 footer、共享 CSS 和共享 payload contract。
- 补一份 dialog cluster extraction 护栏，避免后续继续只靠分散的 primitive / duplicate-action / backoffice dialog 测试间接覆盖。

## 非目标

- 本轮不改 `AWDOperationsDialogHub.vue` 的上层 open / save wiring。
- 本轮不继续下沉 `useAwdOperationsDialogState.ts` 的 workflow owner。
- 本轮不重塑 challenge link / team data 的上游 contract。

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这块 surface 已经处在正确 feature owner 下，当前技术债不在目录归属，而在三个父 dialog 继续同时承接稳定 field section、footer 与 scoped CSS。
- 交互 owner 仍应留在三个父 dialog 内，因为 open watch、field error reset、service id 解析、JSON parse 与 duplicate-action guard 都是单个提交流程的局部 owner，不适合拆散到子组件里各自维护。
- 最小正确边界是：父层保留 props / emits contract、form state、validation、submit；子组件承接稳定视图；共享 payload type 收口到 feature contract；样式独立成 feature CSS 文件。

### `AWDRoundCreateDialog.vue` 本轮继续负责

- `open / nextRoundNumber / saving` props contract
- `update:open`、`save` emits contract
- open watch 下的 form reset
- field error owner
- `validate()`、`handleSubmit()` 与 duplicate-action guard

### `AWDRoundCreateSettingsSection.vue` 本轮负责

- 轮次编号
- 初始状态

### `AWDRoundCreateScoreSection.vue` 本轮负责

- 攻击分
- 防守分

## 设计边界

### `AWDAttackLogDialog.vue` 本轮继续负责

- `open / teams / challengeLinks / saving` props contract
- `update:open`、`save` emits contract
- open watch 下的 form reset
- field error owner
- `getSelectedServiceId()`、`handleSubmit()` 与 duplicate-action guard

### `AWDAttackLogTargetSection.vue` 本轮负责

- 攻击队伍 / 受害队伍 / 题目选择区

### `AWDAttackLogDetailsSection.vue` 本轮负责

- 攻击类型
- 提交 flag
- 成功复选框
- hint / warning 文案

### `AWDServiceCheckDialog.vue` 本轮继续负责

- `open / teams / challengeLinks / saving` props contract
- `update:open`、`save` emits contract
- open watch 下的 form reset
- field error owner
- `getSelectedServiceId()`、`parseCheckResult()`、`handleSubmit()` 与 duplicate-action guard

### `AWDServiceCheckTargetSection.vue` 本轮负责

- 队伍 / 题目 / 服务状态选择区

### `AWDServiceCheckResultSection.vue` 本轮负责

- JSON textarea
- error / warning hint

### `AWDOperationsDialogFooter.vue` 本轮负责

- 取消 / 提交按钮壳
- loading label、disabled 态与响应式 footer 布局

### `awdOperationsDialogContracts.ts` 本轮负责

- create round payload
- create service check payload
- create attack log payload

### `awdOperationsDialogs.css` 本轮负责

- round / attack / service 三组 field gap
- checkbox、textarea、warning、footer 响应式样式

## 任务切片

### Slice 1：提取 round / attack / service 三组 dialog 的稳定 section

- 目标：
  - 新增 `AWDRoundCreateSettingsSection.vue`
  - 新增 `AWDRoundCreateScoreSection.vue`
  - 新增 `AWDAttackLogTargetSection.vue`
  - 新增 `AWDAttackLogDetailsSection.vue`
  - 新增 `AWDServiceCheckTargetSection.vue`
  - 新增 `AWDServiceCheckResultSection.vue`
  - 三个父 dialog 只保留 workflow / validation owner
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts`
- Review focus：
  - 父 dialog 是否仍然是唯一 submit / validation owner
  - 子组件是否只消费 props / emits，没有自行持有提交逻辑

### Slice 2：共享 footer / CSS / contract 与 raw-source 护栏收口

- 目标：
  - 新增 `AWDOperationsDialogFooter.vue`
  - 新增 `awdOperationsDialogContracts.ts`
  - 新增 `awdOperationsDialogs.css`
  - 更新 primitive / duplicate-action / backoffice dialog 护栏到聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- Review focus：
  - dialog 是否不再继续保留 scoped style 尾巴
  - 聚合源码护栏是否继续覆盖 shared primitive、duplicate-action、payload contract 与 `AdminSurfaceModal` 约束

### Slice 3：backlog / review 收口

- 目标：
  - 更新 backlog 中当前 P1 residual 的优先目标说明
  - 补本轮 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上这组 dialog 大组件债是否真的收口
  - 没有只是把 section / footer / CSS 平移后仍留在父层

## 实施记录

- [x] Slice 1：已新增 `AWDRoundCreateSettingsSection.vue`、`AWDRoundCreateScoreSection.vue`、`AWDAttackLogTargetSection.vue`、`AWDAttackLogDetailsSection.vue`、`AWDServiceCheckTargetSection.vue`、`AWDServiceCheckResultSection.vue`，三个父 dialog 回到唯一 form / validation / submit owner。
- [x] Slice 2：已新增 `AWDOperationsDialogFooter.vue`、`awdOperationsDialogContracts.ts`、`awdOperationsDialogOptions.ts` 与 `awdOperationsDialogs.css`，并补齐 `AWDOperationsDialogsExtraction.test.ts`，同时把 primitive / duplicate-action / backoffice dialog 护栏切到聚合源码视角。
- [x] Slice 3：已更新 backlog 与本轮 review 文档，记录三组 dialog 从约 `254` / `378` / `326` 行降到 `136` / `175` / `175` 行，并明确 residual 重点已不再停留在这组三个 dialog 的 section / footer / CSS 混写。

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 三个 dialog 即使完成这轮拆分，仍会继续保留 service id 解析、本地 form reset 与 round form validation；如果后续再叠更复杂预填或联动逻辑，再考虑抽局部 composable，而不是现在提前拆散。
- 这轮先收口三组 dialog 自身的 section / footer / payload contract；`AWDOperationsDialogHub.vue` 和 `useAwdOperationsDialogState.ts` 的更深层 workflow owner 清理仍留在下一刀。
- review 很可能仍只能做到同上下文 self-review，独立 reviewer gate 需要在交付说明里继续显式说明。
