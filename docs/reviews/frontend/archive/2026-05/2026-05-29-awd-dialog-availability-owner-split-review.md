# AWD Dialog Availability Owner Split 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-dialog-availability-owner-split-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-dialog-availability-owner-split.md`
    - `docs/plan/impl-plan/2026-05-29-awd-dialog-availability-owner-split-plan.md`
    - `docs/reviews/frontend/2026-05-29-awd-dialog-availability-owner-split-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogAvailability.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogAvailability.test.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
    - `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- Classification check：同意按 `contest-awd-admin` feature 内 deeper owner 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useAwdOperationsDialogState.ts` 现在只保留 dialog open/close state、`nextRoundNumber`、save success -> close dialog workflow 与 override close guard，不再继续承接 availability / hint 派生。
- `useAwdOperationsDialogAvailability.ts` 已成为 AWD operations dialog cluster 内“能不能录入 / 为什么不能录入”的唯一 owner，明确承接 `canRecordServiceChecks`、`canRecordAttackLogs`、`serviceCheckHint`、`attackLogHint`。
- `AWDOperationsPanel.vue` 已改为在组合面上同时消费 availability owner 和 state owner，继续向 `AWDRoundInspector` 透传既有 props，不改变用户可见行为。
- `useAwdOperationsDialogState.test.ts`、`useAwdOperationsDialogAvailability.test.ts` 与 `awdOperationsPanelTabsExtraction.test.ts` 已同步覆盖新的边界：state owner 不再内联 `computed hint`，availability owner 则明确承接这组规则。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogAvailability.test.ts src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- 当前 availability 规则仍只依赖 `teams` 和 `challengeLinks`；如果后续增长到依赖 readiness、权限或 round status，新 owner 依然可能继续变大，但至少不会再污染 dialog state owner。

## Touched known-debt status

- 这轮 touched surface 上，AWD dialog cluster 的 availability / hint owner 已经从 dialog local state 中分离；当前 residual 重点已不再是这组派生规则的 owner 混写，而转向未来可能新增的 gating 维度本身。
