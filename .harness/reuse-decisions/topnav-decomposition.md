# Reuse Decision

## Change type
frontend architecture / layout component / decomposition

## Existing code searched
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/AppLayout.vue
- code/frontend/src/components/layout/Sidebar.vue
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/composables/useWorkspaceShellNavigation.ts
- code/frontend/src/composables/useBackofficeBreadcrumbDetail.ts
- code/frontend/src/composables/useTheme.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `NotificationDrawer.vue`、`Sidebar.vue` 刚完成的两轮拆分都证明，布局层超大组件可以保留 owner 在父组件，把稳定的视图壳拆到同目录子组件。
- `TopNav.vue` 当前的主要体量来自 breadcrumb、brand picker、通知 trigger、用户卡片和移动 toggle 的模板与样式堆叠，而不是多处独立业务 owner。
- `AppLayout.vue` 已经是 `sidebarCollapsed` / `sidebarOpen` 的外层 owner，`TopNav.vue` 只消费 props 并往外 emit，因此本轮不需要改 layout contract。

## Decision
refactor_existing

## Reason
这次仍然是布局层超大组件收口，不是重做导航或主题系统。最小正确改动是让 `TopNav.vue` 继续保留 route、breadcrumb、brand picker open state、theme/brand action 与通知 trigger slot 的 owner，把稳定展示区块拆到 `components/layout/topnav/*`，降低单文件模板和样式密度，同时不引入第二套状态机。

## Files to modify
- .harness/reuse-decisions/topnav-decomposition.md
- docs/plan/impl-plan/2026-05-27-topnav-decomposition-implementation-plan.md
- docs/reviews/frontend/2026-05-27-topnav-decomposition-review.md
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/topnav/TopNavMobileToggle.vue
- code/frontend/src/components/layout/topnav/TopNavBreadcrumbs.vue
- code/frontend/src/components/layout/topnav/TopNavBrandPicker.vue
- code/frontend/src/components/layout/topnav/TopNavNotificationTrigger.vue
- code/frontend/src/components/layout/topnav/TopNavUserCard.vue
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `TopNav.vue` 继续负责 route-derived breadcrumb detail、brand picker open/close、theme/brand action、`NotificationDrawer` slot 装配与 logout owner。
- 新子组件只承接 header 中的稳定视图壳，不新增 router/store/composable owner。
- 如果拆分顺利，这条 `P2` 布局层超大组件 backlog 在 touched surface 上会把 `TopNav.vue` 也推进到第一轮结构收口完成。
