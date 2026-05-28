# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/index.ts`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/components/contests/awd/__tests__/AWDDefenseConnectionPanel.test.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `code/frontend/src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `challenge-detail` 刚完成的 `ChallengeWorkspaceShell.vue` / `ChallengeSolutionsPanel.vue` / `ChallengeSubmissionRecordsPanel.vue` 迁位说明，单一 feature 私有 UI cluster 应整体迁入 `features/*/ui`，并同步清掉 allowlist 与 raw-source 护栏。
- `contest-projector`、`contest-awd-admin` 最近几轮都已把单一 capability UI cluster 从旧 `components/platform/*` 收回到 feature `ui`，说明当前迁移基线已经接受“page / workspace shell 留原 owner，稳定子面板迁 feature”的模式。
- `ContestAWDWorkspacePanel.vue` 当前已经通过 `useContestAWDWorkspace()`、`useAwdDefenseAccessPanel()`、`useAwdDefenseServiceSelection()` 持有防守区 workflow owner，`AWDDefense*` 三件套只承接稳定装配和展示，不需要继续挂在 legacy contest 组件目录。

## Decision
refactor_existing

## Reason
- `AWDDefenseColumn.vue`、`AWDDefenseAlertsPanel.vue`、`AWDDefenseOperationsPanel.vue`、`AWDDefenseConnectionPanel.vue`、`AWDDefenseServiceList.vue` 只服务 `contest-awd-workspace` 这一条 feature，是典型的 feature 私有 UI。
- 继续留在 `components/contests/awd/*` 会让 `componentFeatureImportAllowlist` 保留 3 条明显的历史例外。
- 最小正确改动是把整个 defense cluster 一起迁入 `features/contest-awd-workspace/ui`，并让 `ContestAWDWorkspacePanel.vue` 改为 feature 内部相对 import；本轮不顺手动 `AWDAttackVectorPanel.vue`、`AWDWorkspaceHudStrip.vue`、`AWDWorkspaceIntelColumn.vue`。

## Files to modify
- `.harness/reuse-decisions/contest-awd-workspace-defense-cluster-feature-ui-batch.md`
- `docs/plan/impl-plan/2026-05-28-contest-awd-workspace-defense-cluster-feature-ui-batch-plan.md`
- `docs/reviews/frontend/2026-05-28-contest-awd-workspace-defense-cluster-feature-ui-batch-review.md`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/features/contest-awd-workspace/ui/index.ts`
- `code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/AWDDefenseColumn.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/AWDDefenseConnectionPanel.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/AWDDefenseServiceList.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/features/contest-awd-workspace/ui/__tests__/AWDDefenseConnectionPanel.test.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `code/frontend/src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `contest-awd-workspace` 对应的 3 条 `componentFeatureImportAllowlist` 应该清空。
- `ContestAWDWorkspacePanel.vue` 不再直连旧 `components/contests/awd/AWDDefense*.vue` 路径。
- 剩余 `componentFeatureImportAllowlist` 应主要收敛到 `layout` 与 `challenge-topology-studio model consumer` 这类不再属于单纯 feature 私有 UI 落位错误的例外。
