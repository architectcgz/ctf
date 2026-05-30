# Reuse Decision

## Change type
frontend refactor / layout topnav local view-state cleanup

## Existing code searched
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/topnav/TopNavBrandPicker.vue
- code/frontend/src/components/layout/topnav/TopNavBreadcrumbs.vue
- code/frontend/src/components/layout/topnav/TopNavMobileToggle.vue
- code/frontend/src/components/layout/topnav/TopNavNotificationTrigger.vue
- code/frontend/src/components/layout/topnav/TopNavUserCard.vue
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/composables/useBackofficeBreadcrumbDetail.ts
- code/frontend/src/composables/useWorkspaceShellNavigation.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `TopNav.vue` 上一轮已经按“route/theme/notification/logout owner 留父组件，移动 toggle / breadcrumbs / brand picker / notification trigger / user card 拆子组件”的方式收口到 `components/layout/topnav/*`；这说明本轮更适合继续收局部 view-model 和 shell CSS，而不是再做新的模板拆分。
- `NotificationDrawer.vue`、`Sidebar.vue` 本轮都已经按“父层只保留 shell composition owner，本地 view state + shell CSS 下沉到 composable / css file”的方式继续瘦身；`TopNav.vue` 当前剩余问题与这两者同型，适合沿用同一模式。
- `useWorkspaceShellNavigation()` 与 `useBackofficeBreadcrumbDetail()` 已经分别承接 workspace 导航事实和详情标题 override；本轮不应把这些职责散回多个模板分支或子组件里。

## Decision
refactor_existing

## Reason
`TopNav.vue` 当前约 `680` 行，父组件仍同时混放：

- 本地 view-state：`isMobile`、`brandPickerOpen`、brand picker outside click / `Escape` 关闭
- breadcrumb detail label 推导：各类 challenge / class / student / contest / review 详情标题派生
- 顶栏 shell / theme token 样式

最小正确改动不是继续叠加 helper 函数，也不是把 route/theme/session bridge 重新分散给子组件，而是：

- 保持 `TopNav.vue` 继续作为 layout shell owner，负责 props contract、notification trigger slot 和 top-level composition
- 新增 `useTopNavViewState.ts`，承接 mobile 判定、brand picker lifecycle、本地展示派生和 breadcrumb detail label 推导
- 新增 `topNavShell.css`，承接 topnav shell 与 theme token 样式
- 同步把 raw-source 与 theme-token 护栏改成“父 SFC + view-state composable + css + 子组件”的聚合源码视角

本轮不调整 `useTheme()`、`useLayoutSessionActionsBridge()`、`useWorkspaceShellNavigation()`、`useBackofficeBreadcrumbDetail()` 的 public contract，不改变 breadcrumb 展示语义，也不修改 `NotificationDrawer` 的 trigger slot contract。

## Files to modify
- .harness/reuse-decisions/topnav-owner-cleanup.md
- docs/plan/impl-plan/2026-05-28-topnav-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-28-topnav-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/topnav/useTopNavViewState.ts
- code/frontend/src/components/layout/topnav/topNavShell.css
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts

## After implementation
- `TopNav.vue` 会回到“layout shell owner + top-level composition owner”这一层，不再继续内联 mobile 判定、brand picker lifecycle、breadcrumb detail 推导和壳体样式。
- layout P2 的三大布局组件收口会基本完成，剩余重点将从“大组件壳体/本地 view-state 混写”转向更深层 contract 或表现层细节。
