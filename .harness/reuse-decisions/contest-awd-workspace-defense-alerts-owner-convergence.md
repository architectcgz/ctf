# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 已经连续按稳定展示区块抽到 `awd/*` 子组件。左侧顶部的 `defenseAlerts` 只消费父组件现成的 `defenseAlerts` 计算结果，不直接持有服务选择、SSH、重启、复制或刷新逻辑，是继续收口当前 touched surface 的最小安全切片。现有 `AWDDefenseServiceList.vue` 和 `AWDDefenseOperationsPanel.vue` 已承担防守动作 owner，因此本轮只补一个纯展示告警块即可，不需要新建 feature 或调整 composable。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续沿 AWD war-room 区块拆层，补 `.harness/reuse-index/` 镜像索引，记录“left rail display block -> awd child component” 模式。
