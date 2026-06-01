# Reuse Decision

## Change type
layout / composition

## Existing code searched
- code/frontend/src/shared/model/layout/useLayoutSessionActionsBridge.ts
- code/frontend/src/shared/model/layout/useLayoutNotificationDrawerBridge.ts
- code/frontend/src/shared/model/layout/useLayoutNotificationRealtimeBridge.ts
- code/frontend/src/shared/ui/layout/AppLayout.vue
- code/frontend/src/shared/ui/layout/TopNav.vue
- code/frontend/src/shared/ui/layout/NotificationDrawer.vue
- code/frontend/src/shared/model/layout/topnav/useTopNavViewState.ts

## Similar implementations found
- 无 — 三个 bridge 文件是 1:1 迁移，不是新功能。shared→features 层间桥接模式在仓库中无先例。

## Decision
refactor_existing

## Reason
三个 bridge 文件从 shared/model/layout/ 迁至 features/layout/model/，内容不变，仅：
1. 移除 vue-router 直接依赖，改为回调注入（`navigateToLogin`、`goToNotifications` 等）
2. AppLayout/TopNav/NotificationDrawer 改为 props 驱动，AppShellRoutePage 负责组装接线

这是架构边界收口（P0 配置腐化修复），非新功能开发。无需搜索可复用实现。

## Files to modify
- code/frontend/src/features/layout/model/useLayoutSessionActionsBridge.ts (新)
- code/frontend/src/features/layout/model/useLayoutNotificationDrawerBridge.ts (新)
- code/frontend/src/features/layout/model/useLayoutNotificationRealtimeBridge.ts (新)
- code/frontend/src/pages/AppShellRoutePage.vue (新)
- code/frontend/src/shared/model/layout/notificationDrawerController.ts (新)
- code/frontend/src/shared/model/layout/useLayoutSessionActionsBridge.ts (删)
- code/frontend/src/shared/model/layout/useLayoutNotificationDrawerBridge.ts (删)
- code/frontend/src/shared/model/layout/useLayoutNotificationRealtimeBridge.ts (删)
- code/frontend/src/shared/model/layout/useLayoutSessionActionsBridge.test.ts (删)
- code/frontend/src/shared/ui/layout/AppLayout.vue
- code/frontend/src/shared/ui/layout/TopNav.vue
- code/frontend/src/shared/ui/layout/NotificationDrawer.vue
- code/frontend/src/shared/model/layout/topnav/useTopNavViewState.ts
- code/frontend/src/shared/model/layout/index.ts
- code/frontend/src/router/routes/appShellRoute.ts
- code/frontend/scripts/frontend-architecture-policy.json
- code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts

- code/frontend/src/features/layout/model/useLayoutBridgeBoundaries.test.ts (新)
- code/frontend/src/features/layout/model/useLayoutPublicApiBoundary.test.ts (新)
- code/frontend/src/features/auth/model/useAuthPublicApiBoundary.test.ts (新)
- code/frontend/src/features/notifications/model/useNotificationDrawerPublicApiBoundary.test.ts (新)
