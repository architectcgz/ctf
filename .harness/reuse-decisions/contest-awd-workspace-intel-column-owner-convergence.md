# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`
- `code/frontend/src/components/platform/topology/TopologyPackageContextPanel.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 已经按区块抽出过防守相关子组件，当前最小安全收口路径是继续沿稳定展示区块下沉 owner。右侧 intelligence rail 只消费现成的 scoreboard 和 recent events，不拥有路由、请求、提交、复制或服务动作，适合抽成独立子组件，而不是新建 feature 或改 page model。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续 AWD 学员战场继续沿区块抽层，补 `.harness/reuse-index/` 对应镜像索引，记录“war-room panel -> stable section component” 模式。
