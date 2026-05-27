# Reuse Decision

## Change type
frontend architecture / layout component / decomposition

## Existing code searched
- code/frontend/src/components/layout/Sidebar.vue
- code/frontend/src/components/layout/AppLayout.vue
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/__tests__/Sidebar.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/composables/useWorkspaceShellNavigation.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 刚完成的 `NotificationDrawer.vue` 拆分已经证明，布局层大组件可以保留 owner 在 `components/layout/`，只把稳定视图壳拆到同目录子组件。
- `Sidebar.vue` 当前的复杂度主要不是业务流程，而是移动端/桌面端双壳和导航树模板重复；route、active module、expand/collapse、navigate owner 已经集中在父组件里。
- `useWorkspaceShellNavigation()` 已经提供统一的 backoffice module contract，本轮不需要新增 route/composable owner。

## Decision
refactor_existing

## Reason
这次仍然是收口布局层超大组件债，不是重做导航架构。最小正确改动是让 `Sidebar.vue` 继续持有 route、module、展开态和导航 owner，把移动端壳、桌面端壳和导航树渲染拆成局部子组件，减少模板重复和样式混堆，同时不打散导航判断逻辑。

## Files to modify
- .harness/reuse-decisions/sidebar-decomposition.md
- docs/plan/impl-plan/2026-05-27-sidebar-decomposition-implementation-plan.md
- docs/reviews/frontend/2026-05-27-sidebar-decomposition-review.md
- code/frontend/src/components/layout/Sidebar.vue
- code/frontend/src/components/layout/sidebar/SidebarDesktopPanel.vue
- code/frontend/src/components/layout/sidebar/SidebarMobilePanel.vue
- code/frontend/src/components/layout/sidebar/SidebarNavTree.vue
- code/frontend/src/components/layout/sidebar/SidebarPanelHeader.vue
- code/frontend/src/components/layout/sidebar/SidebarWorkspaceLabel.vue
- code/frontend/src/components/layout/sidebar/types.ts
- code/frontend/src/components/layout/__tests__/Sidebar.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `Sidebar.vue` 继续作为 layout owner，负责 route shell、`useWorkspaceShellNavigation()`、expanded menu state、active 判定和导航跳转。
- 新子组件只承接桌面壳、移动壳、header/workspace label、nav tree 的稳定渲染。
- 如果拆分顺利，这条 `P2` backlog 会从 `NotificationDrawer.vue` 推进到 `Sidebar.vue`，只剩 `TopNav.vue` 作为下一批布局层大组件重点。
