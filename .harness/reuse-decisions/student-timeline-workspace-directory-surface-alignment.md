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
- `TrainingTimelinePanel.vue` 当前已经是 student dashboard 与 teacher 学员洞察共享的中立展示面板，因此记录列表区域的样式 owner 应继续收口在该共享组件内部，而不是回退到某个调用页局部覆盖。

## Decision
extend_existing

## Reason
这次需求是把训练记录区域对齐到现有通用列表区域样式。最小正确改动是直接让 `TrainingTimelinePanel.vue` 复用现有 `workspace-directory-*` 壳层语义，再用少量局部变量桥接到时间线的分组和事件行表现；没有必要新增新组件，也不应该把样式散落到 student dashboard 或 teacher 学员详情页面做双份覆盖。

## Files to modify
- .harness/reuse-decisions/student-timeline-workspace-directory-surface-alignment.md
- code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts

## After implementation
- 训练记录面板会继续由 `TrainingTimelinePanel.vue` 单点 owner，分页和分组逻辑不变。
- 记录列表区域会复用 `workspace-directory-section`、`workspace-directory-list` 和 `workspace-directory-pagination` 的通用样式契约。
- 这次只是现有列表区域样式的复用收敛，不新增新的长期复用模式，默认不需要补 `.harness/reuse-index/`。
