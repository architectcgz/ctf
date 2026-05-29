# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts
- code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts
- code/frontend/src/views/platform/ClassManage.vue
- code/frontend/src/views/teacher/ClassManagement.vue
- code/frontend/src/features/teacher-class-management/ui/ClassManagementPage.vue
- code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue
- code/frontend/src/views/platform/__tests__/ClassManage.test.ts
- code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts
- code/frontend/src/features/platform-overview/model/platformOverviewRoutes.ts

## Similar implementations found
- `useChallengePackageFormatPage.ts` 已从单次 `router.push()` wrapper 收口成纯 route target contract。
- `platformOverviewRoutes.ts` 已示范同 feature 内将多个薄导航 owner 收口成 route target helper。

## Decision
refactor_existing

## Reason
这轮的两条 allowlist 都属于“数据 owner + 顺手 push”的薄导航 wrapper：

- `usePlatformClassManagementPage.ts` 只剩 `openClass(className)` 一个单次跳转。
- `useClassManagementPage.ts` 里 `openClass(className)` 和 `openDashboard()` 都只是单次跳转，本地 dialog owner `openClassReportDialog()` 才是它真正该保留的本地 workflow。

最小正确改动是：

- 给 `platform-class-management`、`teacher-class-management` 各自补本地 route target helper
- 两个 page model 去掉 `vue-router`，改成返回 route target contract
- 平台班级目录和教师班级目录通过共享 route link bridge 消费 route target
- 保留教师端“导出班级报告”弹窗 owner，不把非路由 workflow 也顺手打散

这样可以一次收掉：

- `features/platform-class-management/model/usePlatformClassManagementPage.ts -> vue-router`
- `features/teacher-class-management/model/useClassManagementPage.ts -> vue-router`

本轮不做：

- 不处理 `useTeacherDashboardPage.ts` 或 `useClassStudentsPage.ts` 的其它 router owner
- 不重写班级目录的数据加载、分页或报告导出流程
- 不继续做班级目录的大组件拆分

## Files to modify
- .harness/reuse-decisions/class-management-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-class-management-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-class-management-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/features/platform-class-management/model/platformClassManagementRoutes.ts
- code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts
- code/frontend/src/components/teacher/class-management/TeacherClassManagementHeaderActions.vue
- code/frontend/src/components/teacher/class-management/TeacherClassManagementRowLink.vue
- code/frontend/src/features/teacher-class-management/model/teacherClassManagementRoutes.ts
- code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts
- code/frontend/src/views/platform/ClassManage.vue
- code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue
- code/frontend/src/features/teacher-class-management/ui/ClassManagementPage.vue
- code/frontend/src/views/teacher/ClassManagement.vue
- code/frontend/src/views/platform/__tests__/ClassManage.test.ts
- code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts

## After implementation
- `usePlatformClassManagementPage.ts` 不再 import `vue-router`
- `useClassManagementPage.ts` 不再 import `vue-router`
- 平台 / 教师班级目录都改成 route target contract + 共享 route link bridge
- `featureRouterImportAllowlist` 再减少 2 条
