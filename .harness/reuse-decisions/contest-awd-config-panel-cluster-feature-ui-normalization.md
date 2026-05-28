# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue
- code/frontend/src/components/platform/contest/ContestAwdConfigTopbar.vue
- code/frontend/src/components/platform/contest/ContestAwdConfigFooter.vue
- code/frontend/src/components/platform/contest/ContestAwdDebugStation.vue
- code/frontend/src/components/platform/contest/ContestAwdEditorHeader.vue
- code/frontend/src/components/platform/contest/ContestAwdScoreWeights.vue
- code/frontend/src/components/platform/contest/ContestAwdServiceDirectory.vue
- code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue
- code/frontend/src/components/platform/contest/ContestAwdHttpStandardFields.vue
- code/frontend/src/components/platform/contest/ContestAwdLegacyProbeFields.vue
- code/frontend/src/components/platform/contest/ContestAwdScriptCheckerFields.vue
- code/frontend/src/components/platform/contest/ContestAwdTcpStandardFields.vue
- code/frontend/src/components/platform/contest/contestAwdConfigTypes.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestAwdConfigWorkspaceShell.vue` 已经迁入 `features/contest-awd-config/ui`，说明 AWD 配置这条 route shell owner 已经明确落在 feature。
- `platform-contests` 线上最近几刀已经按 “feature shell + feature-owned subpanel + raw-source 测试同步” 的模式，连续迁走 `AWDChallengeConfigPanel.vue`、`PlatformContestTable.vue`、`ContestAwdPreflightPanel.vue`。
- `ContestAwdCheckerConfigSection.vue` 当前还继续通过相对路径串起四个 fields 子件和 `contestAwdConfigTypes.ts`，这组本身已经形成稳定的小型 UI cluster。

## Decision
refactor_existing

## Reason
`ContestAwdConfigWorkspaceShell.vue` 现在仍从旧 `components/platform/contest/*` 拉取 7 个 panel，加上 `ContestAwdCheckerConfigSection.vue` 内部再回头依赖 4 个 fields 子件和 `contestAwdConfigTypes.ts`。这导致 AWD 配置 feature 的主要 UI owner 已经在 `features/contest-awd-config/ui`，但它的核心编辑画布仍留在 legacy component 目录里。最小正确改动是：

- 把 `ContestAwdConfigTopbar.vue`、`ContestAwdConfigFooter.vue`、`ContestAwdDebugStation.vue`、`ContestAwdEditorHeader.vue`、`ContestAwdScoreWeights.vue`、`ContestAwdServiceDirectory.vue`、`ContestAwdCheckerConfigSection.vue` 迁入 `features/contest-awd-config/ui`
- 同时把 `ContestAwdHttpStandardFields.vue`、`ContestAwdLegacyProbeFields.vue`、`ContestAwdScriptCheckerFields.vue`、`ContestAwdTcpStandardFields.vue` 和 `contestAwdConfigTypes.ts` 一并迁入同目录，保持 checker config cluster 自洽
- `ContestAwdConfigWorkspaceShell.vue` 改为 feature 内部相对 import
- 更新 `components.d.ts`、raw-source 测试和 backlog 记录

这样能把 AWD 配置这组明显只服务单一 feature 的 UI cluster 完整收口，不再留下 “shell 在 feature，panel/fields 在 legacy component” 的半迁移状态。

## Files to modify
- .harness/reuse-decisions/contest-awd-config-panel-cluster-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-contest-awd-config-panel-cluster-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-contest-awd-config-panel-cluster-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigTopbar.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigFooter.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdDebugStation.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdEditorHeader.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdScoreWeights.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdServiceDirectory.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdCheckerConfigSection.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdHttpStandardFields.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdLegacyProbeFields.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdScriptCheckerFields.vue
- code/frontend/src/features/contest-awd-config/ui/ContestAwdTcpStandardFields.vue
- code/frontend/src/features/contest-awd-config/ui/contestAwdConfigTypes.ts
- code/frontend/src/features/contest-awd-config/ui/index.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- AWD 配置 workspace shell 会完整持有自己的 panel / fields / UI types cluster。
- `ContestAwdConfigWorkspaceShell.vue` 和相关 raw-source 测试不再引用旧 `components/platform/contest/*` 这组 AWD config panel 路径。
- `contest-awd-config` 这条 feature 的主要编辑 UI surface 会从 legacy component 目录整体退场。
