# AWD Instance Orchestration Panel Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-instance-orchestration-panel-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-review.md`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
    - `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `contest-awd-admin` 单一 feature UI owner 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `AWDInstanceOrchestrationPanel.vue` 已迁入 `features/contest-awd-admin/ui`，实例编排面板不再继续滞留在旧 `components/platform/contest/*` 目录。
- `AWDOperationsPanel.vue` 已切到 feature 内部相对 import，实例编排 panel owner 继续清楚地挂在 `contest-awd-admin` 这条 runtime 线下，而没有被误拆成新的共享 capability。
- 本次没有顺手调整 `usePlatformContestAwd()` 的实例编排 workflow，也没有继续迁 `AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue`；因此 panel owner 收口与更深层 runtime / dialog cluster 收口继续分刀，边界保持清楚。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `contest-awd-admin` 运行态子件里仍有 `AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue` 等 legacy component 路径；如果后续继续清这条线，应按 runtime / dialog cluster 再单独分刀。

## Touched known-debt status

- `contest-awd-admin` runtime cluster 已在 touched surface 内继续把单一 feature 的实例编排 panel 从旧 contest 组件目录收口到 `features/contest-awd-admin/ui`；当前这条线的后续重点开始进一步转向 service check、attack log 与 round create dialogs。
