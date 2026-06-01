# Reuse Decision

## Change type
frontend refactor / teacher class students page owner cleanup

## Existing code searched
- `code/frontend/src/pages/teacher/TeacherClassStudentsRoutePage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/features/teacher/class-management/ui/TeacherClassManagementPage.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `TeacherClassManagementPage.vue` 已经把教师班级管理页的 page shell owner 收口到 feature 内部，route page 只保留薄壳。
- `TeacherInstanceManagementPage.vue`、`PlatformInstanceManagementPage.vue` 等最近目录页迁移都在让 route page 退回薄壳，feature page 自己组合 page model 与内部 workspace。
- 当前 `TeacherClassStudentsRoutePage.vue` 仍直接组合 `ClassStudentsPage`、`ClassReportExportDialog` 和 `useClassStudentsPage()`，是 teacher class workspace 线上剩余的 route-page owner 残片。
- `useClassStudentsPage()` 同时还被 `PlatformClassStudentsRoutePage.vue` 复用，说明这条 page model 属于共享 class students workspace，而不是 teacher 私有命名面。

## Decision
refactor_existing

## Reason
下一轮最小正确切片不是继续拆班级学生 workspace workflow，而是先把教师班级学生页对齐到当前 page-owner 模式：

- 在 `features/teaching/class-students-workspace/ui` 下新增 `TeacherClassStudentsPage.vue`，让 feature 自己组合 page model、workspace 壳和导出对话框。
- `TeacherClassStudentsRoutePage.vue` 退回只渲染 feature page 的薄壳，不再直接 import page model 或导出对话框。
- 继续保留共享 `useClassStudentsPage()` 命名，不把 platform 也在消费的 page model 误改成 teacher-specific owner。
- 不移动文件路径，不改班级学生页的 query tab、insight window、学生跳转或导出默认参数逻辑。

## Files to modify
- `.harness/reuse-decisions/teacher-class-students-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-teacher-class-students-page-owner-cleanup-plan.md`
- `code/frontend/src/features/teaching/class-students-workspace/ui/TeacherClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/index.ts`
- `code/frontend/src/pages/teacher/TeacherClassStudentsRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 教师班级学生页会和教师班级管理页一样，由 feature page 持有 page shell owner。
- route page 只保留 route-level 渲染职责，不再直接耦合导出对话框或 page model。
- 共享 `class-students-workspace` page model 继续保持中性命名，避免和 platform consumer 发生边界冲突。
