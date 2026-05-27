# Platform Challenge Detail Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-platform-challenge-detail-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/platform-challenge-detail-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-27-platform-challenge-detail-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-27-platform-challenge-detail-feature-ui-normalization-review.md`
    - `code/frontend/src/features/platform-challenge-detail/index.ts`
    - `code/frontend/src/features/platform-challenge-detail/ui/index.ts`
    - `code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeTopbarPanel.vue`
    - `code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeWorkspaceTabs.vue`
    - `code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeProfilePanel.vue`
    - `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`
    - `code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的 feature-owned UI 收口，只迁题目详情这组单一 feature UI，不改 page model / route owner。
- Gate verdict：Pass after targeted verification

## Findings

- None.

## Material findings

- None.

## Senior implementation assessment

- `AdminChallengeTopbarPanel.vue`、`AdminChallengeWorkspaceTabs.vue`、`AdminChallengeProfilePanel.vue` 已迁入 `features/platform-challenge-detail/ui`，`PlatformChallengeDetailWorkspace.vue` 改为从 `features/platform-challenge-detail` public API 组合这组三件套。
- route view、widget 和 feature model 的 owner 没有被重新打散：`ChallengeDetail.vue` 继续只组合 `usePlatformChallengeDetailRoutePage()` 与 widget，`PlatformChallengeDetailWorkspace.vue` 继续只做事件转发和 writeup slot 组合，`useRouteQueryTabs` 仍留在 feature route model。
- `architectureAllowlist.ts` 中 `platform-challenge-detail` 的 2 条 component->feature 例外和 2 条 widget->legacy component 例外都已移除，相关 raw-source 护栏与 `components.d.ts` 也已切到新路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Residual risk

- 这次只收口题目详情 workspace 三件套，不会顺手迁 `ChallengeManageDirectoryPanel.vue` 等其他 challenge 业务组件。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中“应属于单一 feature 的 page-sized UI / feature-owned UI 收口”这条在 `platform-challenge-detail` 这一组又减少了 4 条 allowlist 残留。
