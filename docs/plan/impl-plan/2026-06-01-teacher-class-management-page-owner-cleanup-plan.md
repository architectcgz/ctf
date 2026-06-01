# Teacher Class Management Page Owner Cleanup 计划

## Objective

- 把教师班级管理页的 page shell owner 从 `pages/teacher/ClassManagementRoutePage.vue` 收回 `features/teacher/class-management/ui`。
- 把 `useClassManagementPage()` 的命名收紧成 teacher-specific page-model owner。
- 保持班级目录加载、筛选、分页、跳转和导出默认班级名行为不变。

## Non-goals

- 不移动文件路径。
- 不改 `useTeacherClassDirectory.ts` 的分页 / 加载行为。
- 不改 `ClassManagementPage.vue` 的目录展示结构和筛选交互。
- 不改 `features/teaching/class-report-export` 的导出流程 owner。

## Source Inputs

- `code/frontend/src/pages/teacher/ClassManagementRoutePage.vue`
- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceManagementPage.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这条线优先收口 page owner 和 page-model 命名，不继续拆 class directory workflow。
- route page 退回薄壳后，feature page 要成为唯一组合层；不能只换一个文件名，继续让 route page 持有 report dialog 或 page model。

## Task Slices

### Slice 1: 新增 feature-owned teacher class management page shell

- 目标：在 `features/teacher/class-management/ui` 下新增 `TeacherClassManagementPage.vue`，由它直接调用 teacher-specific page model 并组合目录页壳与导出对话框。
- 风险：需要保持导出对话框的默认班级名和打开行为不变。

### Slice 2: 收紧 teacher class management page-model 命名

- 目标：把 `useClassManagementPage()` 改成 `useTeacherClassManagementPage()`，同步 feature public API 与新 page shell 消费。
- 风险：raw-source 护栏和 route page 引用要一起更新，否则容易残留旧 owner 表达。

### Slice 3: route page 退回薄壳并补护栏

- 目标：`pages/teacher/ClassManagementRoutePage.vue` 只渲染 `TeacherClassManagementPage`，并同步测试与 backlog 进展。
- 风险：只补 owner 护栏，不重写目录页已有运行态用例。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision teacher-class-management-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/ClassManagement.test.ts`
- `git diff --check -- .harness/reuse-decisions/teacher-class-management-page-owner-cleanup.md docs/plan/impl-plan/2026-06-01-teacher-class-management-page-owner-cleanup-plan.md code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts code/frontend/src/features/teacher/class-management/model/index.ts code/frontend/src/features/teacher/class-management/ui/TeacherClassManagementPage.vue code/frontend/src/features/teacher/class-management/ui/index.ts code/frontend/src/pages/teacher/ClassManagementRoutePage.vue code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- route page 是否真正退回薄壳，而不是继续持有导出弹窗或 page model。
- feature page 是否成为教师班级管理页的唯一组合层。
- 命名收口后，班级目录加载、筛选、分页、跳转和导出入口是否保持不变。

## Rollback / Recovery

- 如果 route page 或 raw-source 护栏还漏了旧名字，可以继续补公共导出或断言。
- 不能回退成 route page 继续直接组合 `ClassReportExportDialog` 和 page model。
