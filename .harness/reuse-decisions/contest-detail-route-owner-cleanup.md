# Reuse Decision

## Change type
frontend refactor / contest detail route owner cleanup

## Existing code searched
- code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts
- code/frontend/src/features/contest-detail/model/useContestDetailPage.ts
- code/frontend/src/views/contests/ContestDetail.vue
- code/frontend/src/views/contests/__tests__/ContestDetail.test.ts
- code/frontend/src/composables/useRouteQueryTabs.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/useRouteQueryTabs.ts`
- `composables/routeQueryTransport.ts`
- `features/challenge-list/model/useChallengeListPage.ts`
- `features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts`

## Decision
refactor_existing

## Reason
`useContestDetailRoutePage.ts` 当前同时碰了三层 route surface：

- `contestId` route param
- `challenge / panel` query 读取与写回
- workspace tab state owner

如果新建 wrapper 或把这层平移到 route view，只是换壳，不是真正收口；如果整条改成 route props，又会连带重做现有测试装配。更小也更合理的做法是：

- 继续把 `useContestDetailRoutePage.ts` 留作 contest detail 的 route-aware page owner
- `panel` tab state 改为复用共享 `useRouteQueryTabs()`
- `params / query / replaceQuery` 下沉到共享 transport，让 feature page model 不再直接 import `vue-router`

这样能真实减少 `featureRouterImportAllowlist`，同时保持 challenge query sync、AWD 默认页签和 contest detail 派生状态仍由 page owner 自己负责。

## Files to modify
- .harness/reuse-decisions/contest-detail-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-detail-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-detail-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts
- code/frontend/src/views/contests/ContestDetail.vue
- code/frontend/src/views/contests/__tests__/ContestDetail.test.ts

## After implementation
- `useContestDetailRoutePage.ts` 不再 import `vue-router`
- contest detail 的 tab/query owner 继续留在 page model
- AWD 默认页签与 challenge query sync 继续工作，但 router transport 改为共享 composable
- `featureRouterImportAllowlist` 再收掉 `features/contest-detail/model/useContestDetailRoutePage.ts -> vue-router`
