> 状态：Current
> 事实源：training timeline shared panel owner
> 替代：无

# Training Timeline Panel Owner Normalization Plan

## 目标

- 把 `StudentTimelinePage.vue` 收口成中立共享 panel，停止以 `components/**Page.vue` 的方式继续存在。
- 让 student dashboard 和 teacher 学员洞察都切到新的共享 timeline panel 入口。

## 非目标

- 本轮不重做时间线视觉样式或分页交互。
- 本轮不改 `timelineSummary`、`timelineTypeLabel`、`timelineTypeTone` 的语义。
- 本轮不动 `StudentAnalysisPage.vue` 的 section owner，只替换它透传到 `StudentInsightPanel.vue` 的共享 panel 入口。

## 输入依据

- `code/frontend/src/components/dashboard/student/StudentTimelinePage.vue`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts`
- `code/frontend/src/components/dashboard/student/utils.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 时间线面板不是单一 feature 的 page，它已经被 student dashboard 和 teacher 学员洞察两处消费。
- 继续把它留在 `components/dashboard/student/StudentTimelinePage.vue` 只会让“共享 panel”继续伪装成“student dashboard 页面”。
- 最小收口面是改成中立 panel 命名，并同步切换引用和 raw-source 护栏。

## 设计边界

### `components/training/TrainingTimelinePanel.vue` 本轮负责

- 训练时间线展示
- metric-panel 摘要
- 本地分页状态
- timeline 分组和 item 呈现

### Student dashboard / teacher insight 本轮继续负责

- 各自的数据 owner
- 何时显示 timeline section
- timeline 数据的上游加载与错误策略

## 任务切片

### Slice 1：共享 panel 命名收口

- 目标：
  - 新增 `components/training/TrainingTimelinePanel.vue`
  - student dashboard registry 与 `StudentInsightPanel.vue` 都切到新入口
- 验证：
  - `npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Review focus：
  - timeline 的分页 / 分组 owner 是否仍留在共享 panel
  - dashboard / teacher 是否只替换入口，不重新实现 timeline

### Slice 2：allowlist 与 raw-source 护栏同步

- 目标：
  - 更新 `architectureAllowlist.ts`
  - 更新所有 student timeline 相关 raw-source 测试到新路径
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/metricPanelSurfaceOwnership.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - `legacyComponentPageAllowlist` 是否清空 timeline 这一条
  - metric-panel 和 theme token 护栏是否仍覆盖到新 panel

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/metricPanelSurfaceOwnership.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- 时间线展示虽然已改成中立共享 panel，但它的工具函数还会继续放在原 `components/dashboard/student/utils.ts`。后续如果继续清理 shared/student 命名，可以再把 timeline 相关 helper 从这个 student 命名文件里拆出来。
