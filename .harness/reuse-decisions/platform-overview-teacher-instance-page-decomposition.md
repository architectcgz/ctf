# Reuse Decision

## Change type
page / component / layout / table / dashboard / workspace

## Existing code searched
- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/views/platform/PlatformOverview.vue`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewWorkspace.ts`
- `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/views/teacher/InstanceManagement.vue`
- `code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue`
- `code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/platform/user/UserGovernancePage.vue`

## Similar implementations found
- `TeacherDashboardPage.vue`、`ClassStudentsPage.vue`、`UserGovernancePage.vue` 已经证明当前仓库的大 page 收口模式是：父页继续持有 route / feature owner，把稳定展示块拆到清晰命名的子组件。
- `InstanceManageHeroPanel.vue` 与 `InstanceManageWorkspacePanel.vue` 已经把平台实例管理拆成 “hero + directory workspace” 两块，可直接作为教师实例页展示块拆分的近邻参考。
- `PlatformOverviewPage.vue` 当前 hero / alerts / hotspots 三块边界天然稳定，适合沿最近 workspace page decomposition 的模式拆成 overview hero panel 与 directory sections。

## Decision
refactor_existing

## Reason
- 这次不是新增功能，而是继续收口两处过宽的前端 page/component owner。现有 route view、feature composable、workspace shell、summary card、directory list 和 toolbar 原语都已存在，最小正确方案是复用既有 page decomposition 模式。
- `PlatformOverview.vue` 与 `InstanceManagement.vue` 已经是薄 route view，`usePlatformOverviewPage.ts` 与 `useInstanceManagementPage.ts` 也已经承担 page workflow owner；不应把 owner 再挪进新的 child 组件或新 composable，只应拆稳定展示块。
- 平台概览页与教师实例页当前的 raw-source 护栏很多，拆分后最安全的测试策略仍然是“父页源码 + 实际承载展示块的子组件源码组合断言”，而不是回填注释或放松测试。

## Files to modify
- `.harness/reuse-decisions/platform-overview-teacher-instance-page-decomposition.md`
- `docs/plan/impl-plan/2026-05-26-platform-overview-teacher-instance-page-decomposition-implementation-plan.md`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`

## After implementation
- 如果这两页拆分后形成稳定模式，可以把 “overview hero + directory section 双段式 workspace page 拆分” 记录到本地 `.harness/reuse-index/`，作为后续 `PlatformOverviewPage` / `TeacherInstanceManagementPage` 同类页面的直接参考。
