# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前 template-library 模式里剩余最大的稳定模板块是 `topology-workbench`。标签切换、画布区、节点/网络/策略区和模板侧栏都已经各自有明确 owner 或子组件，因此下一刀应把这块装配壳抽成 `TopologyTemplateWorkbench.vue`，而不是继续把页面级数据、导出、保存或 route owner 下沉。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩拓扑页，补 `.harness/reuse-index/` 镜像索引，记录“page owner + mode-specific workbench shell” 模式。
