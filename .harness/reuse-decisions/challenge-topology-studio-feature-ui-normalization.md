# Reuse Decision

## Change type
frontend refactor / feature-owned topology studio UI normalization

## Existing code searched
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.vue
- code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts
- code/frontend/src/components/platform/topology/*.vue
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts

## Similar implementations found
- 最近几轮 `challenge-writeup-editor`、`contest-awd-admin`、`layout-shell` 的收口方式，都已经采用“feature-owned UI 回收到 `features/*/ui`，父页只保留 workflow owner，测试改到聚合源码”的模式。
- `challenge-topology-studio` 当前也已经有明确的 `model/` 与 `ui/` 分层，但专属工作台 UI 仍散落在 `components/platform/topology/`。

## Decision
refactor_existing

## Reason
当前 `componentFeatureImportAllowlist` 剩余的 6 条全部来自：

- `components/platform/topology/TopologyCanvasBoard.vue`
- `components/platform/topology/TopologyConnectivitySections.vue`
- `components/platform/topology/TopologyNetworkSection.vue`
- `components/platform/topology/TopologyNodeEditor.vue`
- `components/platform/topology/TopologyNodeSection.vue`
- `components/platform/topology/TopologyTemplateWorkbench.vue`

这不是“共享组件合理依赖 feature model”，而是 `challenge-topology-studio` 自己的专属 UI 还停留在 legacy component 目录。外部真实消费面只剩：

- `ChallengeTopologyStudioPage.vue`
- 对应 raw-source / theme / async chunk 测试

因此最小正确改动不是继续保留 allowlist，也不是增加 bridge，而是：

- 把整组 topology studio 专属 UI 从 `components/platform/topology/` 一次性迁回 `features/challenge-topology-studio/ui/`
- 同步更新 feature page、内部相对依赖、测试 raw-source 路径、`components.d.ts`
- 删除这 6 条 `componentFeatureImportAllowlist`

## Files to modify
- .harness/reuse-decisions/challenge-topology-studio-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-29-challenge-topology-studio-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-29-challenge-topology-studio-feature-ui-normalization-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/challenge-topology-studio/ui/*
- code/frontend/src/components/platform/topology/*
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts

## After implementation
- `challenge-topology-studio` 的专属 UI 与 model 将都落在 feature 自己目录下。
- `components/platform/topology/` 不再承载这组专属工作台 UI，也不再需要为它们保留 `componentFeatureImportAllowlist`。
- 这组拓扑工作台的 raw-source / boundary / theme / async-chunk 护栏会同步切到新 owner，避免留下“代码迁了、测试还盯旧路径”的中间态。
