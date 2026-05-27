# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/teacher/student-management/StudentManagementPage.vue
- code/frontend/src/views/teacher/TeacherStudentManagement.vue
- code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts
- code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts
- code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts
- code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedPaginationControls.test.ts
- code/frontend/src/features/teacher-student-management/index.ts
- code/frontend/src/features/teacher-student-management/model/index.ts
- code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts
- code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue
- code/frontend/src/features/teacher-instances/ui/index.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `TeacherDashboardPage.vue`、`TeacherInstanceManagementPage.vue`、`UserGovernancePage.vue` 已证明，page-sized UI 可以迁到 `features/*/ui`，route view 继续保留组合壳。
- `useStudentManagementPage.ts` 当前已经是学生管理页的业务 owner：初始化、班级筛选、搜索防抖、分页切换、班级管理跳转和学员分析跳转都在 feature model 内。
- `TeacherStudentManagement.vue` 已经是薄 route 壳，只额外组合 `ClassReportExportDialog`，不需要像 `UserGovernancePage` 那样再补 route query owner。

## Decision
refactor_existing

## Reason
这次不是新增学生管理能力，而是继续收口 `components/*Page.vue -> @/features/*` 的遗留例外。最小正确改动是把只承担 page shell 职责的 `StudentManagementPage.vue` 迁到 `features/teacher-student-management/ui/`，并让 `views/teacher/TeacherStudentManagement.vue` 通过 `features/teacher-student-management` public API 组合 page-sized UI 与 page model，同时收掉它对应的 `legacyComponentPageAllowlist`。

## Files to modify
- .harness/reuse-decisions/student-management-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-student-management-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-student-management-feature-ui-migration-review.md
- code/frontend/src/features/teacher-student-management/index.ts
- code/frontend/src/features/teacher-student-management/ui/*
- code/frontend/src/views/teacher/TeacherStudentManagement.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts
- code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts
- code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts
- code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedPaginationControls.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果迁移顺利，教师学生管理页会从 `legacy component page` 通道退出，`teacher-student-management` 也会和 `teacher-dashboard`、`teacher-instances` 一样拥有明确的 `model + ui` public API。
- 本轮不重排 `useStudentManagementPage.ts` 的 router / API / debounce owner，也不改 `ClassReportExportDialog` 的组合位置。
