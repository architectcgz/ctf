# Reuse Decision

## Change type
frontend architecture / oversized component / decomposition

## Existing code searched
- code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue
- code/frontend/src/views/platform/ContestAwdConfig.vue
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts
- code/frontend/src/components/platform/contest/ContestAwdConfigTopbar.vue
- code/frontend/src/components/platform/contest/ContestAwdServiceDirectory.vue
- code/frontend/src/components/platform/contest/ContestAwdEditorHeader.vue
- code/frontend/src/components/platform/contest/ContestAwdScoreWeights.vue
- code/frontend/src/components/platform/contest/ContestAwdDebugStation.vue
- code/frontend/src/components/platform/contest/ContestAwdConfigFooter.vue
- code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue
- code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestChallengeEditorDialog.vue` 已按“父壳保留 form / validation / submit owner，稳定展示块抽到 section”模式拆分。
- `AWDChallengeLibraryPage.vue` 已按“共享 page surface + library / import 分区”模式拆分，证明 AWD 线上的大组件可以先按稳定展示块收口，而不必连同 route / feature owner 一起迁移。
- `ContestAwdConfigWorkspaceShell.vue` 当前已经把 topbar、服务目录、标题、分值区、debug station、footer 拆出，剩余最肥的是 `Checker Parameters` 画布及其四种 checker type 分支。

## Decision
refactor_existing

## Reason
这次不是新增 AWD 配置能力，也不是 `feature-owned UI` 迁移。最小正确改动是把 `ContestAwdConfigWorkspaceShell.vue` 中剩余的 checker 配置画布按稳定分区拆开：父壳继续保留 draft、字段错误、服务选择、保存、预览和 checker type 判定 owner；画布层只负责模板与局部展示结构。这样可以继续收口 `TD-1`，同时避免把 route / async owner 从 `useContestAwdConfigPage()` 打散。

## Files to modify
- .harness/reuse-decisions/contest-awd-config-workspace-shell-decomposition.md
- docs/plan/impl-plan/2026-05-27-contest-awd-config-workspace-shell-decomposition-implementation-plan.md
- docs/reviews/frontend/2026-05-27-contest-awd-config-workspace-shell-decomposition-review.md
- code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue
- code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue
- code/frontend/src/components/platform/contest/ContestAwdLegacyProbeFields.vue
- code/frontend/src/components/platform/contest/ContestAwdHttpStandardFields.vue
- code/frontend/src/components/platform/contest/ContestAwdTcpStandardFields.vue
- code/frontend/src/components/platform/contest/ContestAwdScriptCheckerFields.vue
- code/frontend/src/components/platform/contest/contestAwdConfigTypes.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果这轮 checker 画布收口顺利，`ContestAwdConfigWorkspaceShell.vue` 会从“剩余超大壳”进一步缩成真正的 workspace surface。
- `ContestChallengeEditorDialog.vue` 和 `AWDChallengeLibraryPage.vue` 的拆分模式可继续作为 contest / AWD 线后续大组件壳收口样板。
