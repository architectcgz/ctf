# Reuse Decision

## Change type
frontend shared style owner cleanup

## Existing code searched
- `code/frontend/src/components/teacher/teacher-workspace-subpanel.css`
- `code/frontend/src/features/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts`
- `code/frontend/src/assets/styles/teacher-panel-shell.css`

## Similar implementations found
- `teacher-panel-shell.css` 已从 `components/teacher` 收口到 `assets/styles/teacher-panel-shell.css`，说明 teacher 侧跨 feature 共享 panel 样式应落到中立共享样式目录，而不是继续依附某个 feature 或 legacy component 目录。

## Decision
refactor_existing

## Reason
- `teacher-workspace-subpanel.css` 当前同时服务 `class-students-workspace` 与 `teacher-dashboard`，已经不是单个 feature 私有样式。
- 继续挂在 `components/teacher` 下，会让 feature 页面保留对 legacy component 目录的样式反向依赖。
- 最小正确改动是迁入 `assets/styles`，并同步改 import 与测试断言。

## Files to modify
- `.harness/reuse-decisions/teacher-workspace-subpanel-style-owner.md`
- `code/frontend/src/components/teacher/teacher-workspace-subpanel.css`
- `code/frontend/src/assets/styles/teacher-workspace-subpanel.css`
- `code/frontend/src/features/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts`

## After implementation
- `teacher-workspace-subpanel.css` 不再位于 `components/teacher`。
- `ClassStudentsPage.vue` 与 `TeacherDashboardPage.vue` 统一从中立共享样式目录读取 workspace subpanel 样式。
