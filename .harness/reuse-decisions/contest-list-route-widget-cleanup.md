# Reuse Decision

## Change type
page / component

## Existing code searched
- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/features/contest-detail/*`

## Similar implementations found
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`

## Decision
refactor_existing

## Reason
竞赛列表页已经有稳定的 page owner：`useContestListPage.ts`。这次不新建查询、筛选、分页或 route target 逻辑，只复用现有 feature model，把 route page 上过重的摘要卡、筛选区和目录壳层下沉到 widget，让 `pages` 层回到纯组合入口。

## Files to modify
- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/widgets/contest-list-workspace/index.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-contest-list-route-widget-cleanup-plan.md`
- `.harness/reuse-decisions/contest-list-route-widget-cleanup.md`

## After implementation
- `ContestListRoutePage.vue` 只保留 feature model 调用和 widget 组合。
- 竞赛列表页壳层迁到 `widgets/contest-list-workspace`。
- 迁移台账不再把 `ContestListRoutePage.vue` 记为当前优先瘦身对象。
