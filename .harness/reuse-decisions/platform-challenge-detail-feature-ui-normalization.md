# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/challenge/AdminChallengeTopbarPanel.vue
- code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue
- code/frontend/src/components/platform/challenge/AdminChallengeProfilePanel.vue
- code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue
- code/frontend/src/features/platform-challenge-detail/ui/index.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts
- code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `platform-awd-challenges` 上一轮已经把 `AWDChallengeEditorDialog.vue` 与 `AwdChallengeImportSection.vue` 迁入 `features/platform-awd-challenges/ui`，并同步收掉 component->feature allowlist。
- `features/platform-challenge-detail/ui/*` 当前已经承接 Flag 配置相关 UI，说明这个 feature 已有稳定的 UI 落点。
- `PlatformChallengeDetailWorkspace.vue` 现在通过 widget 组合 topbar、tabs、profile 和 writeup slot，这三块 UI 都只服务题目详情这一条 feature/workspace，不属于跨 feature 共享原语。

## Decision
refactor_existing

## Reason
`AdminChallengeTopbarPanel.vue`、`AdminChallengeWorkspaceTabs.vue`、`AdminChallengeProfilePanel.vue` 只服务 `platform-challenge-detail`，并直接依赖该 feature 暴露的 contract 或下游 flag 配置 UI。继续把它们留在 `components/platform/challenge` 只会让 `componentFeatureImportAllowlist` 和 `widgetLegacyComponentImportAllowlist` 继续冻结历史路径。最小正确改动是把这组三件套迁入 `features/platform-challenge-detail/ui`，由 widget 改为从 feature public API 组合它们，同时保留 route page、widget 和 feature model 的 owner 不变。

## Files to modify
- .harness/reuse-decisions/platform-challenge-detail-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-27-platform-challenge-detail-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-27-platform-challenge-detail-feature-ui-normalization-review.md
- code/frontend/src/features/platform-challenge-detail/index.ts
- code/frontend/src/features/platform-challenge-detail/ui/index.ts
- code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeTopbarPanel.vue
- code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeWorkspaceTabs.vue
- code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeProfilePanel.vue
- code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts
- code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts
- code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `platform-challenge-detail` 会同时持有 topbar、tabs、profile 与 flag panel 这一整组单一 feature UI。
- `PlatformChallengeDetailWorkspace.vue` 会继续是 widget 组合层，但不再直连旧 `components/platform/challenge` 路径。
- `componentFeatureImportAllowlist` 中这两条 `platform-challenge-detail` 例外，以及 widget 对旧 challenge detail 组件的两条 legacy import 例外，都可以收掉。
