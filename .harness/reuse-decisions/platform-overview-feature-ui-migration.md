# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue
- code/frontend/src/views/platform/PlatformOverview.vue
- code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts
- code/frontend/src/features/platform-overview/model/usePlatformOverviewWorkspace.ts
- code/frontend/src/features/platform-overview/index.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts
- code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `features/challenge-writeup-editor/ui/*` 已经证明单一 feature 的 page-sized UI 可以直接挂到 `features/*/ui`，而不是继续经过 `components/*Page.vue` 中转。
- `PlatformOverviewPage.vue` 当前直接消费 `usePlatformOverviewWorkspace()`，并且只被 `views/platform/PlatformOverview.vue` 使用，已经符合 `feature-owned UI` 的判定条件。
- `PlatformOverviewHeroPanel.vue`、`PlatformOverviewAlertsSection.vue`、`PlatformOverviewHotspotsSection.vue` 已经是稳定子区块，因此这次可以只迁移 page shell owner，不必顺手重排这些展示分区。

## Decision
refactor_existing

## Reason
这次不是新增平台总览能力，而是沿着已经写入事实源的 `feature-owned UI` 规则继续收口 legacy component page。最小正确改动是把 `PlatformOverviewPage.vue` 从 `components/platform/dashboard/` 迁到 `features/platform-overview/ui/`，让 route view 继续只组合 feature page model 与 feature ui，同时移除它对应的 component->feature 和 legacy component page 例外。

## Files to modify
- .harness/reuse-decisions/platform-overview-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-platform-overview-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-platform-overview-feature-ui-migration-review.md
- code/frontend/src/features/platform-overview/index.ts
- code/frontend/src/features/platform-overview/ui/*
- code/frontend/src/views/platform/PlatformOverview.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts
- code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果 `PlatformOverviewPage` 能顺利收口，下一批同模式候选可以继续看 `TeacherDashboardPage.vue` 和 `ContestOrchestrationPage.vue`。
- 本轮只处理平台总览这个 page shell，不额外把 dashboard 子分区或其它 admin 页面一起迁移。
