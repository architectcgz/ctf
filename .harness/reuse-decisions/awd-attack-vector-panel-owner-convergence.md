# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- `code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
- `code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 当前剩余最稳定的大块模板是中区 `攻击向量` 装配壳：它只组合标题、筛选条、目标卡片区和提交结果 footer，本身不直接持有远端 workflow。父页仍应保留 challenge 筛选、目标筛选、Flag 输入、打开目标和提交攻击这些 state / async owner，因此最小安全切片是把中区壳层抽成 `AWDAttackVectorPanel.vue`，由子组件只承接展示装配和事件透传。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackVectorPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续收口 AWD 工作台，补 `.harness/reuse-index/` 镜像索引，记录“attack vector shell” 组合模式。
