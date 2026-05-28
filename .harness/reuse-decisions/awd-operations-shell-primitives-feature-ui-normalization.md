# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDContestSelectorField.vue
- code/frontend/src/components/platform/contest/AWDRuntimePendingState.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `AWDInstanceOrchestrationPanel.vue` 与 runtime dialog cluster 已迁入 `features/contest-awd-admin/ui`，说明 AWD operations 子件 owner 已经开始和 `contest-awd-admin` 对齐。
- `AWDContestSelectorField.vue` 与 `AWDRuntimePendingState.vue` 当前都只被 `AWDOperationsPanel.vue` 消费，没有第二个 feature / route 复用面。
- `contestUiPrimitiveAdoptionPhase4.test.ts`、`sharedThemeTokenAdoption.test.ts` 已经把这两个子件作为 AWD operations surface 的 raw-source 护栏。

## Decision
refactor_existing

## Reason
`AWDContestSelectorField.vue` 和 `AWDRuntimePendingState.vue` 是 `AWDOperationsPanel.vue` 内部的 operations shell primitives，不属于跨 feature capability。最小正确改动是：

- 把这两个子件迁入 `features/contest-awd-admin/ui`
- `AWDOperationsPanel.vue` 改为 feature 内部相对 import
- 更新 `components.d.ts`、raw-source / theme token 护栏和 backlog 记录

本轮不继续调整 `AWDOperationsPanel.vue` 的 page owner、tab state 或 runtime workflow，也不把这两个子件抽到 shared/common。

## Files to modify
- .harness/reuse-decisions/awd-operations-shell-primitives-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-operations-shell-primitives-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-operations-shell-primitives-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDContestSelectorField.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePendingState.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- AWD operations shell primitives 会归 `features/contest-awd-admin/ui` 持有。
- `AWDOperationsPanel.vue` 不再回头引用旧 `components/platform/contest/*` 下的 selector / pending state 子件。
- `contest-awd-admin` touched surface 内剩余的 legacy AWD operations 组件会进一步缩小。
