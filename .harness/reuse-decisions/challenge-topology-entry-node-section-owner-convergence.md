# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`

## Decision
refactor_existing

## Reason
- `ChallengeTopologyStudioPage.vue` 里“入口节点”卡片在模板库模式和挑战模式各有一份，结构稳定，只有删除按钮显隐不同，说明它已经是清晰的展示 / 局部交互区块。
- 当前区块只消费 `draft.entry_node_key`、`nodeOptions`、`saving`、`topology` 和删除动作，不拥有远端请求、模板写回、画布选中态或页面布局 owner。
- 因此本轮应沿用既有 topology 子组件分层模式，让父页继续持有 `draft` 和删除动作，新组件只承接 card 模板和事件发射。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyEntryNodeSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果 topology 页剩余 card 都沿用“父页 owner + section 组件承接稳定模板”的方式收口，可把这条模式补进 `.harness/reuse-index/`。
