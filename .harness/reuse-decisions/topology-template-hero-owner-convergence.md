# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/components/platform/topology/TopologySummaryGrid.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前 template-library 模式里剩余最明显的稳定模板块是顶部 hero 区。它只组合了 kicker、标题、副标题、摘要和 `TopologyStatusNotes`，没有页面级动作 owner，因此适合抽成 `TopologyTemplateHeroSection.vue`。父页继续持有 `hero*` 文案和 summary/status 数据 owner。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩拓扑页，补 `.harness/reuse-index/` 镜像索引，记录“mode-specific hero shell” 模式。
