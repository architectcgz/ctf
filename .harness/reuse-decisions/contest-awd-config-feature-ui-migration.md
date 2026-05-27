# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/views/platform/ContestAwdConfig.vue
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue
- code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue
- code/frontend/src/components/platform/contest/ContestAwdConfigTopbar.vue
- code/frontend/src/components/platform/contest/ContestAwdServiceDirectory.vue
- code/frontend/src/components/platform/contest/ContestAwdDebugStation.vue
- code/frontend/src/components/platform/contest/ContestAwdConfigFooter.vue
- code/frontend/src/features/contest-awd-config/index.ts
- code/frontend/src/features/contest-awd-config/model/index.ts
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `TeacherDashboardPage.vue`、`ContestOrchestrationPage.vue`、`StudentManagementPage.vue`、`ClassManagementPage.vue` 已证明，page-sized workspace shell 可以迁到 `features/*/ui`，route view 保留组合壳。
- `ContestAwdConfig.vue` 当前已经是薄 route 壳：只消费 `useContestAwdConfigPage()`，把大量状态和回调透传给 `ContestAwdConfigWorkspaceShell`。
- `useContestAwdConfigPage.ts` 当前已经是 AWD 配置页的 router / loader / preview / save owner，不需要本轮再补 route hook。

## Decision
refactor_existing

## Reason
这次不是新增 AWD 配置能力，而是继续收口“单一 feature 的 page-sized UI 壳仍停在 `components/`”的遗留。最小正确改动是把 `ContestAwdConfigWorkspaceShell.vue` 迁到 `features/contest-awd-config/ui/`，并让 `views/platform/ContestAwdConfig.vue` 通过 `features/contest-awd-config` public API 组合 page model 与 workspace shell。

## Files to modify
- .harness/reuse-decisions/contest-awd-config-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-contest-awd-config-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-contest-awd-config-feature-ui-migration-review.md
- code/frontend/src/features/contest-awd-config/index.ts
- code/frontend/src/features/contest-awd-config/ui/*
- code/frontend/src/views/platform/ContestAwdConfig.vue
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果迁移顺利，`contest-awd-config` 会和最近这批 feature 一样拥有明确的 `model + ui` public API。
- 本轮不移动 `ContestAwdCheckerConfigSection.vue`、`ContestAwdConfigTopbar.vue`、`ContestAwdServiceDirectory.vue` 这些子组件，也不改 preview/save/load owner。
