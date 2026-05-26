# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/components/contests/ContestOverviewPanel.vue`
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`
- `code/frontend/src/components/contests/ContestAnnouncementsPanel.vue`
- `code/frontend/src/components/contests/ContestTeamPanel.vue`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/views/contests/__tests__/contestDetailPanelExtraction.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/ContestOverviewPanel.vue`
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`

## Decision
refactor_existing

## Reason
`ContestDetail.vue` 当前剩余的重量主要来自三个还留在路由页里的壳层：公告 tab section、队伍 tab section，以及创建/加入队伍两个 `CFocusedInputDialog`。这些部分已经有明确的 page owner 边界：路由页继续持有 `activeWorkspaceTab`、远端数据、队伍动作和路由同步，子组件只承接 section / dialog 模板与局部事件转发。因此本轮应继续沿用既有的 `父页 owner + 子组件壳层` 模式，把剩余壳层收回到 `components/contests`。

## Files to modify
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/components/contests/ContestAnnouncementsWorkspaceSection.vue`
- `code/frontend/src/components/contests/ContestTeamWorkspaceSection.vue`
- `code/frontend/src/components/contests/ContestTeamDialogs.vue`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/views/contests/__tests__/contestDetailPanelExtraction.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀完成后 `ContestDetail.vue` 仍明显偏重，下一步优先判断是否要把顶层 `workspace-tabbar` 和 `loading / empty` 壳继续抽成稳定的 contests workspace shell，而不是把 `useContestDetailRoutePage` 再切碎。
