> 状态：Current
> 事实源：`AWDOperationsPanel.vue` 当前 runtime owner、`usePlatformContestAwd()`、现有 AWD operations 护栏测试
> 替代：无

# AWD Operations Panel Decomposition Plan

## 目标

- 把 `AWDOperationsPanel.vue` 从“workflow owner + pre-runtime 壳 + runtime 壳 + 样式”收口成明确的 operations owner
- 在 `features/contest-awd-admin/ui` 内补齐 tabs / pre-runtime stage / runtime stage 三个稳定视图区块
- 把 live region + runtime dialog cluster 收口到独立 dialog hub，避免父面板停在半拆分状态
- 保持对外 props / emits contract 不变，让 `ContestOperations.vue` 和现有挂载测试继续按原契约工作

## 非目标

- 本轮不改 `usePlatformContestAwd()` 的加载、创建、巡检、实例启动、override gate owner
- 本轮不改 `AWDRoundInspector.vue`、`AWDReadinessSummary.vue`、`AWDInstanceOrchestrationPanel.vue` 的内部结构
- 本轮不顺手处理 `ContestProjectorAttackMap.vue`

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/features/contest-awd-admin/index.ts`
- `code/frontend/src/features/contest-awd-admin/ui/AWDContestSelectorField.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePendingState.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDOperationsPanel.vue` 的真实行为 owner 在 `usePlatformContestAwd()`、顶层 contest 选择 guard、tab 状态和 dialog open state
- 模板层最明显的问题是 pre-runtime 与 runtime 两个阶段的组合壳依然混在父文件里，且 tab 视图重复出现两次
- 阶段壳体拆分后，父文件剩余最明显的模板噪音是底部 live region + 4 个 dialog cluster；这里也不应继续留在 parent owner
- 这类壳体拆分不会改变 runtime workflow，只需要保持 props / emits 桥接清楚

## 设计边界

### `AWDOperationsPanel.vue` 本轮继续负责

- `selectedContest` / `runtimeStageReady` / `activePanel` / `visibleOperationTabs`
- `usePlatformContestAwd()` model wiring
- `useTabKeyboardNavigation()` owner
- dialog open state、`createRound/createServiceCheck/createAttackLog` 桥接
- `open:awd-config` / `open:contest-edit` / `update:selectedContestId` 发射

### `AWDOperationsTabs.vue` 本轮负责

- tabs 导航按钮视图
- button ref 和键盘事件回传到父层 owner

### `AWDOperationsPreRuntimeStage.vue` 本轮负责

- 未开赛阶段 header
- pending state / readiness summary / instance orchestration 的稳定组合壳

### `AWDOperationsRuntimeStage.vue` 本轮负责

- 运行态 readiness strip
- round inspector / instance orchestration 的稳定组合壳

### `AWDOperationsDialogHub.vue` 本轮负责

- live region
- round / service check / attack log / readiness override 四个 dialog 的展示桥接
- dialog open / save / confirm 事件回传父层 owner

### 本轮不动

- runtime workflow / API owner
- AWDRoundInspector 的内部 tab、traffic、scoreboard、检查动作
- readiness capability 的 decision / override 内部实现

## 任务切片

### Slice 1：提取 tabs、stage shell 与 dialog hub

- 目标：
  - 新增 `AWDOperationsTabs.vue`
  - 新增 `AWDOperationsPreRuntimeStage.vue`
  - 新增 `AWDOperationsRuntimeStage.vue`
  - 新增 `AWDOperationsDialogHub.vue`
  - 父面板改为只组合 contest selector / empty guard / 两个 stage / dialog owner bridge
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 workflow owner / tab owner / dialog owner
  - 子组件是否只消费 props 和发射展示层事件，没有偷偷接入 composable 或 API

### Slice 2：更新 extraction / primitive 护栏

- 目标：
  - 调整 `awdOperationsPanelTabsExtraction.test.ts` 与 `contestUiPrimitiveAdoptionPhase4.test.ts`
  - 让 raw-source 检查适配拆分后的聚合源码
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- Review focus：
  - tab owner、selector/pending owner 与 UI primitive 约束是否仍被有效覆盖
  - 没有因为子组件拆分让原来的护栏失效

### Slice 3：backlog 与 review 收口

- 目标：
  - 更新 backlog 里 `AWDOperationsPanel.vue` 的状态
  - 补 frontend review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的超大 operations panel 债是否真的收口
  - 没有把 parent owner 转移成新的 feature 内大组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 如果 `AWDInstanceOrchestrationPanel.vue` 后续仍继续膨胀，需要单独按实例编排内部 owner 再开一刀；这不属于本轮 `AWDOperationsPanel` 壳体收口范围。
- 运行态与未开赛阶段当前仍共享部分视觉 class 命名；本轮先以 owner 清晰为主，不顺手做样式系统重命名。
- `AWDOperationsPanel.vue` 经过 stage shell + dialog hub 收口后仍约 `500` 行，剩余主要是 `usePlatformContestAwd()` wiring 与 handler owner；如果后续继续增长，应改按 workflow handler / inspector action owner 再开单独切片，而不是把展示壳体再塞回父文件。
