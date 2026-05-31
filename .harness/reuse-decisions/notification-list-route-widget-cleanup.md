# Reuse Decision

## Change type
page / component

## Existing code searched
- `code/frontend/src/pages/notifications/NotificationListRoutePage.vue`
- `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/features/notifications/*`

## Similar implementations found
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`

## Decision
refactor_existing

## Reason
通知列表页已经有稳定的 feature model owner：`useNotificationListPage.ts`。这次不新建分页、筛选或已读逻辑，只复用现有 feature model，把 route page 上过重的页头、目录和分页壳层下沉到一个 widget，让 `pages` 层回到纯组合入口。

## Files to modify
- `code/frontend/src/pages/notifications/NotificationListRoutePage.vue`
- `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/notification-list-workspace/index.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-notification-list-route-widget-cleanup-plan.md`
- `.harness/reuse-decisions/notification-list-route-widget-cleanup.md`

## After implementation
- `NotificationListRoutePage.vue` 只保留 feature model 调用和 widget 组合。
- 通知列表页壳层迁到 `widgets/notification-list-workspace`。
- 迁移台账不再把 `NotificationListRoutePage.vue` 记为当前优先瘦身对象。
