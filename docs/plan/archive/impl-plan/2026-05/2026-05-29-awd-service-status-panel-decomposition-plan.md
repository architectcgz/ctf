> 状态：Current
> 事实源：`AWDServiceStatusPanel.vue` 当前 owner、AWD inspector extraction / theme token 测试、frontend tech debt backlog
> 替代：无

# AWD Service Status Panel Decomposition Plan

## 目标

- 把 `AWDServiceStatusPanel.vue` 从“feature owner 正确但父层仍混放 toolbar / matrix / summary table / CSS”的状态继续收口成唯一 presentation owner。
- 为 `awd-inspector` 补齐服务状态 toolbar、matrix、round performance summary 和共享 CSS。
- 补一份服务状态 extraction 护栏，避免后续拆分继续只靠 `AWDRoundInspectorExtraction.test.ts` 间接覆盖。

## 非目标

- 本轮不改 `AWDRoundInspector.vue`、`awdInspector.types.ts` 的外部 props / emits contract。
- 本轮不改服务状态筛选语义、导出行为或 `getServiceCheckPresentationResult()` / `getCheckStatusLabel()` 这类 formatter 逻辑。
- 本轮不处理 `AWDTrafficPanel.vue`、`AWDAttackLogPanel.vue` 或 `contest-awd-admin` dialog cluster。

## 输入依据

- `code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.vue`
- `code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/reviews/frontend/2026-05-28-awd-round-inspector-decomposition-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这块 surface 已经处在正确 feature owner 下，当前技术债不在目录归属，而在父组件继续混放 toolbar、matrix row、round performance summary 和 CSS。
- 服务状态相关 presentation model 仍应留在 `AWDServiceStatusPanel.vue`，因为挑战列、队伍聚合、checker / source / status / checked-at 展示推导不适合分散到子组件里重复做。
- 最小正确边界是：父层保留 props / emits contract、filter forwarding 和 presentation model；子组件承接稳定视图，样式独立成 CSS 文件。

## 设计边界

### `AWDServiceStatusPanel.vue` 本轮继续负责

- `AWDServiceStatusPanelProps` / `AWDServiceStatusPanelEmits` contract
- 服务矩阵和 round performance 相关 presentation model
- filter change forwarding 与状态标签 / checker / source / checked-at 的格式化拼装

### `AWDServiceStatusToolbar.vue` 本轮负责

- 标题与队伍统计
- team / status / source / alert filters
- export action

### `AWDServiceStatusMatrix.vue` 本轮负责

- matrix table shell
- 队伍行和状态单元格展示
- 空态

### `AWDServiceRoundPerformanceTable.vue` 本轮负责

- round performance summary 标题与 table

### `awdServiceStatusPanel.css` 本轮负责

- panel shell、toolbar、filters、matrix、status card、summary table 和响应式样式

## 任务切片

### Slice 1：提取 toolbar / matrix / summary table

- 目标：
  - 新增 `AWDServiceStatusToolbar.vue`
  - 新增 `AWDServiceStatusMatrix.vue`
  - 新增 `AWDServiceRoundPerformanceTable.vue`
  - 父面板只保留服务状态 presentation model 和 outward emits
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDServiceStatusPanel.test.ts src/components/platform/__tests__/AWDServiceStatusPanelExtraction.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 presentation owner
  - 子组件是否只消费 props / emits，没有接入 inspector model 或 route owner

### Slice 2：样式和 theme/extraction 护栏收口

- 目标：
  - 新增 `awdServiceStatusPanel.css`
  - 更新 extraction / theme token 护栏到聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - panel 是否不再继续保留 scoped style 尾巴
  - 聚合源码护栏是否继续覆盖 toolbar filter / matrix status / summary table surface

### Slice 3：backlog / review 收口

- 目标：
  - 更新 backlog 中当前 P2 residual 的优先目标说明
  - 补本轮 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的服务状态大组件债是否真的收口
  - 没有把 toolbar / matrix / summary / CSS 只是换文件继续留下

## 实施记录

- [x] Slice 1：已新增 `AWDServiceStatusToolbar.vue`、`AWDServiceStatusMatrix.vue`、`AWDServiceRoundPerformanceTable.vue` 与 `awdServiceStatusPanel.types.ts`，父面板收口为唯一服务状态 presentation owner。
- [x] Slice 2：已新增 `awdServiceStatusPanel.css`，并补齐 `AWDServiceStatusPanelExtraction.test.ts`，同时把 `AWDRoundInspectorExtraction.test.ts` 和 `sharedThemeTokenAdoption.test.ts` 切到聚合源码视角。
- [x] Slice 3：已更新 backlog 与本轮 review 文档，记录父面板从约 `509` 行降到 `193` 行，并明确后续 residual 不再停留在服务状态 toolbar / matrix / summary 壳体本身。

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDServiceStatusPanel.test.ts src/components/platform/__tests__/AWDServiceStatusPanelExtraction.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDServiceStatusPanel.vue` 即使完成这轮拆分，仍会继续保留挑战列、队伍聚合、状态展示和 summary 的 view model 组装；如果后续再叠加 drilldown、排序或更复杂交互，再考虑把 presentation model 继续切成局部 composable。
- 这轮不改 `awdInspector.types.ts` 的 contract；如果上层以后要重塑服务状态 props 结构，不应把这种变化和本轮展示壳拆分混在一起。
- review 仍很可能只能做到同上下文 self-review，独立 reviewer gate 需要在交付说明里继续显式说明。
