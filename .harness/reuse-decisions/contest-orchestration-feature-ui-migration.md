# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/features/platform-contests/model/useContestManagePage.ts
- code/frontend/src/features/platform-contests/model/usePlatformContests.ts
- code/frontend/src/features/platform-contests/index.ts
- code/frontend/src/components/platform/contest/PlatformContestTable.vue
- code/frontend/src/components/platform/contest/PlatformContestFormPanel.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/ContestManage.test.ts
- code/frontend/src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `features/platform-overview/ui/*` 与 `features/teacher-dashboard/ui/*` 已经证明，单一 feature 的 page-sized shell 可以直接迁到 `features/*/ui`，route view 只保留组合面。
- `ContestOrchestrationPage.vue` 当前只服务 `ContestManage.vue`，并直接消费 `features/platform-contests` contract，符合 `feature-owned UI` 判定。
- 这页和前两刀不同的地方是它还持有 `vue-router`。已有题解迁移已经证明，迁移前应先把导航 owner 收回 route-aware feature model，而不是把 router 一起搬进 feature ui。

## Decision
refactor_existing

## Reason
这次不是新增竞赛目录能力，而是继续收口 `components/*Page.vue -> @/features/*` 例外，同时修正 `ContestOrchestrationPage.vue` 持有 router 的 owner 漂移。最小正确改动是把竞赛编辑页 / 运维页跳转 owner 收回 `useContestManagePage()`，再把 page shell 迁到 `features/platform-contests/ui/`，移除它对应的 component->feature 和 legacy component page 例外。

## Files to modify
- .harness/reuse-decisions/contest-orchestration-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-contest-orchestration-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-contest-orchestration-feature-ui-migration-review.md
- code/frontend/src/features/platform-contests/index.ts
- code/frontend/src/features/platform-contests/model/useContestManagePage.ts
- code/frontend/src/features/platform-contests/ui/*
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/ContestManage.test.ts
- code/frontend/src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase21.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts
- code/frontend/src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts
- code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts
- code/frontend/src/views/__tests__/journalNoteStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果 `ContestOrchestrationPage` 顺利收口，下一批 feature-owned page shell 候选会继续收敛到其它仍留在 `components/**` 的单一 feature 页面。
- 本轮不把 `ContestAwdConfigWorkspaceShell.vue` 的超大组件拆分混进来；那是另一条技术债。
