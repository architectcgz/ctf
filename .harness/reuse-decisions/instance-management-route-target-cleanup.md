# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts
- code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts
- code/frontend/src/views/platform/InstanceManage.vue
- code/frontend/src/views/teacher/InstanceManagement.vue
- code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue
- code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue
- code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue
- code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue
- code/frontend/src/views/platform/__tests__/InstanceManage.test.ts
- code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/features/platform-student-management/model/platformStudentManagementRoutes.ts
- code/frontend/src/features/teacher-student-management/model/teacherStudentManagementRoutes.ts

## Similar implementations found
- `class-management-route-target-cleanup` 已示范“page model 产出 route target，UI 通过 `AppRouteLink` 消费”的模式。
- `student-management-route-target-cleanup` 已示范“平台 / 教师成对目录页”去掉 `vue-router` 后，如何把非路由 workflow owner 原地保留。

## Decision
refactor_existing

## Reason
实例目录这组剩余的 router 依赖仍然是典型的“数据 owner + 顺手单次导航”：

- `usePlatformInstanceManagementPage.ts` 里只有 `openOverview()` 和 `openStudent(studentId, className)` 两条导航，真正需要保留在 page model 里的 owner 是实例目录数据加载、搜索节流、销毁确认和销毁后刷新。
- `useInstanceManagementPage.ts` 里只剩 `openDashboard()` 一条导航，真正需要保留的是教师实例目录的筛选 / 分页 / 销毁 workflow owner。

最小正确改动是：

- 给 `platform-instance-management`、`teacher-instances` 各自补本地 route target helper
- 两个 page model 去掉 `vue-router`
- 平台实例目录的“返回概览”“所属用户”以及教师实例目录的“返回教学概览”改为通过共享 `AppRouteLink` 消费 route target
- 保留平台 / 教师端原有的实例销毁、分页、筛选和刷新 owner

这样可以一次收掉：

- `features/platform-instance-management/model/usePlatformInstanceManagementPage.ts -> vue-router`
- `features/teacher-instances/model/useInstanceManagementPage.ts -> vue-router`

本轮不做：

- 不处理实例目录数据查询 `useInstances.ts` 本身的 API / stale-response owner
- 不扩到学生分析页或概览页内部的 route owner
- 不继续做实例目录的大组件拆分

## Files to modify
- .harness/reuse-decisions/instance-management-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-instance-management-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-instance-management-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-instance-management/model/platformInstanceManagementRoutes.ts
- code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts
- code/frontend/src/features/teacher-instances/model/teacherInstanceManagementRoutes.ts
- code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts
- code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue
- code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue
- code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue
- code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue
- code/frontend/src/views/platform/InstanceManage.vue
- code/frontend/src/views/teacher/InstanceManagement.vue
- code/frontend/src/views/platform/__tests__/InstanceManage.test.ts
- code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts

## After implementation
- `usePlatformInstanceManagementPage.ts` 不再 import `vue-router`
- `useInstanceManagementPage.ts` 不再 import `vue-router`
- 平台 / 教师实例目录都改成 route target contract + 共享 `AppRouteLink`
- `featureRouterImportAllowlist` 再减少 2 条
