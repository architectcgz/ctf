# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/notifications/model/useNotificationListPage.ts
- code/frontend/src/views/notifications/NotificationList.vue
- code/frontend/src/views/notifications/__tests__/NotificationList.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts

## Similar implementations found
- `useChallengePackageFormatPage.ts` 已从单次 `router.push()` wrapper 收口成纯 route target contract。
- `NotificationList.vue` 本身是完整 route view 模板，允许直接消费 `RouterLink`，不需要回退到 `useRouter()`。

## Decision
refactor_existing

## Reason
`useNotificationListPage.ts` 当前对 `vue-router` 的依赖只剩一处 `openNotificationDetail()`，本质上只是把通知行跳到详情页：

- 没有 route params / query owner
- 没有 query-tab / alias redirect / role redirect
- 没有必须保留在 router-aware page owner 里的复杂导航工作流

最小正确改动是：

- 把 `openNotificationDetail()` 改成纯 `notificationDetailRoute()` route target helper
- `NotificationList.vue` 直接通过 `RouterLink` 渲染通知行
- 更新 `NotificationList` 测试与 allowlist

这样可以在不改变数据 owner 的前提下，收掉 `featureRouterImportAllowlist` 里 `useNotificationListPage.ts` 这一条。

本轮不做：

- 不处理 `useNotificationDetailPage.ts`
- 不重写通知详情页“返回列表 / 打开关联对象”的导航
- 不继续拆通知页样式或布局

## Files to modify
- .harness/reuse-decisions/notification-list-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-notification-list-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-notification-list-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/notifications/model/useNotificationListPage.ts
- code/frontend/src/views/notifications/NotificationList.vue
- code/frontend/src/views/notifications/__tests__/NotificationList.test.ts

## After implementation
- `useNotificationListPage.ts` 不再 import `vue-router`
- 通知行改由 `RouterLink` 直达详情页
- `featureRouterImportAllowlist` 再减少 1 条
