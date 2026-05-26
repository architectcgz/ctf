# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/scoreboard/ScoreboardView.vue`
- `code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`

## Decision
refactor_existing

## Reason
`ScoreboardView.vue` 当前 591 行，超出 route view 阈值，但脚本层已经基本只剩 `useScoreboardRoutePage()` 与 `useScoreboardView()` 的数据 owner、筛选 owner 和分页 owner。主要重量仍在顶部 tab rail、contest / points 两个面板模板，以及本页局部布局样式。最小安全切片是延续已有的 workspace shell 模式：父页保留 route/query、筛选、加载、刷新和数据 owner，新子组件只承接稳定的展示壳层。

## Files to modify
- `code/frontend/src/views/scoreboard/ScoreboardView.vue`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`
- `code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀后 `ScoreboardView.vue` 仍超阈值，下一步优先继续判断 contest 目录面板或 points table 面板是否需要各自独立，而不是把 route/filter owner 再拆散。
