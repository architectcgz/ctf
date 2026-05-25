# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyEntryNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/topology/TopologyEntryNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`

## Decision
refactor_existing

## Reason
- `ChallengeTopologyStudioPage.vue` 的 challenge 右侧 context rail 里，“题包来源 / 题包文件 / 修订历史”三张卡片已经形成稳定展示簇，只有导出按钮会回到父页触发动作。
- 这组卡片只消费 `packageSourceSummary`、`packageBaselineSummary`、`packageFiles`、`packageRevisionHistory`、`exporting` 和导出事件，不拥有 `draft`、画布选中态、模板写回或保存逻辑。
- 因此本轮应沿用 topology 子组件切片模式，让父页继续保留导出动作 owner，新组件只承接题包上下文卡片模板和 `exportPackage` emit。

## Files to modify
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyPackageContextPanel.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果拓扑页 challenge-only 的 context rail 能继续按“父页 owner + 上下文展示组件”稳定收口，可把这条模式补进 `.harness/reuse-index/`。
