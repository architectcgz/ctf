# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 现在剩余的主要大块就是左侧防守列的编排壳。告警、服务列表、操作面板都已经各自有明确子组件 owner，因此下一刀最合理的方式不是再拆动作，而是把这三块的装配布局抽成 `AWDDefenseColumn.vue`。父组件继续持有服务选择、SSH、复制、刷新和重启动作 owner，只把布局壳下沉。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩 `ContestAWDWorkspacePanel.vue`，补 `.harness/reuse-index/` 镜像索引，记录“page owner + composed rail component” 模式。
