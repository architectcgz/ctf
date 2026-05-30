> 状态：Current
> 事实源：`AWDChallengeConfigPanel.vue` 当前 owner、现有 AWD 配置目录测试与 contest edit feature debt backlog
> 替代：无

# AWD Challenge Config Panel Decomposition Plan

## 目标

- 把 `AWDChallengeConfigPanel.vue` 从“feature owner 正确但父层仍承接稳定展示块 + row + CSS”的状态继续收口成唯一 presentation owner。
- 为 `platform-contests` 补齐 AWD 配置目录的 header、directory section、directory row 和 shared CSS。
- 让现有 raw-source / theme token / surface alignment 护栏继续在聚合源码视角下工作。

## 非目标

- 本轮不改 `ContestEditWorkspacePanel.vue` 的 stage owner。
- 本轮不改 `useAwdCheckResultPresentation()` 的 model owner 或 AWD 校验结果文案逻辑。
- 本轮不改 AWD 配置编辑对话框、preflight stage 或表格列定义。

## 输入依据

- `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue`
- `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `docs/reviews/frontend/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这块 surface 已经处在正确 feature owner 下，当前技术债不在目录归属，而在父组件继续混放稳定 header / summary / table row / CSS。
- 这里的行为 owner 仍应留在 `AWDChallengeConfigPanel.vue`，因为排序、summary 统计和 AWD 校验结果 presentation helper 都不应分散到 row 子组件。
- 最小正确边界是：父层保留排序与 presentation wiring，子组件承接稳定视图，样式独立成 CSS 文件。

## 设计边界

### `AWDChallengeConfigPanel.vue` 本轮继续负责

- `challengeLinks` 输入 contract
- `edit` emit
- 排序与 summary 统计
- `useAwdCheckResultPresentation()` 接线
- directory row view model 组装

### `AWDChallengeConfigHeader.vue` 本轮负责

- overline / title / description
- summary strip metric cards

### `AWDChallengeConfigDirectorySection.vue` 本轮负责

- empty state
- table shell
- row 列表组合

### `AWDChallengeConfigDirectoryRow.vue` 本轮负责

- 单条 challenge row 的 identity、checker、score、config summary、validation、actions 视图

### `awdChallengeConfigPanel.css` 本轮负责

- panel shell、header、summary strip、table、row、validation 和 action 区样式

## 任务切片

### Slice 1：提取 header / directory section / row

- 目标：
  - 新增 `AWDChallengeConfigHeader.vue`
  - 新增 `AWDChallengeConfigDirectorySection.vue`
  - 新增 `AWDChallengeConfigDirectoryRow.vue`
  - 父面板只保留排序、presentation helper 和 row view model
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts src/components/platform/__tests__/AWDChallengeConfigPanelExtraction.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 presentation owner
  - 子组件是否只消费 props / emits，没有继续接入 feature model

### Slice 2：样式与护栏收口

- 目标：
  - 新增 `awdChallengeConfigPanel.css`
  - 更新 theme token / primitive / surface alignment 护栏，切到聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- Review focus：
  - 样式是否已完全脱离父 SFC
  - raw-source 护栏是否仍能覆盖 metric panel、workspace overline、row action primitive 和 dark surface 对齐

### Slice 3：backlog / review 收口

- 目标：
  - 更新 backlog 中 contest / AWD feature 内残余大 surface 的进展
  - 补本轮 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的 `AWDChallengeConfigPanel.vue` 大组件债是否已真正收口
  - 没有把同一批 debt 只换文件名继续留下

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts src/components/platform/__tests__/AWDChallengeConfigPanelExtraction.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDChallengeConfigPanel.vue` 即使完成这轮拆分，仍会继续保留排序和 row view model 组装脚本；如果后续在同一 surface 上继续堆更多筛选或动作，再考虑把目录 presentation model 继续切成局部 composable。
- 这轮不调整 `ContestEditWorkspacePanel.vue` 的 stage 组合关系，因此如果上层再堆新文案或 stage 逻辑，不应把它们回灌到 `AWDChallengeConfigPanel.vue`。
- review 仍很可能只能做到同上下文 self-review，独立 reviewer gate 需要在交付说明里继续显式说明。
