# Reuse Decision

## Change type
style_refactor

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前剩余的大块 `TD-1` 尾项主要是 challenge 模式的页面壳样式仍留在父页：顶部 header、challenge workbench 布局、右侧轨道布局、section card 和输入皮肤，以及画布区的 allow / validation 状态样式。对应的结构组件已经明确存在，因此下一刀应把样式 owner 收回到 `TopologyChallengeWorkspaceHeader.vue`、`TopologyChallengeWorkbench.vue`、`TopologyCanvasWorkspaceSection.vue`，让父页只保留页面壳变量和极少量模式容器样式。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续还要继续压缩拓扑页，优先检查是否还有仅剩 theme token 定义适合提取为共享 challenge workspace shell token。
