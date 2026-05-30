# Reuse Decision

## Change type
frontend refactor / layout sidebar local navigation owner cleanup

## Existing code searched
- code/frontend/src/components/layout/Sidebar.vue
- code/frontend/src/components/layout/sidebar/SidebarDesktopPanel.vue
- code/frontend/src/components/layout/sidebar/SidebarMobilePanel.vue
- code/frontend/src/components/layout/sidebar/SidebarNavTree.vue
- code/frontend/src/components/layout/sidebar/types.ts
- code/frontend/src/components/layout/__tests__/Sidebar.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/composables/useWorkspaceShellNavigation.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `Sidebar.vue` 上一轮已经按“route/navigation owner 留父组件，移动壳 / 桌面壳 / nav tree 拆子组件”的方式收口到 `components/layout/sidebar/*`；说明这条线已经接受“父层只做明确 owner，其余稳定区块继续下沉”的模式。
- `NotificationDrawer.vue` 本轮刚完成“父层保留 shell owner，局部 view state 与 shell CSS 下沉到 composable / css file”的收口；`Sidebar.vue` 当前剩余问题与它同型，适合沿同样的切法推进。
- `useWorkspaceShellNavigation()` 已经是 layout 识别 student / academy / platform workspace 的统一桥接；本轮不应把这些 workspace 识别规则重新散落回各个 item 判定分支里。

## Decision
refactor_existing

## Reason
`Sidebar.vue` 当前约 `617` 行，父组件仍同时混放：

- 本地 nav view-model owner：`expandedMenus`、parent/child active 判定、expanded/highlight class 判定
- navigation workflow：`navigate()` 的同路由短路与 mobile close
- sidebar shell / theme token 样式

最小正确改动不是继续在父 SFC 上叠加辅助函数，也不是把 route / auth shell owner 再分散给桌面或移动面板，而是：

- 保持 `Sidebar.vue` 继续作为 layout shell owner，负责 props/emits 和 desktop/mobile panel 组合
- 新增 `useSidebarNavigationViewState.ts`，承接 sidebar 的本地 nav 派生、展开态和 navigate workflow
- 新增 `sidebarShell.css`，承接侧栏壳体与 token 样式
- 同步把 raw-source 与 theme-token 护栏改成“父 SFC + composable + css + 子组件”的聚合源码视角

本轮不调整 `useWorkspaceShellNavigation()` 的 contract，不改 backofficeNavigation 配置，不改变 student / teacher / admin 的可见模块与二级导航语义。

## Files to modify
- .harness/reuse-decisions/sidebar-owner-cleanup.md
- docs/plan/impl-plan/2026-05-28-sidebar-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-28-sidebar-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components/layout/Sidebar.vue
- code/frontend/src/components/layout/sidebar/useSidebarNavigationViewState.ts
- code/frontend/src/components/layout/sidebar/sidebarShell.css
- code/frontend/src/components/layout/__tests__/Sidebar.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts

## After implementation
- `Sidebar.vue` 会回到“layout shell owner + panel composition owner”这一层，不再继续内联 nav 判定、展开态与壳体样式。
- sidebar 的本地 nav state 与 navigate 短路会有单点 owner，后续继续清理 `TopNav.vue` 时可沿同一模式推进。
- backlog 里 layout P2 的剩余重点会进一步收敛到 `TopNav.vue` 以及 sidebar/topnav 更深层 route / breadcrumb / user-card owner。
