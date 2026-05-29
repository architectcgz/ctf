# Reuse Decision

## Change type
frontend refactor / composable boundary cleanup

## Existing code searched
- code/frontend/src/composables/useWebSocket.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/notifications/model/useNotificationRealtime.ts
- code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.ts
- code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts
- code/frontend/src/features/awd-inspector/model/useContestAwdPreviewRealtime.ts

## Similar implementations found
- `useNotificationRealtime.ts`、`useContestScoreboardRealtime.ts`、`useContestAnnouncementRealtime.ts`、`useContestAwdPreviewRealtime.ts` 都把业务事件处理留在 feature model，只把 WebSocket 连接、状态、重连和心跳复用给 `useWebSocket()`。
- `useWebSocket.ts` 当前已经通过 `getWsTicket()` 和 `handleGlobalSessionExpired()` 承接 shared transport / runtime 语义，没有必要再直接碰 store。

## Decision
refactor_existing

## Reason
`composableMultiBoundaryAllowlist` 只剩 `composables/useWebSocket.ts -> api+store` 这一条。实际代码里 `useWebSocket.ts` 的 `useAuthStore` import 没有被使用，这说明当前 store 边界不是设计要求，而是历史残留。

最小正确改动是：

- 删除 `useWebSocket.ts` 中未使用的 `useAuthStore` import
- 清空 `composableMultiBoundaryAllowlist`
- 同步更新 backlog / plan / review 事实源

本轮不做：

- 不把 `useWebSocket.ts` 下沉到某个 feature
- 不改 `getWsTicket()`、重连、心跳或全局 session expired 行为
- 不顺手处理 `featureRouterImportAllowlist`

## Files to modify
- .harness/reuse-decisions/websocket-composable-boundary-cleanup.md
- docs/plan/impl-plan/2026-05-29-websocket-composable-boundary-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-websocket-composable-boundary-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/composables/useWebSocket.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## After implementation
- `useWebSocket.ts` 继续作为共享 WebSocket transport composable，只保留 `api + runtime` 边界。
- `composableMultiBoundaryAllowlist` 清空。
- 前端 allowlist 残余进一步收敛到 `featureRouterImportAllowlist`。
