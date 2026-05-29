# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts
- code/frontend/src/views/platform/ContestOperationsHub.vue
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue
- code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/platform-contests/model/contestManageRoutes.ts`
- `features/teacher-dashboard/model/teacherDashboardRoutes.ts`
- `features/contest-detail/model/contestListRoutes.ts`

## Decision
refactor_existing

## Reason
`useContestOperationsHubPage.ts` 当前的主 owner 是 AWD 赛事目录加载、preferred contest 选择、分页和加载错误态。额外保留的两条导航：

- 进入单场运维台
- 返回竞赛目录

都只是薄路由动作，不再值得继续把 `useRouter()` 留在 page model 里。当前仓库对这类债的主模式已经明确：

- page model 保留数据、分页、筛选和 workflow owner
- 单独新增 route target helper
- view / feature UI 直接通过 `AppRouteLink` 消费目标路由

因此这轮最小正确改动是把 contest operations hub 的返回与进入动作收口成显式 route target contract，而不是继续让 `useContestOperationsHubPage.ts` 直接 import `vue-router`。

## Files to modify
- .harness/reuse-decisions/contest-operations-hub-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-contest-operations-hub-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-contest-operations-hub-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts
- code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts
- code/frontend/src/features/platform-contests/model/index.ts
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue
- code/frontend/src/views/platform/ContestOperationsHub.vue
- code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts

## After implementation
- `useContestOperationsHubPage.ts` 不再 import `vue-router`
- 赛事运维目录的返回 / 进入运维台入口改为显式 route target contract
- hero、空态和目录行直接通过 `AppRouteLink` 消费赛事运维目录路由
- `featureRouterImportAllowlist` 收掉 `useContestOperationsHubPage.ts`
