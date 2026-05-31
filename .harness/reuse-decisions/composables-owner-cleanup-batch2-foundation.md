# Reuse Decision

## Change type
composition / shared-foundation / docs / test

## Existing code searched
- `code/frontend/src/composables/useAbortController.ts`
- `code/frontend/src/composables/usePagination.ts`
- `code/frontend/src/composables/useRouteQueryTabs.ts`
- `code/frontend/src/composables/useSanitize.ts`
- `code/frontend/src/composables/useTabKeyboardNavigation.ts`
- `code/frontend/src/composables/useUrlSyncedTabs.ts`
- `code/frontend/src/composables/__tests__/useAbortController.test.ts`
- `code/frontend/src/composables/__tests__/usePagination.test.ts`
- `code/frontend/src/composables/__tests__/useSanitize.test.ts`
- `code/frontend/src/features/**`
- `code/frontend/src/entities/**`
- `code/frontend/src/shared/lib/**`
- `code/frontend/src/shared/model/common/**`
- `docs/architecture/frontend/03-state-management.md`

## Similar implementations found
- `shared/lib/overlay/useOverlayBehavior.ts` 已承接纯基础交互行为，说明不带业务语义的 Vue 基础能力适合继续进入 `shared/lib/*`
- `shared/model/common/useToast.ts`、`shared/model/common/useDestructiveConfirm.ts` 已承接共享 UI 状态 owner，说明纯状态 owner 与纯基础工具应继续拆开，不再都挂在 `src/composables`
- `frontend-architecture-policy.json` 已把 `shared/lib/` 归入 common layer，说明基础工具迁到这里不会破坏现有分层策略
- `usePagination` 直接依赖 `@/api/contracts` 的 `PageResult`，更适合作为共享 model contract 层能力进入 `shared/model/common/*`

## Decision
refactor_existing

## Reason
- `useAbortController`、`useSanitize`、`useTabKeyboardNavigation` 没有业务语义，属于共享基础能力，不应长期停留在历史 `src/composables`
- `usePagination` 依赖分页响应 contract，适合落到 `shared/model/common/*`
- 这批能力分别服务请求取消、分页状态、HTML sanitize 和 tab 键盘导航，owner 清晰、迁移风险低，适合作为第二批收口
- 本轮不把 `useWebSocket`、`routeNavigationTransport`、`routeQueryTransport`、`useTheme`、`useReportStatusPolling` 一起卷入，避免把 runtime / router / theme owner 混到同一批

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch2-foundation.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch2-foundation-plan.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `code/frontend/src/composables/useAbortController.ts`
- `code/frontend/src/composables/usePagination.ts`
- `code/frontend/src/composables/useSanitize.ts`
- `code/frontend/src/composables/useTabKeyboardNavigation.ts`
- `code/frontend/src/composables/__tests__/useAbortController.test.ts`
- `code/frontend/src/composables/__tests__/usePagination.test.ts`
- `code/frontend/src/composables/__tests__/useSanitize.test.ts`
- `code/frontend/src/shared/lib/request/useAbortController.ts`
- `code/frontend/src/shared/lib/request/__tests__/useAbortController.test.ts`
- `code/frontend/src/shared/model/common/usePagination.ts`
- `code/frontend/src/shared/model/common/__tests__/usePagination.test.ts`
- `code/frontend/src/shared/lib/sanitize/useSanitize.ts`
- `code/frontend/src/shared/lib/sanitize/__tests__/useSanitize.test.ts`
- `code/frontend/src/shared/lib/keyboard/useTabKeyboardNavigation.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/features/audit-log/model/useAuditLogPage.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/features/challenge-list/model/useChallengeListPage.ts`
- `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.ts`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/features/contest-workbench/model/useContestAwdChallengePicker.ts`
- `code/frontend/src/features/image-management/model/useImageManagePage.ts`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/platform/awd-challenges/model/usePlatformAwdChallenges.ts`
- `code/frontend/src/features/platform/challenges/model/usePlatformChallenges.ts`
- `code/frontend/src/features/platform/contests/model/useContestListState.ts`
- `code/frontend/src/features/platform/contests/model/useContestOperationsHubPage.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUsers.ts`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- `code/frontend/src/features/scoreboard/model/useScoreboardView.ts`
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/entities/challenge/ui/ChallengeDescriptionPanel.vue`
- `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`

## After implementation
- 请求取消、分页、sanitize 和 tab 键盘导航会从历史 `src/composables` 进入共享基础层
- `src/composables` 将进一步缩小到 route/runtime/theme/realtime 等仍需继续分批判断 owner 的能力
