# Reuse Decision

## Change type
frontend refactor / teacher dashboard panel owner tightening

## Existing code searched
- `code/frontend/src/features`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePanelRoute.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/useUserGovernancePanelRoute.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/shared/lib/keyboard/useTabKeyboardNavigation.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `usePlatformUserManagePage.ts` 已经通过 `useRouteQueryTransport()` 成为目录 panel query 的唯一 owner。
- `useContestManagePage.ts` 与 `useContestManagePanelRoute.ts` 也已把 `overview/create` panel owner 从 UI 壳收回 page model。
- `TeacherDashboardPage.vue` 当前仍直接使用 `useUrlSyncedTabs()`，把 query owner 和键盘 tab 交互一起绑在 UI 壳里。

## Decision
refactor_existing

## Reason
当前最小正确切片不是重写教师概览 tabs，而是把 query owner 和键盘交互 owner 分开：

- `useDashboardPage.ts` 承接 `panel` query 的读取与切换。
- 新增纯 helper 统一解析 `overview/portrait/insight/trend/review/intervention`。
- `TeacherDashboardPage.vue` 继续持有 tab 按钮、aria 和键盘导航，但不再自己读写 route query。

本轮不做：

- 不改教师概览数据加载、指标构建和班级管理 route target。
- 不调整各个子 panel 的展示内容。
- 不扩到 teacher class management 或 student analysis 页面。

## Files to modify
- `.harness/reuse-decisions/teacher-dashboard-panel-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-teacher-dashboard-panel-owner-tightening-plan.md`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardPanelRoute.ts`
- `code/frontend/src/features/teacher/dashboard/model/index.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `useDashboardPage.ts` 会成为教师概览 `panel` query 的唯一 owner。
- `TeacherDashboardPage.vue` 不再直接依赖 `useUrlSyncedTabs()`。
- 教师概览会和最近几笔 panel owner 收口一样，统一落到 `page model + shared route transport + UI keyboard helper` 模式。
