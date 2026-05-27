# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue
- code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue
- code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue
- code/frontend/src/views/teacher/InstanceManagement.vue
- code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts
- code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts
- code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts
- code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/features/teacher-instances/index.ts
- code/frontend/src/features/teacher-instances/model/index.ts
- code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `TeacherDashboardPage.vue`、`ContestOrchestrationPage.vue`、`UserGovernancePage.vue` 已证明，单一 feature 的 page-sized UI 可以直接迁到 `features/*/ui`，route view 继续保留组合壳。
- `TeacherInstanceManagementPage.vue` 已在 `2026-05-26` 通过 `TeacherInstanceHeroPanel` / `TeacherInstanceDirectorySection` 完成过一次 page decomposition，本轮不是重新拆 owner，而是把剩余 page shell 落到 feature 自己的 `ui/`。
- `useInstanceManagementPage.ts` 当前已经是教师实例页的业务 owner：初始化、筛选、防抖、销毁确认、分页切换和跳回 dashboard 都在 feature model 内，不需要再新增 route hook。

## Decision
refactor_existing

## Reason
这次不是新增教师实例能力，而是继续收口 `components/*Page.vue -> @/features/*` 的遗留例外。最小正确改动是把已经只剩 page shell 职责的 `TeacherInstanceManagementPage.vue` 迁到 `features/teacher-instances/ui/`，并让 `views/teacher/InstanceManagement.vue` 通过 `features/teacher-instances` public API 组合 page model 与 page-sized UI，同时收掉它对应的 `legacyComponentPageAllowlist`。

## Files to modify
- .harness/reuse-decisions/teacher-instance-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-teacher-instance-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-teacher-instance-feature-ui-migration-review.md
- code/frontend/src/features/teacher-instances/index.ts
- code/frontend/src/features/teacher-instances/ui/*
- code/frontend/src/views/teacher/InstanceManagement.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts
- code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts
- code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts
- code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果迁移顺利，教师实例页会从 `legacy component page` 通道退出，`teacher-instances` 也会和 `teacher-dashboard`、`platform-user-management` 一样拥有明确的 `model + ui` public API。
- 本轮不重排 `TeacherInstanceHeroPanel.vue` / `TeacherInstanceDirectorySection.vue` 的目录位置，也不改变教师实例列表的用户可见行为。
