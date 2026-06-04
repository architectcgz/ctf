# Reuse Decision

## Change type
frontend refactor / button contract / style owner

## Existing code searched
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkspaceHeader.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyPackageContextPanel.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue
- code/frontend/scripts/vue-deep-allowlist.json

## Similar implementations found
- `workspace-shell.css` 的 `.ui-btn` 已经通过 `--ui-btn-*` 变量控制尺寸、颜色、hover 和 focus。
- `TopologyChallengeWorkspaceHeader.vue` 已经在本组件内声明 `.topology-action-btn` 变量，说明 topology 按钮视觉本身适合走 `ui-btn` contract。
- challenge / template workbench 当前仍通过父级 `:deep(.topology-action-btn*)` 反向设置 descendant button 变量。

## Decision
extend_existing

## Reason
这轮不是新增按钮组件，而是复用 `.ui-btn` 现有 CSS variable contract，把 topology workbench 中按钮变量从 `:deep(.topology-action-btn*)` 迁到 workbench root。

最小正确改动是：

- 在 `TopologyChallengeWorkbench.vue` 和 `TopologyTemplateWorkbench.vue` 根节点声明 `--ui-btn-*` 变量
- 删除 workbench 父壳里的 `:deep(.topology-action-btn*)`
- 对 icon-only 操作按钮，把尺寸 class owner 放回按钮所在子组件
- 同步收缩 `vue-deep-allowlist.json`

本轮不做：

- 不处理 topology 表单控件 `input / select / textarea`
- 不处理 canvas board、node editor 或 context rail 的穿透样式
- 不改按钮文案、禁用逻辑、事件流或业务动作

## Files to modify
- .harness/reuse-decisions/frontend-topology-action-button-contract-convergence.md
- docs/plan/impl-plan/2026-06-04-topology-action-button-contract-convergence-plan.md
- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue
- code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue
- code/frontend/scripts/vue-deep-allowlist.json

## After implementation
- topology workbench 不再通过 `:deep(.topology-action-btn*)` 设置按钮样式。
- 按钮视觉继续由 `.ui-btn` 的公开变量 contract 承担。
