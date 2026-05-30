# Reuse Decision

## Change type
frontend refactor / contest AWD readiness override workflow cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `useAwdReadinessDecision.ts` 当前已经是 readiness summary + override dialog + override execute 的唯一组合入口。
- `useAwdRoundOperations.ts` 已只在被 readiness gate 拦截时委托 `openOverrideDialog(...)`，没有继续内联 override 执行逻辑。
- `useAwdContestStateFlags.ts` 上一轮已承接 runtime stage / round operation / auto refresh policy，因此本轮不应再把 runtime rule 拉回 `useAwdReadinessDecision.ts`。

## Decision
refactor_existing

## Reason
当前 `useAwdReadinessDecision.ts` 的 owner 基本正确，但还留有两处更深层 workflow 混写：

- `openOverrideDialog()` 直接绑定了“刷新 readiness 摘要 + 打开弹层”，但对刷新失败没有本地兜底，异常会沿着原始操作链路继续冒泡。
- `confirmOverrideAction()` 里直接写死 `create_round` / `run_current_round_check` 两条 override 执行分支；现在分支数还少，但 action executor 已经和 dialog state owner 混在一起。

本轮的最小正确改动是：

- 保留 `useAwdReadinessDecision.ts` 作为 readiness override workflow owner
- 给 `openOverrideDialog()` 补上 refresh failure 的本地错误处理，保证弹层打开链路不会因为 readiness 摘要刷新失败而静默炸穿
- 把 `confirmOverrideAction()` 里的 override action 执行分支收口成更明确的 action executor helper，减少后续继续增长时的 if/else 混写

本轮不调整：

- `useAwdContestStateFlags.ts` 的 runtime policy owner
- `AWDReadinessSummary.vue` / `AWDReadinessOverrideDialog.vue` 的 UI contract
- `useAwdRoundOperations.ts` 的 mutation owner

## Files to modify
- .harness/reuse-decisions/awd-readiness-override-workflow-cleanup.md
- docs/plan/impl-plan/2026-05-30-awd-readiness-override-workflow-cleanup-plan.md
- docs/reviews/frontend/2026-05-30-awd-readiness-override-workflow-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts

## After implementation
- readiness override workflow 仍由 `useAwdReadinessDecision.ts` 单点承接
- override dialog 打开前的 readiness refresh 有清晰失败兜底
- override action 的执行 owner 不再和 dialog state 更新细节完全混在同一段分支里
