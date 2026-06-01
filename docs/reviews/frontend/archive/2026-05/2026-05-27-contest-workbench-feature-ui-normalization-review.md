# Contest Workbench Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-contest-workbench-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-workbench-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-27-contest-workbench-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-27-contest-workbench-feature-ui-normalization-review.md`
    - `code/frontend/src/features/contest-workbench/index.ts`
    - `code/frontend/src/features/contest-workbench/ui/index.ts`
    - `code/frontend/src/features/contest-workbench/ui/ContestWorkbenchStageTabs.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestWorkbenchSummaryStrip.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestChallengeFilterStrip.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestChallengeSummaryStrip.vue`
    - `code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue`
    - `code/frontend/src/features/platform-contests/ui/index.ts`
    - `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
    - `code/frontend/src/views/platform/ContestEdit.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/components/platform/__tests__/ContestWorkbenchStageTabs.test.ts`
    - `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
    - `code/frontend/src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestEditWorkspaceExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase23.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase24.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的 feature-owned UI / workspace shell 双 owner 收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 已修正：`ContestChallengeOrchestrationPanel.vue` 迁入 `features/contest-workbench/ui` 后若继续显式 `import { RouterLink } from 'vue-router'`，会违反“feature router access should stay in reviewed route-aware composables” 护栏。最终实现改为使用全局 `RouterLink` 组件，不再把 router import 留在 feature UI。

## Senior implementation assessment

- `contest-workbench` 现在拥有完整的 workbench 子面板 UI public API，`ContestWorkbenchStageTabs`、`ContestWorkbenchSummaryStrip`、`ContestChallengeFilterStrip`、`ContestChallengeSummaryStrip`、`ContestChallengeOrchestrationPanel` 都不再滞留在 legacy component 目录。
- `ContestEditWorkspacePanel.vue` 则按 route shell owner 收回到 `features/platform-contests/ui`，避免把 basics 表单 owner 重新塞进 `contest-workbench`。
- `ContestEdit.vue` 只通过 `features/contest-workbench` 与 `features/platform-contests` public API 组合编辑页，目录边界比原先清楚。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/ContestWorkbenchStageTabs.test.ts src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只收口 `contest-workbench` 子面板与 `ContestEditWorkspacePanel.vue` 的目录 owner，不会顺手迁移 `ContestEditTopbarPanel.vue` 等其它 contest edit surface。

## Touched known-debt status

- `contest-workbench` 这组由 allowlist 冻结的单一 feature UI 与 `ContestEditWorkspacePanel.vue` legacy shell owner 已在 touched surface 内完成收口。
- 下一批更适合继续看 `ContestEditTopbarPanel.vue`、`ContestChallengeEditorDialog.vue` 这类仍留在 contest edit 线上的存量 surface。
