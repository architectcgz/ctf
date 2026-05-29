# Reuse Decision

## Change type
frontend refactor / platform challenge detail route owner cleanup

## Existing code searched
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts
- code/frontend/src/features/platform-challenge-detail/model/index.ts
- code/frontend/src/features/platform-challenges/model/platformChallengeRoutes.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts
- code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeNavigationTransport.ts`
- `composables/routeQueryTransport.ts`
- `features/platform-challenges/model/platformChallengeRoutes.ts`

## Decision
refactor_existing

## Reason
`usePlatformChallengeDetailPage.ts` 当前直接持有三类 router 职责：

- 读取 `challengeId` route param
- 返回题库、进入拓扑、进入题解查看/编辑的薄导航
- 加载失败后的延迟跳回题库

这些职责都还是 challenge detail page owner 自己的范围，不需要再造一层 feature 外 wrapper；但继续直连 `vue-router` 又会让 `featureRouterImportAllowlist` 停在 detail 这一条。更合理的收口方式是：

- `challengeId` 读取下沉到共享 `routeQueryTransport`
- `push()` 下沉到共享 `routeNavigationTransport`
- challenge detail 自己的“返回题库 / 去拓扑 / 去题解”目标路由落到本地 `platformChallengeDetailRoutes.ts`

这样既能消掉 allowlist，又不会把 detail 页的失败重定向和动作语义错误地平移到 shared。

## Files to modify
- .harness/reuse-decisions/platform-challenge-detail-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-platform-challenge-detail-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-platform-challenge-detail-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-challenge-detail/model/index.ts
- code/frontend/src/features/platform-challenge-detail/model/platformChallengeDetailRoutes.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts
- code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts

## After implementation
- `usePlatformChallengeDetailPage.ts` 不再 import `vue-router`
- `challengeId` 由 `routeQueryTransport` 读取
- 返回题库、拓扑、题解查看/编辑和失败重定向都走本地 route target helper + shared navigation transport
- `featureRouterImportAllowlist` 再收掉 `features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts -> vue-router`
