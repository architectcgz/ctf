# Reuse Decision

## Change type
frontend refactor / layout notification drawer local owner cleanup

## Existing code searched
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/notification-drawer/types.ts
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `NotificationDrawer.vue` 上一轮已经按 layout owner 收口成“trigger / shell / filter state / dismiss owner”，并把 header、summary、tabs、body、footer 稳定视图区块拆到 `components/layout/notification-drawer/*`；这说明当前应该继续沿“父层只保留明确 owner，其余稳定局部再下沉”的方式推进，而不是回退成新的大子组件。
- `widgets/layout-shell/useLayoutNotificationDrawerBridge()` 已经是 layout 对 notifications feature 的唯一桥接；本轮不应把 filter / shell 生命周期之类的本地视图逻辑继续塞回 feature bridge。
- `PlatformContestFormPanel.vue`、`ContestProjectorAttackMap.vue` 最近几轮都已经按“父层保留唯一 workflow/view-model owner，局部 mechanics 和稳定 shell 下沉”的模式完成收口，`NotificationDrawer.vue` 的剩余混写面也适合沿同一模式继续缩小。

## Decision
refactor_existing

## Reason
`NotificationDrawer.vue` 当前约 `496` 行，虽然远端通知 workflow 已经经由 `widgets/layout-shell` bridge 收口，但父组件里仍同时混放：

- 本地视图 owner：`activeFilter`、`filteredItems`、`emptyState`
- 局部展示派生：`hasUnread`、`unreadBadgeLabel`、`drawerSummary`
- overlay lifecycle：`Escape` 关闭、`document.body.style.overflow` scroll lock
- 抽屉壳体样式

最小正确改动不是继续把通知 feature 逻辑拉回 layout，也不是再新增一个混装 shell 子组件，而是：

- 保持 `NotificationDrawer.vue` 继续作为 layout shell owner，负责 trigger slot、bridge workflow 装配和 panel 组合
- 新增 `useNotificationDrawerViewState.ts`，承接 filter / summary / empty state / overlay lifecycle 这组本地视图逻辑
- 新增 `notificationDrawer.css`，把壳体与 trigger 样式从父 SFC 中移出
- 同步把 raw-source 与 theme-token 护栏改成“父 SFC + view-state composable + css + 子组件”的聚合源码视角

本轮不修改 `useNotificationDrawer()` 的 feature workflow owner，不改通知 store、路由跳转、mark-all-read 语义，也不再新增 layout bridge。

## Files to modify
- .harness/reuse-decisions/notification-drawer-owner-cleanup.md
- docs/plan/impl-plan/2026-05-28-notification-drawer-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-28-notification-drawer-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/notification-drawer/useNotificationDrawerViewState.ts
- code/frontend/src/components/layout/notification-drawer/notificationDrawer.css
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts

## After implementation
- `NotificationDrawer.vue` 会回到“layout shell owner + bridge owner”这一层，不再继续内联 filter / summary / dismiss lifecycle / 壳体样式。
- 通知抽屉的本地视图状态和 overlay cleanup 会有单点 owner，后续继续清理 `Sidebar.vue`、`TopNav.vue` 时边界更一致。
- backlog 里这条 layout P2 会从“通知抽屉更深层行为清理”进一步收敛到 `Sidebar.vue`、`TopNav.vue` 及通知抽屉后续是否还存在更深层样式或 contract 债。
