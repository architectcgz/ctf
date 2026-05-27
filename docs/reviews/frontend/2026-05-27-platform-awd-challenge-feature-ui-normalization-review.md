# Platform AWD Challenge Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-platform-awd-challenge-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/platform-awd-challenge-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-27-platform-awd-challenge-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-27-platform-awd-challenge-feature-ui-normalization-review.md`
    - `code/frontend/src/features/platform-awd-challenges/index.ts`
    - `code/frontend/src/features/platform-awd-challenges/ui/index.ts`
    - `code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue`
    - `code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeEditorDialog.vue`
    - `code/frontend/src/features/platform-awd-challenges/ui/AwdChallengeImportSection.vue`
    - `code/frontend/src/views/platform/AWDChallengeLibrary.vue`
    - `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts`
    - `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的 feature-owned UI 收口，只迁单一 feature UI 落点，不改 page model / async owner。
- Gate verdict：Pass after targeted verification

## Findings

- None.

## Material findings

- None.

## Senior implementation assessment

- `AWDChallengeEditorDialog.vue` 与 `AwdChallengeImportSection.vue` 已迁入 `features/platform-awd-challenges/ui`，`AWDChallengeLibrary.vue` route view 与 `AWDChallengeLibraryPage.vue` 都改为通过 feature public API / feature UI 读取这组单一 feature 的 surface。
- 这次没有触碰 `useAwdChallengeLibraryPage.ts`、导入上传或保存流程的 owner；对话框的本地 draft / 校验 / 保存短路、导入区块的确认导入交互都保持原行为。
- `architectureAllowlist.ts` 中 `platform-awd-challenges` 这两条 component->feature 历史例外已移除，相关 raw-source 护栏和 `components.d.ts` 也已切到新路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Residual risk

- `AwdChallengeLibrarySection.vue` 与 `AwdChallengeWorkspaceHeader.vue` 仍会留在原业务组件目录，本轮只收口当前被 allowlist 冻结的单一 feature UI。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中“应属于单一 feature 的 page-sized UI / feature-owned UI 收口”这条在 `platform-awd-challenges` 这一组又减少了两条 allowlist 残留。
