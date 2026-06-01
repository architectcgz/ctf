> 状态：Current
> 事实源：`AWDOperationsPanel.vue` 当前 feature owner、`usePlatformContestAwd()`、现有 AWD operations 护栏测试
> 替代：无

# AWD Operations Panel Owner Cleanup Plan

## 目标

- 把 `AWDOperationsPanel.vue` 从“stage 已拆、但本地 owner 仍堆在父层”的状态继续收口成唯一 feature workflow 组合层。
- 为 `contest-awd-admin` 补齐 panel 专属的 view-state / dialog workflow composable。
- 把父层剩余 shell 样式迁到独立 CSS，并让护栏转向组合源码视角。

## 非目标

- 本轮不改 `usePlatformContestAwd()` 的 API、生命周期、readiness decision、round operations 或实例编排 owner。
- 本轮不改 `AWDOperationsPreRuntimeStage.vue`、`AWDOperationsRuntimeStage.vue`、`AWDOperationsDialogHub.vue` 的展示结构。
- 本轮不继续下钻 `AWDInstanceOrchestrationPanel.vue` 或 `AWDRoundInspector.vue` 的内部结构。

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `docs/plan/impl-plan/2026-05-28-awd-operations-panel-decomposition-plan.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDOperationsPanel.vue` 已经不是模板堆叠问题，当前主要债务是本地 owner 还没有从父组件抽离。
- 这些 owner 都是 panel 私有交互，不适合放进 `usePlatformContestAwd()`，否则会把 feature model 和展示层互动重新耦合。
- 因此本轮的最小正确边界是：保留 `usePlatformContestAwd()` 作为业务 workflow owner，在 panel UI 层新增专属 composable 承接本地 view-state 和 dialog workflow。

## 设计边界

### `AWDOperationsPanel.vue` 本轮继续负责

- `usePlatformContestAwd()` 的唯一接线
- 对外 props / emits contract
- stage / dialog hub 组合
- 对上层 `ContestOperations.vue` 的 feature public UI owner

### `useAwdOperationsPanelViewState.ts` 本轮负责

- `selectedContest`
- `runtimeStageReady`
- `activePanel` / `visibleOperationTabs`
- `runtimeContent` 与 stage visibility 派生
- `useTabKeyboardNavigation()` 相关 tab owner

### `useAwdOperationsDialogState.ts` 本轮负责

- round / service check / attack log 的 open-close
- 保存成功后关闭 dialog
- `nextRoundNumber`
- service / attack capability 与 hint 推导
- override dialog close guard

### `awdOperationsPanel.css` 本轮负责

- `studio-ops-shell`
- `studio-ops-content`

## 任务切片

### Slice 1：收口 panel view-state

- 目标：
  - 新增 `useAwdOperationsPanelViewState.ts`
  - 父组件改为消费该 composable，而不是继续内联 contest / tab / runtime 相关 owner
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- Review focus：
  - 是否把本地 view-state 收口到 panel UI owner，而不是倒灌进 feature model
  - controlled `operationPanel` 模式与 tab keyboard navigation 是否保持不变

### Slice 2：收口 dialog workflow 与父层样式

- 目标：
  - 新增 `useAwdOperationsDialogState.ts`
  - 新增 `awdOperationsPanel.css`
  - 父组件只保留 stage / dialog composition，不再混放 dialog cluster 与 capability hint 计算
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- Review focus：
  - dialog 只能在运行态被打开的约束是否仍成立
  - 保存后关闭、override dialog close、service/attack hint 是否保持原行为

### Slice 3：护栏与 backlog/review 收口

- 目标：
  - 更新 extraction 护栏，让其覆盖新 composable / CSS owner
  - 更新 backlog 和本轮 review 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - 父组件是否已经回到唯一 workflow composition owner
  - touched surface 上的 residual owner debt 是否真的收口，而不是换文件继续堆着

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 本轮只清 `AWDOperationsPanel.vue` 本地 owner，不涉及 `AWDInstanceOrchestrationPanel.vue` 内部结构；如果实例编排面继续膨胀，需要单独开新切片。
- 这轮新增的是 panel 专属 composable，不会成为跨 feature 共享模型；如果后续被多个 panel 复用，再考虑上提。
- review 仍大概率只能做到同上下文 self-review，缺少独立 reviewer 上下文的复核证据，这条缺口需要在交付说明里明确。
