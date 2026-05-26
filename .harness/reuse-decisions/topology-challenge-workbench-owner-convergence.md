# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/components/platform/topology/TopologyEntryNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前 challenge 模式下剩余最大的稳定模板块是左侧主工作区与右侧 `context rail` 的装配壳。画布、入口节点、网络、节点、策略和上下文轨道都已经是独立子组件，父页继续保留这段大模板只会增加装配噪音，因此应继续抽成 `TopologyChallengeWorkbench.vue`，同时把页面级 owner 仍留在父页。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩拓扑页，补 `.harness/reuse-index/` 镜像索引，记录“challenge mode workbench shell” 模式。
