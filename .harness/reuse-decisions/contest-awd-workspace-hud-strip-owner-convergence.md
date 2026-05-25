# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/platform/topology/TopologySummaryGrid.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 已经沿稳定展示区块拆出过 `awd/*` 子组件，下一刀继续沿这个 owner 收敛路径最稳。顶部 HUD strip 只读父组件已有派生值，并只发出一个刷新意图，不直接持有目标筛选、SSH、服务动作或 Flag 提交流程，适合抽成展示子组件，而不是新建 feature、composable 或共享 KPI shell。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续 AWD 学员战场继续沿区块抽层，补 `.harness/reuse-index/` 对应镜像索引，记录“war-room page owner + stable section component” 模式。
