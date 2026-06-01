# Platform Overview / Teacher Instance Page Decomposition 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-26-platform-overview-teacher-instance-page-decomposition-implementation-plan.md`
  - files reviewed：
    - `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
    - `code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue`
    - `code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue`
    - `code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue`
    - `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
    - `code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue`
    - `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
    - 本轮 touched 的 13 个前端测试护栏
- Classification check：同意本轮属于前端 `TD-1` 结构性收口，且改动边界与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `PlatformOverviewPage.vue` 继续持有 `usePlatformOverviewWorkspace` 的派发展示数据装配；route 级 owner 仍留在 `PlatformOverview.vue` 与 `usePlatformOverviewPage.ts`，新拆出的三个子组件只承接 hero、alerts、hotspots 这三块稳定展示区。
- `TeacherInstanceManagementPage.vue` 继续只做页面壳层和事件桥接；真正的列表请求、筛选、防抖、销毁、分页 owner 仍留在 `useInstanceManagementPage.ts` / `useInstances.ts`，新拆出的两个子组件没有反向吸入业务流程。
- raw-source 护栏已经统一改成“父页源码 + 子组件源码”的组合断言模式，避免后续继续细分展示块时被迫把模板和局部样式重新塞回父页。
- 这轮没有触碰 `PlatformOverview.vue`、`InstanceManagement.vue` 的 route 组合职责，也没有改动平台 / 教师实例相关 contract。

## Required re-validation

- `npm run test:run -- src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
- `npm run test:run -- src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/adminRootHeroLayout.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check -- code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts docs/plan/impl-plan/2026-05-26-platform-overview-teacher-instance-page-decomposition-implementation-plan.md .harness/reuse-decisions/platform-overview-teacher-instance-page-decomposition.md docs/reviews/frontend/2026-05-26-platform-overview-teacher-instance-page-decomposition-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Residual risk

- 当前 residual risk 不在本轮 touched surface，而在 backlog 里尚未进入本轮的其他 oversized workspace shell 和跨 feature owner 耦合问题。
- `TeacherInstanceDirectorySection.vue` 仍保留现有 `@username` 展示文案，这与更长期的前端文案 / 语义收口目标有关，但不属于本轮“拆稳定展示块”的边界。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `PlatformOverviewPage.vue` 与 `TeacherInstanceManagementPage.vue` 两个 oversized component page。
- 该债务在 touched surface 上已完成收口：父页继续只保留 page shell、事件桥接和展示数据派发，稳定展示块已经拆成独立子组件，相关 raw-source 护栏、类型检查和 workflow gate 已同步通过。
