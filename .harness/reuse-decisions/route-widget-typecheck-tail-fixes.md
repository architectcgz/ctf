# Reuse Decision

## Change type
frontend refactor / route-widget type contract cleanup

## Existing code searched
- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/contestListRoutes.ts`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/pages/notifications/NotificationListRoutePage.vue`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue`
- `code/frontend/src/pages/scoreboard/ScoreboardDetailRoutePage.vue`
- `code/frontend/src/widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`

## Similar implementations found
- `shared/lib/navigation/routeTarget.ts` 已经给路由 target 提供了统一薄契约，route/widget 之间不需要再各自发明“近似 RouteLocationRaw”的本地对象类型。
- `widgets/notification-list-workspace` 已经把 `categoryOptions` 声明成 `ReadonlyArray`，说明可变数组 owner 不应再反推到下游 filter 组件。
- `ScoreboardDetailWorkspace.vue` 当前已经显式允许 `contest` 为 `null | undefined`，因此 formatter contract 也应该对齐这个页面实际空态。

## Decision
refactor_existing

## Reason
当前 `typecheck` 剩余的 3 个失败都来自 route/widget 迁移后的 contract 没完全收口：

- contest list 仍在本地 route target 类型和共享导航 contract 之间漂移
- notification category filter 还把只读 options 当成可变数组
- scoreboard detail widget 已接受 `contest = null`，但 formatter 签名还停在旧的可选参数语义

最小正确改动是：

- 让 contest list route target 直接复用共享 `AppRouteTarget`
- 让 notification category filter 接受只读 options
- 让 scoreboard detail formatter 签名与 widget 空态 contract 一致

本轮不做：

- 不调整 contest / notification / scoreboard 的页面结构
- 不改 API contract
- 不扩大到新的 route/widget 抽取

## Files to modify
- `.harness/reuse-decisions/route-widget-typecheck-tail-fixes.md`
- `docs/plan/impl-plan/2026-05-31-route-widget-typecheck-tail-fixes-plan.md`
- `code/frontend/src/features/contest-detail/model/contestListRoutes.ts`
- `code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`

## After implementation
- 前端 `typecheck` 不再被这 3 个 route/widget contract 漏口阻塞。
- route target、readonly options 和 scoreboard 空态 formatter 都回到各自已有 owner。
