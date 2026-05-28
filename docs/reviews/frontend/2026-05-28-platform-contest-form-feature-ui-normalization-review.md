# Platform Contest Form Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-platform-contest-form-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/platform-contest-form-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-platform-contest-form-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-platform-contest-form-feature-ui-normalization-review.md`
    - `code/frontend/src/features/platform-contests/ui/PlatformContestFormPanel.vue`
    - `code/frontend/src/features/platform-contests/ui/PlatformContestFormDialog.vue`
    - `code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue`
    - `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
    - `code/frontend/src/features/platform-contests/ui/index.ts`
    - `code/frontend/src/views/platform/ContestManage.vue`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
    - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `platform-contests` 剩余 feature-owned UI surface 的收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 已修正：`PlatformContestFormPanel.vue` 迁入 feature UI 后如果继续从 `@/features/platform-contests` 回读类型，会让 feature UI 走 feature root barrel。最终实现把 `ContestFieldLocks` / `ContestFormDraft` 改成相对 `../model` import，避免 feature 内部自循环依赖。

## Senior implementation assessment

- `PlatformContestFormPanel.vue` 与 `PlatformContestFormDialog.vue` 已整体迁入 `features/platform-contests/ui`，contest create / edit 表单 UI 不再滞留在旧 `components/platform/contest/*` 路径。
- `ContestManage.vue` 已通过 `features/platform-contests` public API 组合 dialog，`ContestOrchestrationPage.vue` 与 `ContestEditWorkspacePanel.vue` 也都改成 feature 内部导入 panel。
- `architectureAllowlist.ts` 中 `PlatformContestFormDialog.vue -> @/features/platform-contests` 与 `PlatformContestFormPanel.vue -> @/features/platform-contests` 两条历史例外已经删掉。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只收口 `PlatformContestFormPanel.vue` / `PlatformContestFormDialog.vue` 的目录 owner，不会顺手拆表单内部 section；如果后续继续在同一表单里叠新编辑分区，下一刀应按 section 拆，不要再把整块表单塞回 `components/`。

## Touched known-debt status

- `platform-contests` 这组表单 UI 对应的两条 `components -> feature` allowlist 已在 touched surface 内完成收口。
