# Contest Edit Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-contest-edit-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/contest-edit-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-contest-edit-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-contest-edit-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/features/platform-contests/model/contestManageRoutes.ts`
  - `code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts`
  - `code/frontend/src/features/platform-contests/model/useContestEditPage.ts`
  - `code/frontend/src/views/platform/ContestEdit.vue`
  - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
  - `code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue`
  - `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
  - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue`
  - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue`
  - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue`
  - `code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue`
  - `code/frontend/src/features/awd-readiness/ui/AWDReadinessChecklist.vue`
  - `code/frontend/src/components/navigation/AppRouteRedirect.vue`
- Classification check：同意按最后一条 `platform-contests` route target cleanup 处理；`useContestEditPage.ts` 当前的 router 依赖既包含薄导航也包含 mutation 后跳转，适合拆成 route target contract + 独立 navigation transport。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useContestEditPage.ts` 继续保留竞赛详情加载、AWD data refresh、stage 派生和保存 workflow owner 是合理的，本轮不应把这些逻辑回退到 view。
- 返回竞赛目录、公告页和 AWD 配置页这些入口如果改成 route target contract，`ContestEdit.vue` 的组合边界会比“接收一堆 emit 然后隐式 push”清楚得多。
- 保存成功后的“返回目录”不应被伪装成纯 UI 点击导航；把它改成独立 navigation transport 更接近真实 owner。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-edit-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-edit-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-edit-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/router/routes/platformRoutes.ts code/frontend/src/features/platform-contests/model/contestManageRoutes.ts code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts code/frontend/src/features/platform-contests/model/useContestEditPage.ts code/frontend/src/views/platform/ContestEdit.vue code/frontend/src/views/platform/__tests__/ContestEdit.test.ts code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue code/frontend/src/features/awd-readiness/ui/AWDReadinessChecklist.vue code/frontend/src/components/navigation/AppRouteRedirect.vue`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `useUrlSyncedTabs()` 和编辑页 stage query 行为不在本轮范围。
- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
