# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts
- code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue
- code/frontend/src/views/teacher/TeacherDashboard.vue
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/features/teacher-class-management/model/teacherClassManagementRoutes.ts
- code/frontend/src/utils/classManagementRouting.ts

## Similar implementations found
- `class-management-route-target-cleanup` 已示范“page model 只产出 route target，UI 通过 `AppRouteLink` 消费”的模式。
- `instance-management-route-target-cleanup` 已示范“header 中单次返回导航改成 route target，不动数据 / retry owner”的做法。

## Decision
refactor_existing

## Reason
`useDashboardPage.ts` 当前的 router 依赖只剩 `openClassManagement()` 一条导航：

- 真正需要留在 page model 里的 owner 是 `getTeacherOverview()` 的加载、错误态与 retry。
- “进入班级管理”只是一次单跳转，不需要继续让 dashboard page model 持有 `vue-router`。

最小正确改动是：

- 给 `teacher-dashboard` 补本地班级管理 route target helper
- `useDashboardPage.ts` 去掉 `vue-router`，改为暴露 `classManagementRoute`
- `TeacherDashboardPage.vue` 通过共享 `AppRouteLink` 消费 route target
- `TeacherDashboard.vue` 继续只组合 feature page model 与 page shell

这样可以收掉：

- `features/teacher-dashboard/model/useDashboardPage.ts -> vue-router`

本轮不做：

- 不处理 student dashboard 的 tab query / challenge 跳转 owner
- 不改教师概览 tab 切换、数据构建和 retry 流程
- 不继续做 dashboard 面板拆分

## Files to modify
- .harness/reuse-decisions/teacher-dashboard-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-teacher-dashboard-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-teacher-dashboard-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/teacher-dashboard/model/teacherDashboardRoutes.ts
- code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts
- code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue
- code/frontend/src/views/teacher/TeacherDashboard.vue
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts

## After implementation
- `useDashboardPage.ts` 不再 import `vue-router`
- 教师概览页的“班级管理”入口改成 route target + `AppRouteLink`
- `featureRouterImportAllowlist` 再减少 1 条
