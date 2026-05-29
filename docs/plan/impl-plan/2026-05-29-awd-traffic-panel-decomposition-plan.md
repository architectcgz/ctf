> 状态：Current
> 事实源：`AWDTrafficPanel.vue` 当前 owner、AWD inspector extraction / primitive / theme / alignment 测试、frontend tech debt backlog
> 替代：无

# AWD Traffic Panel Decomposition Plan

## 目标

- 把 `AWDTrafficPanel.vue` 从“feature owner 正确但父层仍混放 summary / intelligence / drill-down table / CSS”的状态继续收口成唯一 traffic presentation owner。
- 为 `awd-inspector` 补齐流量 summary band、intelligence grid、event drill-down 和共享 CSS。
- 补一份流量态势 extraction 护栏，避免后续拆分继续只靠 `AWDRoundInspectorExtraction.test.ts` 间接覆盖。

## 非目标

- 本轮不改 `useAwdTrafficPanel()` 的 filter / keyword / page owner。
- 本轮不改 `AWDRoundInspector.vue`、`awdInspector.types.ts` 的外部 props / emits contract。
- 本轮不处理 `AWDAttackLogDialog.vue`、`AWDServiceCheckDialog.vue` 或 `AWDAttackLogPanel.vue`。

## 输入依据

- `code/frontend/src/features/awd-inspector/ui/AWDTrafficPanel.vue`
- `code/frontend/src/features/awd-inspector/model/useAwdTrafficPanel.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这块 surface 已经处在正确 feature owner 下，当前技术债不在目录归属，而在父组件继续混放 summary band、intelligence grid、event drill-down 和 CSS。
- traffic 相关交互 owner 仍应留在 `AWDTrafficPanel.vue`，因为 `useAwdTrafficPanel()` 已经持有 keyword / status group / page 的本地交互语义，不适合再分散到子组件里各自实现。
- 最小正确边界是：父层保留 props / emits contract、`useAwdTrafficPanel()` wiring、service option 组装与 label forwarding；子组件承接稳定视图，样式独立成 CSS 文件。

## 设计边界

### `AWDTrafficPanel.vue` 本轮继续负责

- `AWDTrafficPanelProps` / `AWDTrafficPanelEmits` contract
- `useAwdTrafficPanel()` wiring
- service options、status group label forwarding 和 keyword input local owner

### `AWDTrafficSummaryBand.vue` 本轮负责

- metric summary band

### `AWDTrafficIntelligenceGrid.vue` 本轮负责

- 热点实体分析
- 12-bucket trend

### `AWDTrafficEventTable.vue` 本轮负责

- filters
- event table
- pagination

### `awdTrafficPanel.css` 本轮负责

- panel shell、summary band、intelligence grid、trend、drill-down table 和响应式样式

## 任务切片

### Slice 1：提取 summary / intelligence / event table

- 目标：
  - 新增 `AWDTrafficSummaryBand.vue`
  - 新增 `AWDTrafficIntelligenceGrid.vue`
  - 新增 `AWDTrafficEventTable.vue`
  - 父面板只保留 traffic workflow / presentation owner
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDTrafficPanelExtraction.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 `useAwdTrafficPanel()` owner
  - 子组件是否只消费 props / emits，没有接入 inspector model 或 route owner

### Slice 2：样式和 primitive/theme/alignment 护栏收口

- 目标：
  - 新增 `awdTrafficPanel.css`
  - 更新 extraction / primitive / theme / alignment 护栏到聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - panel 是否不再继续保留 scoped style 尾巴
  - 聚合源码护栏是否继续覆盖 summary card、trend、filter 和 pagination surface

### Slice 3：backlog / review 收口

- 目标：
  - 更新 backlog 中当前 P2 residual 的优先目标说明
  - 补本轮 review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的流量态势大组件债是否真的收口
  - 没有把 summary / intelligence / event table / CSS 只是换文件继续留下

## 实施记录

- [x] Slice 1：已新增 `AWDTrafficSummaryBand.vue`、`AWDTrafficIntelligenceGrid.vue`、`AWDTrafficEventTable.vue` 与 `awdTrafficPanel.types.ts`，父面板收口为唯一 traffic workflow / presentation owner。
- [x] Slice 2：已新增 `awdTrafficPanel.css`，并补齐 `AWDTrafficPanelExtraction.test.ts`，同时把 `AWDRoundInspectorExtraction.test.ts`、`contestUiPrimitiveAdoptionPhase4.test.ts`、`platformManagementSurfaceAlignment.test.ts` 与 `sharedThemeTokenAdoption.test.ts` 切到聚合源码视角。
- [x] Slice 3：已更新 backlog 与本轮 review 文档，记录父面板从约 `450` 行降到 `112` 行，并明确后续 residual 不再停留在 traffic summary / intelligence / event table 壳体本身。

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDTrafficPanelExtraction.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDTrafficPanel.vue` 即使完成这轮拆分，仍会继续保留 `useAwdTrafficPanel()`、service option 组装和 status label forwarding；如果后续再叠加导出、drilldown 或更复杂趋势交互，再考虑把 presentation model 继续切成局部 composable。
- 这轮不改 `useAwdTrafficPanel()` 的逻辑 owner；如果后续发现 keyword / status group / page 语义本身要重塑，不应把 تلك个改动和本轮展示壳拆分混在一起。
- review 仍很可能只能做到同上下文 self-review，独立 reviewer gate 需要在交付说明里继续显式说明。
