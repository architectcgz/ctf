# Teacher Class Students Page Owner Cleanup 计划

## Objective

- 把教师班级学生页的 page shell owner 从 `pages/teacher/TeacherClassStudentsRoutePage.vue` 收回 `features/teaching/class-students-workspace/ui`。
- 保持班级学生页的 query tab、insight window、学生跳转和导出默认参数行为不变。

## Non-goals

- 不移动文件路径。
- 不改 `useClassStudentsPage.ts` 内部的数据加载、route target、insight window query 或学生目录 workflow。
- 不改 `ClassStudentsPage.vue` 的展示结构、tab 键盘行为或学生筛选交互。
- 不改 `features/teaching/class-report-export` 的导出流程 owner。

## Source Inputs

- `code/frontend/src/pages/teacher/TeacherClassStudentsRoutePage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/features/teacher/class-management/ui/TeacherClassManagementPage.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这条线优先收口 teacher route page owner，不继续拆 class students workspace workflow。
- `useClassStudentsPage()` 仍被 platform route 复用，因此这次不做 teacher-specific 命名收口，避免把共享 page model 错改成单角色 owner。
- route page 退回薄壳后，feature page 要成为唯一组合层；不能只换个名字，继续让 route page 持有 `ClassReportExportDialog` 或 page model。

## Task Slices

### Slice 1: 新增 feature-owned teacher class students page shell

- 目标：在 `features/teaching/class-students-workspace/ui` 下新增 `TeacherClassStudentsPage.vue`，由它直接调用共享 `useClassStudentsPage()` 并组合 workspace 壳与导出对话框。
- 风险：需要保持导出对话框默认班级、起止日期和打开行为不变。

### Slice 2: route page 退回薄壳并补护栏

- 目标：`pages/teacher/TeacherClassStudentsRoutePage.vue` 只渲染 `TeacherClassStudentsPage`，并同步测试与 backlog 进展。
- 风险：只补 owner 护栏，不重写班级学生页已有运行态用例。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision teacher-class-students-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `git diff --check -- .harness/reuse-decisions/teacher-class-students-page-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-teacher-class-students-page-owner-cleanup-plan.md code/frontend/src/features/teaching/class-students-workspace/ui/TeacherClassStudentsPage.vue code/frontend/src/features/teaching/class-students-workspace/ui/index.ts code/frontend/src/pages/teacher/TeacherClassStudentsRoutePage.vue code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- route page 是否真正退回薄壳，而不是继续持有导出弹窗或 page model。
- feature page 是否成为教师班级学生页的唯一组合层。
- 共享 page model 保持中性命名后，教师班级学生页的 tab、insight window、学生跳转和导出入口是否保持不变。

## Rollback / Recovery

- 如果 route page 或 raw-source 护栏还漏了旧名字，可以继续补公共导出或断言。
- 不能回退成 route page 继续直接组合 `ClassReportExportDialog` 和 page model。
