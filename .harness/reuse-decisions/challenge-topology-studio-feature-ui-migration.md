# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue
- code/frontend/src/views/platform/ChallengeTopologyStudio.vue
- code/frontend/src/features/challenge-topology-studio/index.ts
- code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `features/platform-overview/ui/*`、`features/teacher-dashboard/ui/*`、`features/platform-contests/ui/*` 已证明，单一 feature 的 page-sized UI 可以直接挂到 `features/*/ui`，route view 保持薄壳。
- `ChallengeTopologyStudioPage.vue` 当前只服务 `views/platform/ChallengeTopologyStudio.vue`，并直接消费 `features/challenge-topology-studio` 的 page model，已经满足 `feature-owned UI` 判定。
- 这页和 `ContestOrchestrationPage.vue` 不同，它自身没有持有 `vue-router`，因此这轮不需要额外迁 route owner，只需做 page shell 目录归位和 guardrail 收口。

## Decision
refactor_existing

## Reason
这次不是新增拓扑编辑能力，而是继续收口 `components/*Page.vue -> @/features/*` 的遗留例外。最小正确改动是把 `ChallengeTopologyStudioPage.vue` 从 `components/platform/topology/` 迁到 `features/challenge-topology-studio/ui/`，让 route view 继续只组合 route model 与 feature ui，同时移除它对应的 component->feature 和 legacy component page 例外。

## Files to modify
- .harness/reuse-decisions/challenge-topology-studio-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-challenge-topology-studio-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-challenge-topology-studio-feature-ui-migration-review.md
- code/frontend/src/features/challenge-topology-studio/index.ts
- code/frontend/src/features/challenge-topology-studio/ui/*
- code/frontend/src/views/platform/ChallengeTopologyStudio.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果 `ChallengeTopologyStudioPage` 顺利收口，下一批同模式低风险候选可以继续看 `UserGovernancePage.vue` 这类已明显归属于单一 feature 的 page shell。
- 本轮不顺手重排 topology 子分区，也不处理 `platform-challenges` route owner。
