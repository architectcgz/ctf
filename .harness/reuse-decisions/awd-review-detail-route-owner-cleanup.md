# Reuse Decision

## Change type
frontend refactor / awd review detail route owner cleanup

## Existing code searched
- code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts
- code/frontend/src/features/awd-review-detail-workspace/model/index.ts
- code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/utils/teachingWorkspaceRouting.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeNavigationTransport.ts`
- `composables/routeQueryTransport.ts`
- `features/contest-awd-config/model/contestAwdConfigRoutes.ts`

## Decision
refactor_existing

## Reason
`useAwdReviewDetailPage.ts` 当前直接持有的 router 语义主要是：

- 读取 `contestId` params 与 `round` query
- 切换 round 时写回当前页 query
- 返回 AWD 复盘目录

这些都还是 AWD 复盘详情页自己的 route owner，没有必要为了消 allowlist 再造一层 route wrapper。更合理的收口方式是：

- `contestId` / `round` 读取下沉到共享 `routeQueryTransport`
- round 写回改为共享 `replaceQuery()`
- 返回列表导航收口到本地 `awdReviewDetailRoutes.ts` + `routeNavigationTransport`

这样能拿掉 allowlist，同时保持 review 加载、team drawer、export polling 和 breadcrumb owner 继续留在 page model。

## Files to modify
- .harness/reuse-decisions/awd-review-detail-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-awd-review-detail-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-awd-review-detail-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/awd-review-detail-workspace/model/awdReviewDetailRoutes.ts
- code/frontend/src/features/awd-review-detail-workspace/model/index.ts
- code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts
- code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts

## After implementation
- `useAwdReviewDetailPage.ts` 不再 import `vue-router`
- `contestId` / `round` 改由 `routeQueryTransport` 读取
- `setRound()` 改由共享 `replaceQuery()` 写回
- `openReviewIndex()` 改走本地 route target helper + shared navigation transport
- `featureRouterImportAllowlist` 再收掉 `features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts -> vue-router`
