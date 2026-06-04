# Reuse Decision

## Change type
frontend refactor / component contract / style owner

## Existing code searched
- code/frontend/src/shared/ui/common/SectionCard.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyCanvasWorkspaceSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyEntryNodeSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyNodeSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeContextRail.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyPackageContextPanel.vue
- code/frontend/scripts/vue-deep-allowlist.json

## Similar implementations found
- `frontend-section-card-style-contract-convergence` 已经把 teacher surfaces 从父级 `:deep(.section-card*)` 收口成 `SectionCard` 的显式 contract。
- topology studio 当前仍在 `TopologyChallengeWorkbench.vue`、`TopologyTemplateWorkbench.vue` 和 `TopologyTemplateSidePanel.vue` 里跨组件反向改 `SectionCard` 内部结构。
- `SectionCard.vue` 现在已经有 `variant` 和 CSS variable contract 的基础，可以继续扩成更适合跨页面复用的公共样式入口。

## Decision
extend_existing

## Reason
这轮不是为 topology studio 单独复制一套卡片组件，而是继续扩展 `SectionCard` 的公开样式 contract，把 topology 这条共享 owner 链从 `:deep(.section-card*)` 收口出来。

最小正确改动是：

- 扩展 `SectionCard` 的公开 CSS variable contract，让父级或 consumer 可以不穿透内部类名也能声明 padding / border / header / title 等差异
- 只迁移 topology studio 里与 `.section-card*` 直接相关的 owner 链
- 删除 `TopologyChallengeWorkbench.vue`、`TopologyTemplateWorkbench.vue`、`TopologyTemplateSidePanel.vue` 里对 `.section-card*` 的深度覆盖
- 同步更新 `:deep` allowlist，确认这批 selector 真正退场

本轮不做：

- 不处理 topology studio 里的 `input / select / textarea / data-node-editor / topology-action-btn` 深度样式
- 不处理 modal / drawer / action menu 的 slot-style contract
- 不调整 topology studio 的数据流、tab owner、交互模式或业务文案

## Files to modify
- .harness/reuse-decisions/frontend-topology-section-card-contract-convergence.md
- docs/plan/impl-plan/2026-06-04-topology-section-card-contract-convergence-plan.md
- code/frontend/src/shared/ui/common/SectionCard.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyCanvasWorkspaceSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyEntryNodeSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyNodeSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyPackageContextPanel.vue
- code/frontend/scripts/vue-deep-allowlist.json

## After implementation
- topology studio 不再依赖 `:deep(.section-card*)` 反向修改 `SectionCard` 内部结构。
- `SectionCard` 的公开样式 contract 可以覆盖 topology 这类“整页统一 section rhythm”场景。
- `:deep` allowlist 里与 topology section card owner 链相关的 selector 会缩减到更小范围。
