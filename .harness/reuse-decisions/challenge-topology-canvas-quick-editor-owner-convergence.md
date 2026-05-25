# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/components/platform/topology/TopologySummaryGrid.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`

## Decision
refactor_existing

## Reason
- `ChallengeTopologyStudioPage.vue` 已经连续通过 section / panel 抽层收口模板侧栏、摘要指标、状态说明、网络分段、节点编辑区和链路策略编辑区，说明当前最小风险路径不是新建第二套 page model，而是继续沿用 `topology/*` 子组件分层。
- 当前剩余最稳定且边界清楚的 debt surface 是“画布快速编辑”区块：它消费现成的 `selectedNodeDraft / selectedEdgeMeta / nodeOptions / draft.networks` 和页面 owner 提供的更新动作，但不拥有远端请求、拓扑保存、模板写回、路由或草稿整体序列化。
- 因此本轮应把快速编辑区抽成独立展示组件，由父页面继续持有草稿、选中态、保存动作和 `useChallengeTopologyStudioPage`，子组件只通过 props / emits 承接当前选中对象的局部编辑。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/plan/impl-plan/2026-05-25-challenge-topology-canvas-quick-editor-owner-convergence-implementation-plan.md`

## After implementation
- 如果 `TopologyCanvasQuickEditor.vue` 形成稳定模式，再考虑把 “父页持有 topology draft / 子组件承接局部 quick editor” 线索补进 `.harness/reuse-index/`。
