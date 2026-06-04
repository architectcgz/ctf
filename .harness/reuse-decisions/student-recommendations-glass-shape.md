# Reuse Decision

## Change type
page shell / layout polish

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## Similar implementations found
- `TeacherDashboardPage.vue` 和 `StudentManagementPage.vue` 都让教师端 `workspace-shell` 直接承担 `flex min-h-full flex-1 flex-col`，用来保证外层 surface 在短内容页面里也能撑满 route 容器。
- `StudentAnalysisPage.vue` 已经是学生分析页的 shell owner，适合在这里局部补齐高度与布局约束，而不是改共享 `workspace-shell.css`。

## Decision
extend_existing

## Reason
- 问题只出现在学生分析页的推荐任务 tab：内容较短时，外层玻璃背景没有像其它教师端工作台一样撑满容器，看起来像壳体形状和页面不贴合。
- 现有 `StudentAnalysisPage.vue` 就是 route shell owner，最小正确改动是在这个页面补齐 shell 的高度与内容区伸展约束，不扩大到共享样式。

## Files to modify
- `.harness/reuse-decisions/student-recommendations-glass-shape.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- 学员分析页在 `recommendations` 这类短内容 panel 下，外层玻璃背景会继续撑满工作区高度。
- 推荐任务的 loading、empty、list 三态复用同一个列表容器，并让 `#student-recommendations > div > section > div > div` 这层容器成为玻璃屏形状 owner，避免玻璃面落到外层 tabpanel 或与实际内容区域错位。
- 修正只影响学生分析页 shell，不改共享 `workspace-shell` 的全局形状。
