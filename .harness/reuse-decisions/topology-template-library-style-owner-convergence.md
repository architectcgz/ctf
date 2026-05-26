# Reuse Decision

## Change type
style_refactor

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前剩余的大块 `TD-1` 尾项主要是 template-library 模式样式仍集中在父页：header、hero、workbench 和局部表单/按钮皮肤都已经有明确组件 owner，却还通过父页 scoped CSS 做跨组件覆盖。下一刀应把这些样式收回到 `TopologyTemplateLibraryHeader.vue`、`TopologyTemplateHeroSection.vue`、`TopologyTemplateWorkbench.vue`，让父页只保留页面壳、主题变量和极少量 section 容器样式。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩拓扑页，优先处理 challenge / template 共用的通用视觉 token 是否还需要页面级 shared wrapper。
