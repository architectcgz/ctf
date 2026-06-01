# Reuse Decision

## Change type
frontend refactor / teacher class management page owner cleanup

## Existing code searched
- `code/frontend/src/pages/teacher/ClassManagementRoutePage.vue`
- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceManagementPage.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `TeacherInstanceManagementPage.vue` 已经把教师实例页的 page shell 收口在 feature 内部，route page 只保留组合。
- `PlatformInstanceManagementPage.vue`、`TeacherDashboardRoutePage.vue` 等最近迁移都在让 route page 退回薄壳，feature page 自己组合 page model 与内部 panel。
- 当前 `ClassManagementRoutePage.vue` 仍直接组合 `ClassManagementPage`、`ClassReportExportDialog` 和 `useClassManagementPage()`，是 teacher 目录页里剩余的 route-page owner 残片。

## Decision
refactor_existing

## Reason
下一轮最小正确切片不是继续拆班级目录 workflow，而是先把教师班级管理页对齐到当前 page-owner 模式：

- 在 `features/teacher/class-management/ui` 下新增 `TeacherClassManagementPage.vue`，让 feature 自己组合 page model、目录页壳和导出对话框。
- `ClassManagementRoutePage.vue` 退回只渲染 feature page 的薄壳，不再直接 import page model 或导出对话框。
- `useClassManagementPage()` 改成显式 teacher owner 命名，让 route shell 和 raw-source 护栏直接体现边界。
- 不移动文件路径，不改班级目录加载、筛选、分页、跳转或导出默认班级名逻辑。

## Files to modify
- `.harness/reuse-decisions/teacher-class-management-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-teacher-class-management-page-owner-cleanup-plan.md`
- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher/class-management/model/index.ts`
- `code/frontend/src/features/teacher/class-management/ui/TeacherClassManagementPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/index.ts`
- `code/frontend/src/pages/teacher/ClassManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 教师班级管理页会和教师实例页一样，由 feature page 持有 page shell owner。
- route page 只保留 route-level 渲染职责，不再直接耦合导出对话框或 page model。
- teacher class management feature 内部不再继续暴露无 owner 的 `useClassManagementPage()` 命名。
