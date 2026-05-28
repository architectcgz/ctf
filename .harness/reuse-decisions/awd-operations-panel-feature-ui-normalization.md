# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/AWDOperationsPanel.vue
- code/frontend/src/views/platform/ContestOperations.vue
- code/frontend/src/features/contest-awd-admin/index.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/ContestOperations.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `AWDChallengeConfigPanel.vue` 已迁入 `features/platform-contests/ui`，`ContestAwdConfigWorkspaceShell.vue` 的 panel cluster 也已整体迁入 `features/contest-awd-config/ui`，说明 contest / AWD 线已经在按 feature owner 收口主要 UI。
- 当前 `architectureAllowlist.ts` 已明确存在 `components/platform/contest/AWDOperationsPanel.vue -> @/features/contest-awd-admin` 历史例外，表明 owner 早已确认在 `contest-awd-admin`。
- `ContestOperations.vue` 只是 route shell，当前只负责通过 `useContestOperationsPage()` 取单场 contest 视图数据，再把其传给 `AWDOperationsPanel.vue`。

## Decision
refactor_existing

## Reason
`AWDOperationsPanel.vue` 本身直接依赖 `usePlatformContestAwd()`，是典型的单一 feature UI，但现在还滞留在旧 `components/platform/contest/*` 路径，并靠 allowlist 放行。最小正确改动是：

- 把 `AWDOperationsPanel.vue` 迁入 `features/contest-awd-admin/ui`
- `ContestOperations.vue` 改为通过 `@/features/contest-awd-admin` public API 组合 panel
- `features/contest-awd-admin/index.ts` 增加 UI export
- 更新 `components.d.ts`、组件测试、raw-source 测试和 `architectureAllowlist.ts`

这刀先只迁 panel 本体，不顺手迁它依赖的 round/readiness/dialog 子件，避免把 `contest-awd-admin` 的 UI owner 收口和更深层 AWD runtime primitive 迁移混成一刀。

## Files to modify
- .harness/reuse-decisions/awd-operations-panel-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-operations-panel-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-operations-panel-feature-ui-normalization-review.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/features/contest-awd-admin/index.ts
- code/frontend/src/features/contest-awd-admin/ui/index.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/views/platform/ContestOperations.vue
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `AWDOperationsPanel.vue` 会归 `features/contest-awd-admin/ui` 持有。
- `ContestOperations.vue` 和相关 raw-source / component 测试不再引用旧 `components/platform/contest/AWDOperationsPanel.vue` 路径。
- `architectureAllowlist.ts` 中这条 `AWDOperationsPanel.vue -> @/features/contest-awd-admin` 历史例外会收掉。
