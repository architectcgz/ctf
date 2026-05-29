# Reuse Decision

## Change type
frontend refactor / route owner cleanup batch

## Existing code searched
- code/frontend/src/features/notifications/model/useNotificationDetailPage.ts
- code/frontend/src/views/notifications/NotificationDetail.vue
- code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts
- code/frontend/src/features/scoreboard/model/useScoreboardRoutePage.ts
- code/frontend/src/views/scoreboard/ScoreboardView.vue
- code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts
- code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts
- code/frontend/src/composables/useRouteQueryTabs.ts
- code/frontend/src/router/routes/studentRoutes.ts
- code/frontend/src/components/navigation/routeTarget.ts
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `features/challenge-package-import/model/challengeImportRoutes.ts`
- `features/platform-contests/model/contestManageRoutes.ts`
- `components/navigation/AppRouteLink.vue`
- `components/navigation/AppRouteRedirect.vue`

## Decision
refactor_existing

## Reason
这轮目标是继续压缩 `featureRouterImportAllowlist`，但不引入新的中间迁移态。

- `useNotificationDetailPage.ts` 当前同时持有 route param 输入、返回通知列表导航和关联对象跳转，属于典型的 route-aware page owner，适合改成：
  - route param 通过 route props 下沉
  - 站内导航改成显式 route target contract
  - 外链继续留在 page model 内作为 window transport
- `useScoreboardRoutePage.ts` 当前只是把 `route/router` 原样透传给 `useRouteQueryTabs()`。如果只把 `useRoute/useRouter` 平移到别的 feature wrapper，本质仍是隐式 router contract。更合理的收口方式是直接让 `useRouteQueryTabs()` 成为共享 query-tab route owner，由它自己承接 `useRoute/useRouter`，feature route page 只声明 tab 语义。
- `usePlatformChallengeDetailRoutePage.ts` 与 `useScoreboardRoutePage.ts` 是同一类 query-tab pass-through；既然这轮已经收口 `useRouteQueryTabs()` 的 owner，也应顺手把这条同构 allowlist 一并清掉，避免保留半迁移状态。

`useChallengeListPage.ts` 虽然也在 allowlist 里，但它同时混有 query/filter sync 与三条薄导航。这条如果和本轮一起做，容易把 slice 扩大成 query owner 重构，因此先明确留到下一批单独处理。

## Files to modify
- .harness/reuse-decisions/feature-router-second-batch-cleanup.md
- docs/plan/impl-plan/2026-05-29-feature-router-second-batch-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-feature-router-second-batch-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/composables/useRouteQueryTabs.ts
- code/frontend/src/router/routes/studentRoutes.ts
- code/frontend/src/features/notifications/model/useNotificationDetailPage.ts
- code/frontend/src/features/notifications/model/index.ts
- code/frontend/src/features/scoreboard/model/useScoreboardRoutePage.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts
- code/frontend/src/views/notifications/NotificationDetail.vue
- code/frontend/src/views/scoreboard/ScoreboardView.vue
- code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts
- code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts
- code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts

## After implementation
- `useNotificationDetailPage.ts` 不再 import `vue-router`
- `useScoreboardRoutePage.ts` 不再 import `vue-router`
- `usePlatformChallengeDetailRoutePage.ts` 不再 import `vue-router`
- `useRouteQueryTabs.ts` 成为共享 query-tab route owner
- `featureRouterImportAllowlist` 至少再收掉上述三条
