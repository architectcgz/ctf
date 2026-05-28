# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/PlatformContestFormPanel.vue
- code/frontend/src/components/platform/contest/PlatformContestFormDialog.vue
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/features/platform-contests/model/contestFormSupport.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestEditTopbarPanel.vue`、`AWDChallengeConfigPanel.vue` 已迁入 `features/platform-contests/ui`，contest edit / contest manage 的单一 feature UI 已开始按 `platform-contests` owner 收口。
- `ContestChallengeEditorDialog.vue` 已迁入 `features/contest-workbench/ui`，说明当前迁移模式已经稳定为 “feature public API + feature-owned UI + route shell 组合”。
- `PlatformContestFormDialog.vue` 当前只包裹 `PlatformContestFormPanel.vue` 和 `AdminSurfaceModal`，没有独立 page owner 或跨 feature 逻辑。

## Decision
refactor_existing

## Reason
`PlatformContestFormPanel.vue` 与 `PlatformContestFormDialog.vue` 当前只服务 `platform-contests` 这条 feature：`ContestOrchestrationPage.vue`、`ContestEditWorkspacePanel.vue`、`ContestManage.vue` 都是同一 feature / route shell 线上的调用面。继续把这组表单 UI 留在 `components/platform/contest/*`，会让 `componentFeatureImportAllowlist` 里保留两条历史例外。最小正确改动是：

- 把 `PlatformContestFormPanel.vue`、`PlatformContestFormDialog.vue` 迁入 `features/platform-contests/ui`
- `ContestOrchestrationPage.vue`、`ContestEditWorkspacePanel.vue`、`ContestManage.vue` 改走 feature 内部 UI 或 feature public API
- `PlatformContestFormPanel.vue` 内部类型改为相对 model import，避免 feature UI 自己回头走 feature root barrel
- 更新 raw-source / dialog / theme token / surface alignment 测试与 backlog

这样可以继续清掉 `platform-contests` 线上残余的 legacy component surface，并收掉这两条 allowlist。

## Files to modify
- .harness/reuse-decisions/platform-contest-form-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-platform-contest-form-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-platform-contest-form-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-contests/ui/PlatformContestFormPanel.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestFormDialog.vue
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `platform-contests` 会持有完整的 contest form panel / dialog UI owner。
- `ContestManage.vue`、`ContestOrchestrationPage.vue` 与 `ContestEditWorkspacePanel.vue` 不再引用旧 `components/platform/contest/PlatformContestForm*` 路径。
- `componentFeatureImportAllowlist` 里这两条 `PlatformContestForm*.vue -> @/features/platform-contests` 历史例外会在 touched surface 内收口。
