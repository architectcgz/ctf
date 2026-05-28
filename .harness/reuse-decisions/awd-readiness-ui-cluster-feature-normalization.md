# Reuse Decision

## Change type
frontend architecture / shared feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDReadinessChecklist.vue
- code/frontend/src/components/platform/contest/AWDReadinessDecisionHUD.vue
- code/frontend/src/components/platform/contest/AWDReadinessSummary.vue
- code/frontend/src/components/platform/contest/AWDReadinessOverrideDialog.vue
- code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/components/platform/__tests__/AWDReadinessSummary.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase19.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- code/frontend/src/components.d.ts
- docs/architecture/features/AWD开赛就绪门禁设计.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `features/awd-inspector/ui` 已承接跨页面复用的 AWD 巡检结果视图，说明 AWD 子能力可以作为独立 feature owner 暴露 UI public API。
- `ContestAwdPreflightPanel.vue` 上一刀已经迁入 `features/platform-contests/ui`，但仍显式依赖旧 `components/platform/contest/*` 下的 readiness primitive。
- `AWDOperationsPanel.vue` 与 `ContestManage.vue` 都已经是 feature / route shell owner，本身不应继续直接持有 readiness 子组件的 legacy 路径。

## Decision
refactor_existing

## Reason
`AWDReadinessChecklist.vue`、`AWDReadinessDecisionHUD.vue`、`AWDReadinessSummary.vue`、`AWDReadinessOverrideDialog.vue` 描述的是同一条 AWD readiness capability，而不是单一页面私有 UI。它们同时被 `platform-contests`、`contest-awd-admin` 和 `ContestManage.vue` 消费；如果塞回任一现有 feature，会把旧 `components -> feature` 残片换成新的跨 feature 倾斜。最小正确改动是：

- 新增 `features/awd-readiness/ui/index.ts`
- 把 4 个 readiness UI 组件整体迁入 `features/awd-readiness/ui`
- 新增 `features/awd-readiness/index.ts` 暴露 public API
- 更新 `ContestAwdPreflightPanel.vue`、`AWDOperationsPanel.vue`、`ContestManage.vue`、相关测试、`components.d.ts` 和事实文档

本轮不迁 `AWDInstanceOrchestrationPanel.vue`，也不合并 `platform-contests` / `contest-awd-admin` 各自的 readiness workflow model。

## Files to modify
- .harness/reuse-decisions/awd-readiness-ui-cluster-feature-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-readiness-ui-cluster-feature-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-readiness-ui-cluster-feature-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/awd-readiness/index.ts
- code/frontend/src/features/awd-readiness/ui/index.ts
- code/frontend/src/features/awd-readiness/ui/AWDReadinessChecklist.vue
- code/frontend/src/features/awd-readiness/ui/AWDReadinessDecisionHUD.vue
- code/frontend/src/features/awd-readiness/ui/AWDReadinessSummary.vue
- code/frontend/src/features/awd-readiness/ui/AWDReadinessOverrideDialog.vue
- code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/components/platform/__tests__/AWDReadinessSummary.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase19.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- docs/architecture/features/AWD开赛就绪门禁设计.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `AWDReadiness*` UI cluster 会归 `features/awd-readiness/ui` 持有。
- `platform-contests` 与 `contest-awd-admin` 会通过 `@/features/awd-readiness` public API 复用 readiness UI，而不再各自引用旧 `components/platform/contest/*` 路径。
- AWD readiness 的 UI owner 会从“历史共用组件目录”收口为明确的 capability feature。
