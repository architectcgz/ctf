# Reuse Decision

## Change type
page / component

## Existing code searched
- `code/frontend/src/pages/notifications/NotificationDetailRoutePage.vue`
- `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue`
- `code/frontend/src/features/notifications/*`

## Similar implementations found
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`

## Decision
refactor_existing

## Reason
通知详情页已经有稳定的 feature model owner：`useNotificationDetailPage.ts`。这次不新建读取逻辑或状态 owner，只复用现有 feature model，把 route page 上过重的模板和样式壳收口到一个 widget，让 `pages` 层回到纯组合入口。

## Files to modify
- `code/frontend/src/pages/notifications/NotificationDetailRoutePage.vue`
- `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/widgets/notification-detail-workspace/index.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-notification-detail-route-widget-cleanup-plan.md`
- `.harness/reuse-decisions/notification-detail-route-widget-cleanup.md`

## After implementation
- `NotificationDetailRoutePage.vue` 只保留 props、feature model 调用和 widget 组合。
- 通知详情展示壳迁到 `widgets/notification-detail-workspace`。
- 迁移台账不再把 `NotificationDetailRoutePage.vue` 记为优先瘦身对象。
