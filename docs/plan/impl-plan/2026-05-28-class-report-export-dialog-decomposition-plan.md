> 状态：Current
> 事实源：`ClassReportExportDialog.vue` 当前 owner、`useClassReportExport()` workflow owner、class report export 护栏测试
> 替代：无

# Class Report Export Dialog Decomposition Plan

## 目标

- 把 `ClassReportExportDialog.vue` 从“dialog shell + context sync watch + 多 section 模板 + 大段样式”收口成明确的 feature dialog owner
- 为 class report export dialog 补齐 context section、preview section、task rail 和独立 CSS 文件
- 保持 teacher / platform 页面侧对 `ClassReportExportDialog` 的 public API、预览行为、导出行为和下载行为不变

## 非目标

- 本轮不调整 `useClassReportExport()` 的 workflow owner
- 本轮不改 `ClassTrendPanel`、`ClassReviewPanel`、`ClassInsightsPanel` 的内部实现
- 本轮不改变导出任务轮询、下载、preview 数据源或 teacher / platform 入口页面

## 输入依据

- `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportDialog.vue`
- `code/frontend/src/features/teacher-class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `useClassReportExport()` 已经是远端 workflow 的明确 owner，当前主问题不在 workflow，而在父 SFC 仍继续承载多个稳定 section 和大体量样式。
- `ClassReportExportDialog.vue` 的脚本 owner 比较清楚：`dialogVisible`、两组 `watch()`、`closeDialog()`。继续把展示区块下沉，不会破坏 workflow 边界。
- 测试目前直接读取 `ClassReportExportDialog.vue?raw`；如果把 section 和样式拆走，必须同步改成聚合源码视角，否则会误报。

## 设计边界

### `ClassReportExportDialog.vue` 本轮继续负责

- `modelValue / defaultClassName / defaultFromDate / defaultToDate` props contract
- `update:modelValue` emits contract
- `dialogVisible`
- `closeDialog()`
- `watch(props.modelValue)` 与 `watch(default context)` 的 sync owner
- `AdminSurfaceModal` shell 与 footer 组合

### `ClassReportExportContextSection.vue` 本轮负责

- 当前教师上下文展示
- 导出设置表单壳
- preview snapshot 指标
- `重新加载预览` / `创建导出任务` 动作入口

### `ClassReportExportPreviewSection.vue` 本轮负责

- preview error
- preview loading skeleton
- preview summary KPI
- `ClassTrendPanel` / `ClassReviewPanel` / `ClassInsightsPanel` 组合
- preview empty state

### `ClassReportExportTaskRail.vue` 本轮负责

- latest task banner
- latest task details
- download button
- guide list

### `classReportExportDialog.css` 本轮负责

- dialog shell grid
- section shell
- context / controls / preview / task rail 样式
- footer 与响应式样式

## 任务切片

### Slice 1：抽出三个稳定 section

- 目标：
  - 新增 `ClassReportExportContextSection.vue`
  - 新增 `ClassReportExportPreviewSection.vue`
  - 新增 `ClassReportExportTaskRail.vue`
  - 父 dialog 改为只组合 section
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- Review focus：
  - 父组件是否仍是唯一 dialog shell / sync owner
  - 子组件是否只消费 props / emits，不重新接管 workflow

### Slice 2：抽出 dialog CSS 与 raw-source 护栏

- 目标：
  - 新增 `classReportExportDialog.css`
  - 更新 raw-source 护栏为聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- Review focus：
  - 样式 owner 是否从父 SFC 收口
  - 护栏是否仍覆盖 `AdminSurfaceModal`、表单原语和导出区块结构

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上的 dialog section / style 混写是否真的收口
  - 当前 feature 内是否留下新的大子组件替代旧大父组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `ClassReportExportContextSection.vue` 可能仍会偏大，因为它同时持有表单区和 preview snapshot；如果这块继续增长，下一刀更适合在 feature 内按“form shell / snapshot shell”继续分，而不是再回到父 dialog。
- 当前工作树里已经有一批未提交的 layout P2 收口改动；本轮实现时需要继续保持可按路径拆分提交，不把无关残留混进后续 feature 提交。
