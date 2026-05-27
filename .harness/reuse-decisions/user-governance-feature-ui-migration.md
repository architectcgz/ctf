# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/platform/user/UserGovernancePage.vue
- code/frontend/src/components/platform/user/UserGovernanceOverviewPanel.vue
- code/frontend/src/components/platform/user/UserGovernanceDetailModal.vue
- code/frontend/src/components/platform/user/UserGovernanceImportPanel.vue
- code/frontend/src/views/platform/UserManage.vue
- code/frontend/src/features/platform-user-management/index.ts
- code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts
- code/frontend/src/features/platform-user-management/model/usePlatformUsers.ts
- code/frontend/src/features/scoreboard/model/useScoreboardRoutePage.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts
- code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/UserManage.test.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `PlatformOverviewPage.vue`、`TeacherDashboardPage.vue`、`ChallengeTopologyStudioPage.vue` 已证明，单一 feature 的 page-sized UI 可以直接挂到 `features/*/ui`，route view 保持组合壳。
- `ContestOrchestrationPage.vue` 的迁移进一步证明，如果 page shell 自己还持有 router / query owner，应先把这部分收回 feature model，再迁到 `features/*/ui`。
- `useScoreboardRoutePage.ts`、`usePlatformChallengeDetailRoutePage.ts`、`useContestDetailRoutePage.ts` 提供了现成的 feature route owner 参照：router/query 同步不留在 route view，也不塞回 page-sized UI，而是作为 feature model 的一部分统一暴露。
- `UserGovernancePage.vue` 当前只服务 `UserManage.vue`，并直接围绕 `platform-user-management` 的用户治理 workflow 组装 overview / import 两个面板，符合 `feature-owned UI` 判定。

## Decision
refactor_existing

## Reason
这次不是新增用户治理能力，而是继续收口 `components/*Page.vue -> @/features/*` 的遗留例外。最小正确改动不是把 `useRoute/useRouter` 原样带进 feature ui，而是复用现有 `feature route owner` 模式，先给 `panel` query 补一个 feature model owner，再把 `UserGovernancePage.vue` 迁到 `features/platform-user-management/ui/`，同时移除它对应的 legacy component page 例外。

## Files to modify
- .harness/reuse-decisions/user-governance-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-user-governance-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-user-governance-feature-ui-migration-review.md
- code/frontend/src/features/platform-user-management/index.ts
- code/frontend/src/features/platform-user-management/model/index.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts
- code/frontend/src/features/platform-user-management/ui/*
- code/frontend/src/views/platform/UserManage.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/UserManage.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts
- code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts
- code/frontend/src/views/__tests__/journalNoteStyles.test.ts
- code/frontend/src/views/__tests__/pageTabsStyles.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果 `UserGovernancePage` 顺利收口，`feature-owned UI` 这条 backlog 会继续压缩到更少的遗留 teacher / student page shell。
- 本轮不顺手迁 `PlatformUserFormDialog.vue`，也不处理用户治理子分区目录归位。
