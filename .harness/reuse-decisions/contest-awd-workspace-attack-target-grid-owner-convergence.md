# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/common/WorkspaceDataTable.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 的中区目前只剩目标卡片列表和提交流程是主要 owner 面。当前最小安全收口方式不是整块搬走，而是把目标卡片列表模板和局部输入壳抽成子组件，同时由父组件继续持有 `flagInputs`、`openingTargetKey`、`submittingKey`、`openTarget`、`handleSubmit` 这些状态与动作 owner。这样可以继续降低大文件体积，但不把攻击业务流拆散。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续拆攻击结果提示或提交 owner，再补 `.harness/reuse-index/` 镜像索引，记录“action list child + parent mutation owner” 模式。
