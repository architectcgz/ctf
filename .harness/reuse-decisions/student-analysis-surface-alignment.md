# Reuse Decision

## Change type
component / page styling

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/assets/styles/teacher-surface.css`
- `code/frontend/src/assets/styles/surface-shell-background.css`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## Similar implementations found
- `ClassStudentsPage.vue` 已经使用 `workspace-shell teacher-management-shell teacher-surface` 作为教师班级详情页根壳。
- `ClassManagementPage.vue` 与 `StudentManagementPage.vue` 都复用同一套教师侧 workspace surface 约定，而不是把页面根壳设成透明。
- `teacher-surface.css` 与 `surface-shell-background.css` 已提供教师工作区的共享背景、边框和强调色收口，不需要为学生详情页再保留一套透明特例。

## Decision
extend_existing

## Reason
- 用户反馈的是学生详情页背景和其他教师页不一致，不是要新增一套新主题。
- 最小正确改动是让学生详情页回到现有教师工作区 surface 合同，复用 `teacher-management-shell + teacher-surface`，并去掉页面局部把根壳重置成透明的特例。

## Files to modify
- `.harness/reuse-decisions/student-analysis-surface-alignment.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- 学员详情页根背景与教师侧其他页面保持同一套 surface 语言，light / dark 下都不再直接露出外层主题渐变底。
- 现有内容区、tab、面板和数据流 owner 不变，只收口页面根壳样式。
