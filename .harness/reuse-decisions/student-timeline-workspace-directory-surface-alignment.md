# Reuse Decision

## Change type
page / component / layout

## Existing code searched
- code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue
- code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPanelBindings.ts
- code/frontend/src/style.css
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/__tests__/studentJournalSoftStyles.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## Similar implementations found
- `code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue` 已经使用 `workspace-directory-section`、`workspace-directory-list`、`workspace-directory-pagination` 承接学生侧列表区域的统一壳层和节奏。
- `code/frontend/src/style.css` 已经定义了 `workspace-directory-*` 的列表外框、空态、分页和垂直间距，不需要再为时间线面板单独造一套列表区域容器。
- `TrainingTimelinePanel.vue` 当前暴露出 mixed owner 问题：学生页壳和教师学员详情 section 壳通过 `embedded` 混在一个 entity 组件里，说明它不再适合作为两端共同的 layout owner。

## Decision
extend_existing

## Reason
这次需求表面上是列表区域对齐，但代码上下文已经表明 `embedded` 在切两套布局语义。更优雅的最小正确改动是：
- 抽 `TrainingTimelineContent.vue` 作为共享内容 owner
- 保留 `TrainingTimelinePanel.vue` 作为学生 dashboard 壳
- 新增教师侧 `StudentInsightTimelineSection.vue` 承接 section 壳与 loading skeleton

这样既继续复用现有 `workspace-directory-*` contract，也把 layout owner 从 `embedded` 开关里拆出来。

## Files to modify
- .harness/reuse-decisions/student-timeline-workspace-directory-surface-alignment.md
- code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/__tests__/metricPanelSurfaceOwnership.test.ts

## After implementation
- `TrainingTimelineContent.vue` 会成为训练记录 header、指标卡、列表壳、分页和事件分组的共享内容 owner。
- `TrainingTimelinePanel.vue` 只保留学生 dashboard 页面壳；教师学员详情改由 `StudentInsightTimelineSection.vue` 承接 section 壳。
- 记录列表区域会复用 `workspace-directory-section`、`workspace-directory-list` 和 `workspace-directory-pagination` 的通用样式契约。
- 这次属于既有 contract 的 owner 收口，不新增新的长期复用模式，默认不需要补 `.harness/reuse-index/`。
