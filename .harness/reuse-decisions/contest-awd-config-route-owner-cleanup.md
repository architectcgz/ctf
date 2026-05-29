# Reuse Decision

## Change type
frontend refactor / contest awd config route owner cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts
- code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts
- code/frontend/src/features/contest-awd-config/model/index.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeNavigationTransport.ts`
- `composables/routeQueryTransport.ts`
- `features/platform-challenge-detail/model/platformChallengeDetailRoutes.ts`

## Decision
refactor_existing

## Reason
`useContestAwdConfigPage.ts` 当前直接持有的 router 语义比较薄：

- 读取 `contestId` params
- 读取 / 写回 `service` query
- 返回 `ContestEdit?panel=awd-config`

这些都属于 AWD 配置页自己的 route owner，不需要新建 feature 外 wrapper。更合理的收口方式是：

- `contestId` 和 `service` query 读取改走共享 `routeQueryTransport`
- query 写回直接复用共享 `replaceQuery()`
- 返回赛事工作台改成 detail feature 自己的本地 route target helper + `routeNavigationTransport`

这样能拿掉 allowlist，同时保持 AWD 配置页自己的 mounted 初始化、breadcrumb 和 checker workflow 仍留在 page owner。

## Files to modify
- .harness/reuse-decisions/contest-awd-config-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-awd-config-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-awd-config-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/contest-awd-config/model/contestAwdConfigRoutes.ts
- code/frontend/src/features/contest-awd-config/model/index.ts
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts

## After implementation
- `useContestAwdConfigPage.ts` 不再 import `vue-router`
- `contestId` / `service` query 改由 `routeQueryTransport` 读取
- `service` query 写回改为共享 `replaceQuery()`
- 返回赛事工作台改走本地 route target helper + `routeNavigationTransport`
- `featureRouterImportAllowlist` 再收掉 `features/contest-awd-config/model/useContestAwdConfigPage.ts -> vue-router`
