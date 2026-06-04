# Reuse Decision

## Change type
frontend shared style contract convergence

## Existing code searched
- `code/frontend/src/assets/styles/teacher-surface.css`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`

## Similar implementations found
- `teacher-surface.css` 已经承接 `teacher-summary` / `teacher-summary-grid` 的基础 owner，说明 teacher 页通用 summary family 应继续落在这个共享样式入口，而不是再新建第二套全局 teacher summary CSS。
- `ClassStudentsOverviewPanel.vue` 的顶部 summary 与 teacher 页面概览条带属于同一 teacher summary family，但目前仍靠局部 `class-overview-summary` 补 `padding: 0`。
- `StudentManagementPage.vue` 仍在页面本地持有 `@media` 下的 `.teacher-summary-grid` 响应式规则，属于 shared owner 残留缺口。

## Decision
extend_existing

## Reason
- 这轮不是重做 class-students 页面结构，而是把仍散在 consumer 内的 teacher summary family 基础 contract 收回 `teacher-surface.css`。
- 最小正确路径是扩现有 shared teacher surface owner，让 `class-students` 顶部 summary 和 teacher 页面响应式折叠都走显式共享 class，不再留在单页本地。

## Files to modify
- `.harness/reuse-decisions/class-students-teacher-summary-convergence.md`
- `docs/plan/impl-plan/2026-06-04-class-students-teacher-summary-convergence-plan.md`
- `code/frontend/src/assets/styles/teacher-surface.css`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`

## After implementation
- `class-students` 顶部 summary 不再依赖本地 `padding: 0` 补丁，而是声明共享 teacher panel summary contract。
- teacher summary family 的移动端单列折叠不再散落在单页本地媒体查询里。
- teacher raw-source 护栏测试会指向新的 shared owner，而不是继续容忍局部 contract 漂移。
