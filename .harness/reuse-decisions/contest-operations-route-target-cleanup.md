# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/platform-contests/model/useContestOperationsPage.ts
- code/frontend/src/views/platform/ContestOperations.vue
- code/frontend/src/views/platform/__tests__/ContestOperations.test.ts
- code/frontend/src/router/routes/platformRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/platform-contests/model/contestManageRoutes.ts`
- `features/platform-contests/model/useContestAnnouncementsPage.ts`
- `views/platform/ContestAnnouncements.vue`

## Decision
refactor_existing

## Reason
`useContestOperationsPage.ts` 当前的主 owner 是单场 AWD 运维页的竞赛详情加载、breadcrumb detail 标题维护，以及 runtime/readiness 面板模式判定。额外保留的 `useRoute()` 只是在读单个 `contestId`。

这类输入不值得继续把 `vue-router` 留在 feature model 里。当前仓库对这类债的主模式已经明确：

- route record 负责把路径参数显式下沉成 props
- route view 保持薄组合，只把 route props 转成 `Ref`
- feature model 继续保留数据、async workflow 和页面级派生状态 owner

因此这轮最小正确改动是把 contest operations 的 `contestId` owner 收口成 route props contract，而不是继续让 `useContestOperationsPage.ts` 直接 import `vue-router`。

## Files to modify
- .harness/reuse-decisions/contest-operations-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-operations-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-operations-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-contests/model/useContestOperationsPage.ts
- code/frontend/src/views/platform/ContestOperations.vue
- code/frontend/src/views/platform/__tests__/ContestOperations.test.ts
- code/frontend/src/router/routes/platformRoutes.ts

## After implementation
- `useContestOperationsPage.ts` 不再 import `vue-router`
- `ContestOperations` route 显式把 `contestId` 作为 props 下传
- `ContestOperations.vue` 只负责 route props -> page model 的薄组合
- `featureRouterImportAllowlist` 收掉 `useContestOperationsPage.ts`
