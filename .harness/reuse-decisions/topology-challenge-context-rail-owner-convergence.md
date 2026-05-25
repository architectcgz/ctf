# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/components/platform/topology/TopologyPackageContextPanel.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/platform/topology/TopologyPackageContextPanel.vue`

## Decision
refactor_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前剩余的高复杂度模板主要是 challenge 模式右侧 `context rail` 的编排壳。状态说明、题包上下文、模板侧栏都已经各自有稳定 owner，因此下一刀最合理的是把这三块组合壳抽成独立组件，而不是继续下沉导出、模板载入或页面级数据 owner。父页继续持有 reload、导出题包、模板搜索/应用/删除和所有 draft 编辑动作。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续压缩 `ChallengeTopologyStudioPage.vue`，补 `.harness/reuse-index/` 镜像索引，记录“page owner + composed context rail shell” 模式。
