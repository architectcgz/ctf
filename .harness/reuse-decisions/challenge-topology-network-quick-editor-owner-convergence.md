# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`

## Decision
refactor_existing

## Reason
- 拓扑页已经连续通过 `topology/*` 子组件分层收口 debt，本轮继续沿同一目录和 owner 方向推进，比重新调整 page model 风险更低。
- 当前 challenge-only 的“网络快速编辑”区块只消费 `draft.networks`，并且只改 `key / name / internal` 三个局部字段；它不拥有远端请求、模板工作流、选中态或保存动作，是典型的局部展示 / 编辑区块。
- 因此本轮应让父页继续持有 `draft` 与 `updateNetworkDraft`，新组件只通过 `props.networks + emit(update-network)` 承接快速编辑模板，而不是把 `draft` 或页面级方法下沉到子组件。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果 challenge-only 的 quick editor 收口模式继续稳定，可把 “父页持有 topology draft / 子组件承接局部 quick editor” 作为 topology debt 收口线索补进 `.harness/reuse-index/`。
