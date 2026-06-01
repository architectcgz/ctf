# Contest AWD Preflight Panel Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-contest-awd-preflight-panel-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-awd-preflight-panel-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-contest-awd-preflight-panel-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-contest-awd-preflight-panel-feature-ui-normalization-review.md`
    - `code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue`
    - `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `platform-contests` 剩余 feature-owned UI surface 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `ContestAwdPreflightPanel.vue` 已迁入 `features/platform-contests/ui`，contest edit 的 AWD 赛前检查面板不再滞留在旧 `components/platform/contest/*` 路径。
- `ContestEditWorkspacePanel.vue` 已切到 feature 内部相对 import，contest edit stage owner 仍留在 workspace shell / feature model，没有因为迁位回流到 route view。
- 这次显式保持 `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue` 仍由旧 readiness primitive owner 持有，避免把 primitive 收口和 route shell 收口混成一刀，改动面更清楚。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只处理 `ContestAwdPreflightPanel.vue` 的 feature owner 迁位，不会顺手迁移 `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue`；如果后续继续做 AWD readiness 归位，需要再单独判断 primitive owner。

## Touched known-debt status

- `platform-contests` 线上的 contest edit 展示残片已继续缩小；本次 touched surface 内没有留下新的 `ContestAwdPreflightPanel.vue` 旧路径引用。
