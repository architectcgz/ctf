> 状态：Current
> 事实源：`AWDInstanceOrchestrationPanel.vue` 当前 owner、AWD operations 挂载测试、contest / AWD feature debt backlog
> 替代：无

# AWD Instance Orchestration Panel Decomposition Plan

## 目标

- 把 `AWDInstanceOrchestrationPanel.vue` 从“feature owner 正确但父层仍混放 summary / matrix / row / CSS”的状态继续收口成唯一 view model owner。
- 为 `contest-awd-admin` 补齐实例编排 header、matrix、row 和共享 CSS。
- 补一份实例编排 extraction 护栏，避免后续拆分继续依赖 `AWDOperationsPanel.test.ts` 间接覆盖。

## 非目标

- 本轮不改 `usePlatformContestAwd()` 的实例编排 workflow、刷新策略或启动动作 contract。
- 本轮不改 `AWDOperationsPanel.vue`、`AWDReadinessSummary.vue` 或 `AWDRoundInspector.vue`。
- 本轮不新增实例编排筛选、批量操作或新交互。

## 输入依据

- `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `docs/plan/impl-plan/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-plan.md`
- `docs/reviews/frontend/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这块 surface 已经处在正确 feature owner 下，当前技术债不在目录归属，而在父组件继续混放 summary、empty state、matrix row 和 CSS。
- 实例编排 view model 仍应留在 `AWDInstanceOrchestrationPanel.vue`，因为 instance map、visible services、running summary 与 starting key 判定不适合分散到 row 子组件。
- 最小正确边界是：父层保留 view model 和 action emits，子组件承接稳定视图，样式独立成 CSS 文件。

## 设计边界

### `AWDInstanceOrchestrationPanel.vue` 本轮继续负责

- `orchestration` / `loading` / `startingKey` 输入 contract
- `refresh` / `start-cell` / `start-team` / `start-all` emits
- instance map、visible services、running summary、cell/row starting 判定

### `AWDInstanceOrchestrationHeader.vue` 本轮负责

- 标题
- running summary
- refresh / start-all 按钮

### `AWDInstanceOrchestrationMatrix.vue` 本轮负责

- empty state
- table shell
- row 列表组合

### `AWDInstanceOrchestrationRow.vue` 本轮负责

- 单队伍行
- 各服务单元格状态、访问链接、启动按钮
- 行级“启动本队”动作

### `awdInstanceOrchestrationPanel.css` 本轮负责

- panel shell、header、summary、matrix、row、status、action 和响应式样式

## 任务切片

### Slice 1：提取 header / matrix / row

- 目标：
  - 新增 `AWDInstanceOrchestrationHeader.vue`
  - 新增 `AWDInstanceOrchestrationMatrix.vue`
  - 新增 `AWDInstanceOrchestrationRow.vue`
  - 父面板只保留实例编排 view model 和 outward emits
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/AWDInstanceOrchestrationPanelExtraction.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 view model owner
  - 子组件是否只消费 props / emits，没有接入 feature model 或 API

### Slice 2：样式和 primitive 护栏收口

- 目标：
  - 新增 `awdInstanceOrchestrationPanel.css`
  - 更新 primitive adoption 护栏到聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- Review focus：
  - panel 是否不再继续保留 scoped style 尾巴
  - 共享按钮原语断言是否仍覆盖到实例编排 surface

### Slice 3：backlog / review 收口

- 目标：
  - 更新 backlog 中当前 P2 residual 的优先目标说明
  - 补本轮 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的实例编排大组件债是否真的收口
  - 没有把 matrix / row / CSS 只是换文件继续留下

## 实施记录

- [x] Slice 1：已新增 `AWDInstanceOrchestrationHeader.vue`、`AWDInstanceOrchestrationMatrix.vue`、`AWDInstanceOrchestrationRow.vue` 与 `awdInstanceOrchestration.types.ts`，父面板收口为唯一实例编排 view model / emits owner。
- [x] Slice 2：已新增 `awdInstanceOrchestrationPanel.css`，并补齐 `AWDInstanceOrchestrationPanelExtraction.test.ts` 与 `contestUiPrimitiveAdoptionPhase4.test.ts` 的聚合源码护栏。
- [x] Slice 3：已更新 backlog 与本轮 review 文档，记录父面板从约 `464` 行降到 `145` 行，并明确后续 residual 不再停留在实例编排 table 壳体本身。

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/AWDInstanceOrchestrationPanelExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDInstanceOrchestrationPanel.vue` 即使完成这轮拆分，仍会继续保留 instance map、visible services、starting key 判定和 running summary 组装；如果后续再加筛选或更多动作，再考虑把 view model 继续切成局部 composable。
- 这轮不改 `usePlatformContestAwd()` 的实例编排 workflow，因此如果上层要新增更多实例编排行为，不应回灌到 row 子组件里。
- review 仍很可能只能做到同上下文 self-review，独立 reviewer gate 需要在交付说明里继续显式说明。
