# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasBoard.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`

## Decision
refactor_existing

## Reason
- `ChallengeTopologyStudioPage.vue` 里模板库模式和 challenge 模式都保留了完整的“图形画布”区块，主体结构一致，只在按钮文案、校验提示和 quick editor 布局上有轻差异。
- 该区块消费的核心 owner 仍然是父页持有的 `interactionMode`、selected canvas state、`draftValidationIssues`、`canvasGraph` 和相关更新动作，本身不拥有 route/query、远端请求、模板写回或保存策略。
- 因此本轮应沿用既有 topology 子组件切片模式，让父页继续保留画布交互与草稿 owner，新组件只承接画布工作区模板和事件发射。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果模板库 / challenge 双模态的画布壳能用一套 section 组件稳定覆盖，后续拓扑页剩余 page shell 切片可继续沿用“variant + 父页 owner”模式收口。
