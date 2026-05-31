# Reuse Decision

## Change type
page / component

## Existing code searched
- `code/frontend/src/pages/scoreboard/ScoreboardDetailRoutePage.vue`
- `code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- `code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`

## Similar implementations found
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`

## Decision
refactor_existing

## Reason
排行详情页已经有稳定的 page owner：`useScoreboardDetailPage.ts`。这次不新建加载、刷新、分页或实时更新逻辑，只复用现有 feature model，把 route page 上过重的头部、概况卡和目录壳层下沉到 widget，让 `pages` 层回到纯组合入口。

## Files to modify
- `code/frontend/src/pages/scoreboard/ScoreboardDetailRoutePage.vue`
- `code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`
- `code/frontend/src/widgets/scoreboard-detail-workspace/index.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-scoreboard-detail-route-widget-cleanup-plan.md`
- `.harness/reuse-decisions/scoreboard-detail-route-widget-cleanup.md`

## After implementation
- `ScoreboardDetailRoutePage.vue` 只保留 feature model 调用和 widget 组合。
- 排行详情页壳层迁到 `widgets/scoreboard-detail-workspace`。
- 迁移台账不再把 `ScoreboardDetailRoutePage.vue` 记为当前优先瘦身对象。
