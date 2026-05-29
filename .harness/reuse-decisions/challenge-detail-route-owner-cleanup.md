# Reuse Decision

## Change type
frontend refactor / challenge detail route owner cleanup

## Existing code searched
- code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts
- code/frontend/src/features/challenge-detail/model/index.ts
- code/frontend/src/views/challenges/ChallengeDetail.vue
- code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/features/challenge-list/model/challengeListRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeNavigationTransport.ts`
- `composables/routeQueryTransport.ts`
- `features/platform-challenge-detail/model/platformChallengeDetailRoutes.ts`
- `features/challenge-list/model/challengeListRoutes.ts`

## Decision
refactor_existing

## Reason
`useChallengeDetailPage.ts` 里直接碰 router 的职责只有两块：

- 读取当前 challenge 的 route params
- 在错误态返回题目列表

这层 route 语义仍然是 challenge detail page 自己的 owner，不需要再起一层新的 route wrapper。更合适的收口方式是：

- `challengeId` 改由共享 `routeQueryTransport` 读取
- 返回题目列表改走本地 `challengeDetailRoutes.ts` + `routeNavigationTransport`

这样能清掉 allowlist，同时不碰题目加载、题解 / 提交记录 / 题解编写预取、实例 workflow 和 page-level tab owner。

## Files to modify
- .harness/reuse-decisions/challenge-detail-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-challenge-detail-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-challenge-detail-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/challenge-detail/model/challengeDetailRoutes.ts
- code/frontend/src/features/challenge-detail/model/index.ts
- code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts
- code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts

## After implementation
- `useChallengeDetailPage.ts` 不再 import `vue-router`
- `challengeId` 改由 `routeQueryTransport` 读取
- 返回题目列表改走本地 route target helper + shared navigation transport
- `featureRouterImportAllowlist` 再收掉 `features/challenge-detail/model/useChallengeDetailPage.ts -> vue-router`
