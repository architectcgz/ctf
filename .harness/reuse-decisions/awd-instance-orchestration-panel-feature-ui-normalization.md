# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDInstanceOrchestrationPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/index.ts
- code/frontend/src/components.d.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `AWDOperationsPanel.vue` 已迁入 `features/contest-awd-admin/ui`，说明 AWD 运维主面板 owner 已经与 `contest-awd-admin` 对齐。
- `AWDRoundInspector` cluster 已迁入 `features/awd-inspector/ui`，说明 AWD 运维页内部子 panel 可以按 capability 或单一 feature owner 分刀收口。
- `AWDInstanceOrchestrationPanel.vue` 当前只被 `AWDOperationsPanel.vue` 消费，没有第二个 route / feature 复用面。

## Decision
refactor_existing

## Reason
`AWDInstanceOrchestrationPanel.vue` 是 `AWDOperationsPanel.vue` 内部“实例编排”标签页的单一 runtime 子件，不属于跨 feature 共享 capability。最小正确改动是：

- 把 `AWDInstanceOrchestrationPanel.vue` 迁入 `features/contest-awd-admin/ui`
- `AWDOperationsPanel.vue` 改为 feature 内部相对 import
- 更新 `components.d.ts`、相关测试与 backlog 记录

本轮不迁 `AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue`，也不继续调整 `usePlatformContestAwd()` 的实例编排 workflow owner。

## Files to modify
- .harness/reuse-decisions/awd-instance-orchestration-panel-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-instance-orchestration-panel-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `AWDInstanceOrchestrationPanel.vue` 会归 `features/contest-awd-admin/ui` 持有。
- `AWDOperationsPanel.vue` 不再引用旧 `components/platform/contest/AWDInstanceOrchestrationPanel.vue` 路径。
- `contest-awd-admin` 线上的 runtime 子件 owner 会继续从 legacy contest 组件目录收口。
