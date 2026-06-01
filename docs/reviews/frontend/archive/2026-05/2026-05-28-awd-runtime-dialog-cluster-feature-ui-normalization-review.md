# AWD Runtime Dialog Cluster Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-runtime-dialog-cluster-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-runtime-dialog-cluster-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-awd-runtime-dialog-cluster-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-awd-runtime-dialog-cluster-feature-ui-normalization-review.md`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
    - `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
    - `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `contest-awd-admin` runtime/dialog cluster 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 已整体迁入 `features/contest-awd-admin/ui`，AWD runtime dialogs 不再继续滞留在旧 `components/platform/contest/*` 目录。
- `AWDOperationsPanel.vue` 已切到 feature 内部相对 import，round create / service check / attack log 这组 dialog owner 继续清楚地挂在 `contest-awd-admin` runtime cluster 下。
- 本次没有顺手改 `usePlatformContestAwd()` 的 dialog open state、保存流程或请求节流逻辑，因此 dialog owner 收口与 workflow owner 收口继续分刀，边界保持清楚。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这轮只收 runtime dialog owner，不继续清 `AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue` 等仍在旧目录的 AWD operations 子件；如果后续继续下钻，应再按 `contest-awd-admin` owner 分刀。

## Touched known-debt status

- `contest-awd-admin` runtime/dialog cluster 已在 touched surface 内继续把 3 个单一 feature dialog 从旧 contest 组件目录收口到 `features/contest-awd-admin/ui`；当前这条线的后续重点开始转向剩余的 AWD operations 子件 owner。
