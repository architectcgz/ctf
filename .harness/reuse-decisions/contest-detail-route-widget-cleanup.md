# Reuse Decision

## Change type
page / component

## Existing code searched
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/features/contest-detail/*`

## Similar implementations found
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`

## Decision
refactor_existing

## Reason
竞赛详情页已经有稳定的 route/page owner：`useContestDetailRoutePage.ts`。这次不新建页签、队伍、公告或 AWD 流程逻辑，只复用现有 route model，把 route page 上过重的 loading / 空态 / 页签 / 主工作区壳层下沉到 widget，让 `pages` 层回到纯组合入口。

## Files to modify
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/widgets/contest-detail-workspace/index.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-contest-detail-route-widget-cleanup-plan.md`
- `.harness/reuse-decisions/contest-detail-route-widget-cleanup.md`

## After implementation
- `ContestDetailRoutePage.vue` 只保留 route model 调用和 widget 组合。
- 竞赛详情页壳层迁到 `widgets/contest-detail-workspace`。
- 迁移台账不再把 `ContestDetailRoutePage.vue` 记为当前优先瘦身对象。
