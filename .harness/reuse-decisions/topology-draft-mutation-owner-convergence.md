# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologySelectionState.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`

## Similar implementations found
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyEdgeEditing.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologySelectionState.ts`

## Decision
extend_existing

## Reason
`ChallengeTopologyStudioPage.vue` 当前剩余最成组的脚本 owner 已经集中在本地 `draft` 变更 helper：网络、节点、入口节点、链路和策略的更新/删除都还在页面里直接改 `draft.value`。这些逻辑不依赖模板结构，也不应该继续留在页面 owner。现有 `useTopologyStructureMutations.ts` 已经承接拓扑结构的增删动作，最小安全切片是继续扩这个 composable，把页面本地的 `draft` 变更 helper 一并收进去，再由 `useChallengeTopologyStudioPage.ts` 统一对外暴露。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这组 draft 变更 owner 收口后，拓扑页只剩模板装配和样式大块，则 `TD-1` 可以继续转成更小的局部 presentation / 样式尾项，而不是继续留着页面内状态变更脚本。
