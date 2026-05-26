# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologySummaryGrid.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前 challenge 模式里剩余最稳定的父页模板块是顶部 `workspace-topbar` 与其后的 heading/summary 区。它们只组合展示文案、统计和四个页面动作按钮，页面级 owner 仍然应该留在父页，因此最合理的是抽成 `TopologyChallengeWorkspaceHeader.vue`，由子组件只承接展示壳和事件透传。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩拓扑页，补 `.harness/reuse-index/` 镜像索引，记录“challenge workspace header shell” 模式。
