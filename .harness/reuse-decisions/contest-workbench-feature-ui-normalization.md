# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/ContestWorkbenchStageTabs.vue
- code/frontend/src/components/platform/contest/ContestWorkbenchSummaryStrip.vue
- code/frontend/src/components/platform/contest/ContestChallengeFilterStrip.vue
- code/frontend/src/components/platform/contest/ContestChallengeSummaryStrip.vue
- code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/components/platform/contest/ContestEditWorkspacePanel.vue
- code/frontend/src/features/contest-workbench/index.ts
- code/frontend/src/features/platform-contests/index.ts
- code/frontend/src/features/platform-contests/model/useContestEditPage.ts
- code/frontend/src/views/platform/ContestEdit.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `platform-challenge-detail`、`platform-awd-challenges`、`contest-announcements` 已按“feature model + feature-owned ui + public API 暴露”的方式收掉单一 feature UI 残片。
- `ContestEditWorkspacePanel.vue` 当前同时依赖 `contest-workbench` 与 `platform-contests`，说明它更像 route shell / feature composition surface，而不是纯 `contest-workbench` 子面板。
- `ContestWorkbenchStageTabs.vue`、`ContestWorkbenchSummaryStrip.vue`、`ContestChallengeFilterStrip.vue`、`ContestChallengeOrchestrationPanel.vue` 都只服务 contest edit workbench，不该继续留在 `components/platform/contest`。

## Decision
refactor_existing

## Reason
这组遗留 UI 已经被 `componentFeatureImportAllowlist` 明确标成历史例外。继续把 workbench 子面板留在 `components/**`，会让 `ContestEdit.vue` 和 `ContestEditWorkspacePanel.vue` 一直承担不必要的 feature import。最小正确改动是：

- 把 workbench 子面板迁入 `features/contest-workbench/ui`
- 把编辑页 workspace shell 迁入 `features/platform-contests/ui`
- 让 `ContestEdit.vue` 只从两个 feature public API 组合页面

这样可以同时清掉 `contest-workbench` 的多条 allowlist，并把 `ContestEditWorkspacePanel.vue` 的混合 owner 从 legacy component 目录里拿出来。

## Files to modify
- .harness/reuse-decisions/contest-workbench-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-27-contest-workbench-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-27-contest-workbench-feature-ui-normalization-review.md
- code/frontend/src/features/contest-workbench/index.ts
- code/frontend/src/features/contest-workbench/ui/index.ts
- code/frontend/src/features/contest-workbench/ui/ContestWorkbenchStageTabs.vue
- code/frontend/src/features/contest-workbench/ui/ContestWorkbenchSummaryStrip.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeFilterStrip.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeSummaryStrip.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/views/platform/ContestEdit.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `contest-workbench` 会拥有完整的 workbench 子面板 UI public API。
- `platform-contests` 会拥有 `ContestEditWorkspacePanel.vue` 这类 route shell 级别的 page workspace UI。
- `ContestEdit.vue` 不再直连旧 `components/platform/contest/*` 路径。
- `componentFeatureImportAllowlist` 里这组 `contest-workbench` / `platform-contests` 历史例外可以收掉。
