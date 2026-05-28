# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDRoundCreateDialog.vue
- code/frontend/src/components/platform/contest/AWDServiceCheckDialog.vue
- code/frontend/src/components/platform/contest/AWDAttackLogDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `AWDOperationsPanel.vue`、`AWDInstanceOrchestrationPanel.vue` 已迁入 `features/contest-awd-admin/ui`，说明 AWD runtime 主面板及其实例编排子件 owner 已经与 `contest-awd-admin` 对齐。
- 这 3 个 dialog 当前只被 `AWDOperationsPanel.vue` 消费，没有第二个 route / feature 复用面。
- readiness UI 已单独落到 `features/awd-readiness/ui`，证明 AWD 线上的 shared capability 和单 feature 子件已经开始按 owner 分开收口。

## Decision
refactor_existing

## Reason
`AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 都属于 `contest-awd-admin` runtime cluster 的内部 dialog 子件，不属于跨 feature capability。最小正确改动是：

- 把这 3 个 dialog 迁入 `features/contest-awd-admin/ui`
- `AWDOperationsPanel.vue` 改为 feature 内部相对 import
- 更新 `components.d.ts`、相关 raw-source / duplicate-action / dialog adoption 测试和 backlog 记录

本轮不继续调整 `usePlatformContestAwd()` 的 create/save workflow owner，也不把这 3 个 dialog 再拆成更细粒度 composable。

## Files to modify
- .harness/reuse-decisions/awd-runtime-dialog-cluster-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-runtime-dialog-cluster-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-runtime-dialog-cluster-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- AWD runtime dialog cluster 会归 `features/contest-awd-admin/ui` 持有。
- `AWDOperationsPanel.vue` 不再回头引用旧 `components/platform/contest/*` 下的 runtime dialogs。
- `contest-awd-admin` runtime cluster 在 touched surface 上的 legacy dialog 路径会继续缩小。
