# Reuse Decision

## Change type
refactor_existing / shared-realtime / docs / test

## Existing code searched
- `code/frontend/src/composables/useWebSocket.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts`
- `code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.ts`
- `code/frontend/src/features/awd-inspector/model/useContestAwdPreviewRealtime.ts`
- `code/frontend/src/shared/ui/layout/NotificationDrawer.vue`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/model/layout/useLayoutNotificationDrawerBridge.ts`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/05-websocket-composables.md`
- `docs/architecture/frontend/08-build-deploy.md`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## Similar implementations found
- `shared/model/navigation/*` 已承接 route-aware runtime owner，说明共享 runtime owner 可以落在 `shared/model/*`
- `shared/model/common/*` 受低层 guardrail 约束，不能直接 import `@/api`、`@/runtime` 或 `@/stores`
- `useWebSocket` 直接依赖 `getWsTicket()` 与 `handleGlobalSessionExpired()`，不适合进入 `shared/model/common` 或 `shared/lib`

## Decision
refactor_existing

## Reason
- `useWebSocket` 是跨 feature 复用的 realtime runtime owner，适合收口到 `shared/model/realtime`
- 新建 `shared/model/realtime` 能避免把 API/runtime 依赖误塞进 low-level `common` / `lib`
- 本批只迁 owner，不改 ticket、心跳、pong timeout、重连或 auth close 语义

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch8-websocket-runtime.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch8-websocket-runtime-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/05-websocket-composables.md`
- `docs/architecture/frontend/08-build-deploy.md`
- `code/frontend/src/composables/useWebSocket.ts`
- `code/frontend/src/shared/model/realtime/useWebSocket.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts`
- `code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.ts`
- `code/frontend/src/features/awd-inspector/model/useContestAwdPreviewRealtime.ts`
- `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- `code/frontend/src/shared/ui/layout/NotificationDrawer.vue`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/model/layout/useLayoutNotificationDrawerBridge.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.test.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.test.ts`
- `code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.test.ts`

## After implementation
- `useWebSocket` 从历史 `src/composables` 收口到 `shared/model/realtime`
- realtime 消费方、类型引用与测试 mock 统一切到共享 realtime owner
- 历史 `src/composables` 不再保留运行时代码
