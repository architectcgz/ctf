# Topology Action Button Contract Convergence Plan

> 状态：Current
> 事实源：`.ui-btn` shared contract、topology workbench consumer、`:deep` guard allowlist

## Objective

- 删除 topology workbench 中针对 `.topology-action-btn*` 的深度样式覆盖。
- 复用 `.ui-btn` 的公开 CSS variable contract，让 workbench root 负责按钮视觉变量。
- 保持 challenge / template 两个 topology workbench 的按钮尺寸、颜色、hover 和 focus 行为不变。

## Non-goals

- 不处理 `input / select / textarea` 表单控件穿透。
- 不处理 `data-node-editor`、`topology-canvas-board` 或 `context rail` 相关穿透。
- 不改按钮的事件、禁用条件或文案。

## Source Inputs

- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue`
- `code/frontend/scripts/vue-deep-allowlist.json`

## Architecture Fit Check

- `.ui-btn` 已经是按钮样式的事实 owner，支持 `--ui-btn-*` 变量。
- 当前 debt 是 workbench 父壳通过 `:deep(.topology-action-btn*)` 给 descendant buttons 设置变量；这部分应收口成 root-level 变量继承。
- icon 按钮的最小宽度不是 `.ui-btn` 变量，因此应回到按钮所在组件的本地 class owner。

## Task Breakdown

- [ ] Step 1: 在 `TopologyChallengeWorkbench.vue` root 声明 challenge topology 按钮变量。
- [ ] Step 2: 在 `TopologyTemplateWorkbench.vue` root 声明 template topology 按钮变量。
- [ ] Step 3: 将 icon-only 按钮尺寸放回 `TopologyNetworkSection.vue` / `TopologyConnectivitySections.vue` 本地样式。
- [ ] Step 4: 删除对应 `:deep(.topology-action-btn*)` allowlist 条目。

## Validation

- `cd code/frontend && npm run check:vue-deep`
- `cd code/frontend && npm run test:run -- src/features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.test.ts src/features/challenge-topology-studio/model/topologyStudioBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue code/frontend/scripts/vue-deep-allowlist.json .harness/reuse-decisions/frontend-topology-action-button-contract-convergence.md docs/plan/archive/impl-plan/2026-06/2026-06-04-topology-action-button-contract-convergence-plan.md`
