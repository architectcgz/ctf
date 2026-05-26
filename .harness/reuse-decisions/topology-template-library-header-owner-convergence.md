# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 在 template-library 模式下剩余最稳定的父页模板块是顶部 `PageHeader` 与“新建空白模板 / 刷新”动作条。它们只消费现成文案并向父页透传页面动作，不持有 draft、远端请求或路由状态，适合继续抽成 `TopologyTemplateLibraryHeader.vue`，让父页只保留模式切换与页面级 owner。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩拓扑页，补 `.harness/reuse-index/` 镜像索引，记录“template library header shell” 模式。
