# Reuse Decision

## Change type
frontend architecture / feature-owned UI migration

## Existing code searched
- code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue
- code/frontend/src/features/contest-awd-workspace/ui/contestAwdWorkspaceUiStrategy.test.ts
- code/frontend/src/features/contest-awd-workspace/ui/index.ts
- code/frontend/src/features/contest-awd-workspace/**
- code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue
- code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue
- code/frontend/src/components/contests/awd/AWDAttackToolbar.vue
- code/frontend/src/components/contests/awd/AWDAttackVectorPanel.vue
- code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue
- code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue
- code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue
- code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `contest-awd-workspace-defense-cluster-feature-ui-batch` 已经证明，`ContestAWDWorkspacePanel.vue` 的稳定 UI cluster 应直接落在 `features/contest-awd-workspace/ui`，而不是继续挂在 `components/contests/awd/*`。
- 本轮剩余的 `AWDAttack*`、`AWDWorkspaceHudStrip`、`AWDWorkspaceIntelColumn` 与 `AWDDefenseFileWorkbench` 也都没有跨 feature consumer，owner 仍然清晰。
- `contest-detail feature-owned shell` 刚完成的迁移模式说明，这类单 consumer 历史壳可以直接整体迁位，并把 raw-source 测试与 `components.d.ts` 一并切过去，不保留桥接层。

## Decision
refactor_existing

## Reason
这轮不是扩展 AWD workspace 功能，而是把还留在 `components/contests/awd/*` 的最后一批单 feature UI 全部收回 `features/contest-awd-workspace/ui`：

- `AWDAttackResultFooter.vue`
- `AWDAttackTargetGrid.vue`
- `AWDAttackToolbar.vue`
- `AWDAttackVectorPanel.vue`
- `AWDDefenseFileWorkbench.vue`
- `AWDWorkspaceHudStrip.vue`
- `AWDWorkspaceIntelColumn.vue`

同时把 `AWDDefenseFileWorkbench` 的邻近测试一并迁入 feature UI 邻近测试目录。

这样可以让：

- `ContestAWDWorkspacePanel.vue` 全面改为 feature 内部相对 import
- `components/contests/awd/*` 退出运行时 owner
- `components.d.ts` 和 AWD workspace raw-source 测试统一对齐新 owner

不做：

- 不改 `contest-awd-workspace/model/*` workflow owner
- 不改 `ContestDetailRoutePage.vue` 或 `ContestAWDWorkspacePanel.vue` 的行为
- 不继续扩展到 platform AWD admin / config 路径

## Files to modify
- .harness/reuse-decisions/contest-awd-workspace-feature-owned-shell-migration-batch2.md
- docs/plan/impl-plan/2026-05-30-contest-awd-workspace-feature-owned-shell-migration-batch2-plan.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components.d.ts
- code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue
- code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue
- code/frontend/src/components/contests/awd/AWDAttackToolbar.vue
- code/frontend/src/components/contests/awd/AWDAttackVectorPanel.vue
- code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue
- code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue
- code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue
- code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts
- code/frontend/src/features/contest-awd-workspace/ui/index.ts
- code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDAttackResultFooter.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDAttackTargetGrid.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDAttackToolbar.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDAttackVectorPanel.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDDefenseFileWorkbench.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDWorkspaceHudStrip.vue
- code/frontend/src/features/contest-awd-workspace/ui/AWDWorkspaceIntelColumn.vue
- code/frontend/src/features/contest-awd-workspace/ui/__tests__/AWDDefenseFileWorkbench.test.ts
- code/frontend/src/features/contest-awd-workspace/ui/contestAwdWorkspaceUiStrategy.test.ts

## After implementation
- `components/contests/awd/*` 将不再保留活动运行时代码 owner。
- `features/contest-awd-workspace/ui` 会成为学生 AWD workspace 这批壳组件的唯一 owner。
- `components/contests` 剩余活动 surface 将不再包含 AWD workspace 这一支。
