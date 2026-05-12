# Reuse Decision

## Change type
component / layout

## Existing code searched
- code/frontend/src/components/teacher/teacher-workspace-subpanel.css
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/components/teacher/teacher-panel-shell.css
- code/frontend/src/assets/styles/teacher-surface.css
- code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts

## Similar implementations found
- `teacher-panel-shell.css` 已经把 panel 边框、分割线、surface 都收口到 `--panel-border` / `--panel-divider` / `--panel-surface*`。
- `teacher-surface.css` 已经提供 `--teacher-card-border`、`--teacher-divider`、`--journal-surface`、`--journal-surface-subtle` 这些教师工作区共享 token。
- `teacher-workspace-subpanel.css` 当前是共享 subpanel 深选择器 owner，应该直接复用上面这些 token，而不是要求每个页面单独补一组 `--teacher-workspace-*` 变量。

## Decision
refactor_existing

## Reason
这次问题不是单个 review panel 的局部边框公式，而是共享 subpanel 壳层引用了一组未定义的 `--teacher-workspace-*` 变量，导致边框和背景回退异常。把 fallback 直接收口在共享壳层里，能一次修正班级详情和教师概览等复用面，不需要在每个页面再手工补变量。

## Files to modify
- .harness/reuse-decisions/teacher-workspace-subpanel-token-fallbacks.md
- code/frontend/src/components/teacher/teacher-workspace-subpanel.css
- code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts
