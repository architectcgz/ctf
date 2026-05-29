# AWD Round Dialog And Contract Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-round-dialog-and-contract-owner-cleanup-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-round-dialog-and-contract-owner-cleanup.md`
    - `docs/plan/impl-plan/2026-05-29-awd-round-dialog-and-contract-owner-cleanup-plan.md`
    - `docs/reviews/frontend/2026-05-29-awd-round-dialog-and-contract-owner-cleanup-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts`
    - `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- Classification check：同意按 `contest-awd-admin` feature 内 round dialog + dialog contract owner 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `awdOperationsDialogContracts.ts` 已在上一刀成为 AWD operations dialog cluster 的单点 contract owner；本轮 `AWDOperationsDialogHub.vue` 与 `useAwdOperationsDialogState.ts` 已统一转到这组共享 contract，不再继续内联 payload / override state shape。
- `useAwdOperationsDialogState.ts` 继续保留 dialog cluster 的 workflow owner，同时已用局部 helper 收口 save success 后关闭对应 dialog 的 shared workflow 语义，避免 round / service / attack 三条保存链路重复写三遍 close 逻辑。
- `awd operations panel tabs` 护栏已进一步切到 panel + dialog hub + dialog availability + dialog state + shared contract 的聚合源码视角，继续覆盖 dialog contract owner 与 operations runtime shell 约束。
- `AWDOperationsDialogHub.vue` 从约 `113` 行降到 `87` 行，`useAwdOperationsDialogState.ts` 从本轮实现前的 `178` 行降到 `150` 行；本轮 touched surface 上“dialog cluster 重复内联 payload contract / close-after-save 语义”的债已经完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据，这条缺口需要在交付说明里明确。
- `useAwdOperationsDialogState.ts` 当前仍继续保留 open state owner，规模仍在合理范围内；如果后续 availability / hint workflow 再继续增长，更合适再把 availability presentation 和 mutation workflow 分成两个 composable，而不是现在提前拆散。
- 本轮不改 `AWDOperationsPanel.vue` 的上层 wiring，也不改 `usePlatformContestAwd()` 的远端 mutation owner；如果后续这些边界继续出现 owner 漂移，不应把它和当前 round dialog / contract 收口混在一起。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 里这条 contest / AWD feature 内残余大 surface，已在 AWD operations dialog contract owner 这一组 touched surface 上完成一刀 deeper owner 收口；当前 residual 重点已不再是重复 payload contract，而转向后续如果继续增长的 availability / hint workflow owner 与其它未拆大 surface。
