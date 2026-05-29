# Reuse Decision

## Change type
frontend refactor / feature router owner cleanup

## Existing code searched
- code/frontend/src/features/notifications/model/useNotificationDrawer.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `widgets/layout-shell/useLayoutNotificationDrawerBridge()` 已经是 layout 对 notifications feature 的唯一 bridge，适合作为通知抽屉这组跨切面 workflow 的 route-aware owner。
- `NotificationDrawer.vue` 当前只通过 layout-shell bridge 消费通知 workflow，本轮不需要再把 router owner 留在 feature model。

## Decision
refactor_existing

## Reason
`featureRouterImportAllowlist` 中，`features/notifications/model/useNotificationDrawer.ts -> vue-router` 不是合理长期例外。这个文件现在的职责是通知抽屉的 feature workflow owner，不应该直接认识 `useRouter()`。

最小正确改动是：

- 让 `useNotificationDrawer()` 改为接收导航 callback，而不是直接使用 `vue-router`
- 让 `useLayoutNotificationDrawerBridge.ts` 成为 route-aware bridge owner，统一接住“查看全部通知 / 查看通知详情”这组导航
- 删除对应 allowlist 条目，并补 raw-source 护栏，防止 router 再漂回 feature model

本轮不做：

- 不改 `NotificationDrawer.vue` 当前的 layout P2 view-state / CSS 收口
- 不改通知 store、`markAllRead()`、trigger slot 或通知详情路由语义
- 不处理 `featureRouterImportAllowlist` 其它剩余条目

## Files to modify
- .harness/reuse-decisions/notification-drawer-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-notification-drawer-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-notification-drawer-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts

## After implementation
- `useNotificationDrawer.ts` 不再 import `vue-router`
- 通知抽屉的导航动作明确回到 `widgets/layout-shell` bridge
- `featureRouterImportAllowlist` 缩小一条
