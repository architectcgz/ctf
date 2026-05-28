# Contest Edit Surface Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-contest-edit-surface-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-edit-surface-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-27-contest-edit-surface-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-27-contest-edit-surface-feature-ui-normalization-review.md`
    - `code/frontend/src/features/platform-contests/index.ts`
    - `code/frontend/src/features/platform-contests/ui/index.ts`
    - `code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue`
    - `code/frontend/src/features/contest-workbench/index.ts`
    - `code/frontend/src/features/contest-workbench/ui/index.ts`
    - `code/frontend/src/features/contest-workbench/ui/ContestChallengeEditorDialog.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestAwdChallengeSelectorSection.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestChallengeSettingsSection.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue`
    - `code/frontend/src/views/platform/ContestEdit.vue`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestEditTopbarExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
    - `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
    - `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
    - `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 contest edit 剩余 feature-owned UI surface 的收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。当前实现没有把 route shell owner、dialog submit owner 或 AWD 目录筛选 owner 回流到旧 `components/platform/contest/*` 路径。

## Senior implementation assessment

- `ContestEditTopbarPanel.vue` 已迁入 `features/platform-contests/ui`，`ContestEdit.vue` 现在只通过 `features/platform-contests` public API 组合 topbar、workspace 和 page model。
- `ContestChallengeEditorDialog.vue` 与 `ContestAwdChallengeSelectorSection.vue`、`ContestChallengeSettingsSection.vue` 已整体迁入 `features/contest-workbench/ui`，避免 dialog 父子组件继续分裂在两套目录 owner 之下。
- `ContestChallengeOrchestrationPanel.vue` 改为 feature 内部相对 import dialog，contest edit 线上这一组单一 feature UI surface 的目录边界已经对齐。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestEditTopbarExtraction.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只收口 contest edit 线上的 topbar 与 challenge editor dialog，不会顺手迁 `AWDChallengeConfigPanel.vue`；下一刀如果继续压缩 `components/platform/contest/*`，更适合直接瞄准它。

## Touched known-debt status

- contest edit 线上这组仍滞留在 `components/platform/contest/*` 的单一 feature UI surface 已在 touched surface 内完成收口。
