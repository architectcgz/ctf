# Reuse Decision

## Change type
page / component

## Existing code searched
- `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`

## Similar implementations found
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`

## Decision
refactor_existing

## Reason
题目详情页已经有稳定的 page owner：`useChallengeDetailPage.ts`。这次不新建题目加载、题解、提交记录、writeup 或实例 workflow，只复用现有 feature model，把 route page 上过重的 loading / 错误态 / workspace shell 外壳和本地样式下沉到 widget，让 `pages` 层回到纯组合入口。

## Files to modify
- `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
- `code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`
- `code/frontend/src/widgets/challenge-detail-workspace/index.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-challenge-detail-route-widget-cleanup-plan.md`
- `.harness/reuse-decisions/challenge-detail-route-widget-cleanup.md`

## After implementation
- `ChallengeDetailRoutePage.vue` 只保留 feature model 调用和 widget 组合。
- 题目详情页壳层迁到 `widgets/challenge-detail-workspace`。
- 迁移台账不再把 `ChallengeDetailRoutePage.vue` 记为当前优先瘦身对象。
