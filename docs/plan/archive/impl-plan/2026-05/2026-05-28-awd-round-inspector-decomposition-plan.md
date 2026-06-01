> 状态：Current
> 事实源：`AWDRoundInspector.vue` 当前 owner、AWD inspector 护栏测试、`AWD巡检结果视图架构.md`
> 替代：无

# AWD Round Inspector Decomposition Plan

## 目标

- 把 `AWDRoundInspector.vue` 从“workflow wiring + HUD + tabbed canvas + pane shell + scoped CSS”收口成明确的 inspector owner
- 为 AWD inspector 补齐 stats HUD、canvas workspace 和独立 CSS 文件
- 保持 AWD 轮次巡检的 public API、tab 切换、导出、loading/empty、traffic 转发和 slot contract 不变

## 非目标

- 本轮不调整 `AWDRoundHeaderPanel.vue`、`AWDServiceStatusPanel.vue`、`AWDAttackLogPanel.vue`、`AWDTrafficPanel.vue`、`AWDScoreboardSummaryPanel.vue` 的内部结构
- 本轮不改 `ContestOperations.vue`、`AWDOperationsPanel.vue` 的组合方式
- 本轮不修改 `useAwdInspector*` model 层的 owner、DTO 或导出语义

## 输入依据

- `code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.vue`
- `code/frontend/src/features/awd-inspector/ui/awdInspector.types.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/ContestOperations.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `docs/architecture/features/AWD巡检结果视图架构.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDRoundInspector.vue` 之前已经把 header、service table、attack table、traffic panel、scoreboard panel 拆成独立子件，当前主问题不在 table/filter owner，而在父层继续承载 HUD、sub-tab shell、loading/empty shell 和整段 scoped CSS。
- inspector 的脚本 owner 目前比较清楚：props / emits / slots contract、`activeSubTab`、composable wiring、export / traffic forwarding、`getServiceCheckPresentationResult()`。
- 当前 raw-source 护栏直接读取 `AWDRoundInspector.vue?raw`；如果把 HUD、canvas workspace 和 CSS 拆走，必须同步改成聚合源码视角，否则 extraction / primitive / theme / surface 护栏会误报。

## 设计边界

### `AWDRoundInspector.vue` 本轮继续负责

- `AWDRoundInspectorProps` / emits / slots contract
- `initialTab` -> `activeSubTab` owner
- `useAwdInspectorCoreState` / `useAwdInspectorFormatting` / `useAwdInspectorDerivedData` / `useAwdInspectorExports` wiring
- `getServiceCheckPresentationResult()`
- `exportReviewPackage` / traffic 相关 emit forwarding

### `AWDInspectorStatsHud.vue` 本轮负责

- 四张 summary HUD 卡
- HUD 内图标、文案和统计展示

### `AWDInspectorCanvasWorkspace.vue` 本轮负责

- sub-tabs header
- 导出复盘包按钮
- loading/empty shell
- matrix / scoreboard / attacks / traffic 四个 pane 的组合
- `service-alerts` slot 转接

### `awdRoundInspector.css` 本轮负责

- inspector shell
- HUD 样式
- canvas / tab / loading 样式
- 响应式样式

## 任务切片

### Slice 1：抽出 HUD 与 canvas workspace

- 目标：
  - 新增 `AWDInspectorStatsHud.vue`
  - 新增 `AWDInspectorCanvasWorkspace.vue`
  - 父 inspector 改为只组合 header + HUD + canvas workspace
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspector.test.ts`
- Review focus：
  - 父组件是否仍是唯一 tab / export / traffic forwarding owner
  - 子组件是否只消费 props / callbacks，不重新接管 workflow owner

### Slice 2：抽出 inspector CSS 与 raw-source 护栏

- 目标：
  - 新增 `awdRoundInspector.css`
  - 更新 AWD inspector 相关 raw-source 护栏为聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- Review focus：
  - 样式 owner 是否从父 SFC 收口
  - 护栏是否仍覆盖 extracted panels、theme token 和 canvas shell

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/platform/__tests__/ContestOperations.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上的 HUD / canvas / CSS 混写是否真的收口
  - 当前 feature 内是否留下新的大子组件替代旧大父组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/platform/__tests__/ContestOperations.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDInspectorCanvasWorkspace.vue` 可能仍偏大，因为它继续承接 tabs + loading/empty + pane composition；如果 AWD inspector 后续继续增长，下一刀更适合在 feature 内继续把 canvas header 和 pane shell 细分，而不是回退到父 inspector 再混写。
- 当前工作树里已经有多条未提交的前端收口改动；本轮实现时需要继续按路径保持可拆分提交，不把无关残留混进后续 feature 提交。
