# AWD Inspector UI Cluster Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-inspector-ui-cluster-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-inspector-ui-cluster-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-awd-inspector-ui-cluster-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-awd-inspector-ui-cluster-feature-ui-normalization-review.md`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/features/awd-inspector/index.ts`
    - `code/frontend/src/features/awd-inspector/ui/index.ts`
    - `code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDTrafficPanel.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDScoreboardSummaryPanel.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDRoundHeaderPanel.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDAttackLogPanel.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDServiceAlertBanner.vue`
    - `code/frontend/src/features/awd-inspector/ui/awdInspector.types.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
    - `code/frontend/src/views/platform/ContestOperations.vue`
    - `code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts`
    - `code/frontend/src/views/platform/__tests__/ContestOperations.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
    - `code/frontend/src/views/platform/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `awd-inspector` 单一 feature UI owner 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `AWDRoundInspector.vue` 及其配套的 traffic、service status、scoreboard、header、attack log、alert banner 和共享 types 已整体迁入 `features/awd-inspector/ui`，不再继续滞留在旧 `components/platform/contest/*` 目录。
- `features/awd-inspector/index.ts` 现在同时暴露 `model` 与 `ui` public API，`AWDOperationsPanel.vue` 与 `ContestOperations.vue` 已改为通过 feature owner 引用 `AWDRoundInspector`、`AWDServiceAlertBanner`。
- 本次没有顺手调整 `AWDReadiness*`、实例编排、服务检查或题解 dialog/runtime owner，因此 inspector 结果视图 owner 与 awd admin 运行时 workflow owner 继续分刀，边界保持清楚。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/components/platform/__tests__/AWDServiceStatusPanel.test.ts src/views/platform/__tests__/ContestOperations.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只处理 `awd-inspector` 结果视图 cluster，不继续拆 `AWDReadiness*`、`AWDInstanceOrchestrationPanel.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue`；如果后续继续下钻，应按 `contest-awd-admin` 或运行时 owner 重新分刀评估。

## Touched known-debt status

- `awd-inspector` 已知的大颗粒结果视图组件在 touched surface 内已从旧 `components/platform/contest/*` 收口到 feature owner，当前这一线的遗留重点开始转向 `contest-awd-admin` 侧的 readiness、实例编排和 dialog/runtime cluster。
