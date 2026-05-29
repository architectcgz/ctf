# Contest Manage Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-contest-manage-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/contest-manage-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-contest-manage-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-contest-manage-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-contests/model/useContestManagePage.ts`
  - `code/frontend/src/features/platform-contests/model/contestManageRoutes.ts`
  - `code/frontend/src/features/platform-contests/model/index.ts`
  - `code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue`
  - `code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue`
  - `code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue`
  - `code/frontend/src/views/platform/ContestManage.vue`
  - `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
  - `code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- Classification check：同意按单条 feature route target cleanup 处理；`useContestManagePage.ts` 的三条导航都只是薄路由动作，不应继续保留 `vue-router` 依赖。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useContestManagePage.ts` 继续保留竞赛目录数据、公告抽屉、创建流程和 readiness override owner 是合理的，本轮不应把这些 workflow 往 view 或 link contract 里回退。
- 竞赛表格里的编辑 / 运维入口和公告抽屉里的完整页入口都是典型薄导航，收口成 route target contract 后，`ContestManage.vue` 的 owner 边界会更清楚。
- 本轮关键不是“减少几行代码”，而是删除一条不该继续存在的 `feature -> vue-router` allowlist，并让显式导航 contract 留在更合适的位置。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-manage-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-manage-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-manage-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-contests/model/useContestManagePage.ts code/frontend/src/features/platform-contests/model/contestManageRoutes.ts code/frontend/src/features/platform-contests/model/index.ts code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue code/frontend/src/views/platform/ContestManage.vue code/frontend/src/views/platform/__tests__/ContestManage.test.ts code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `useContestEditPage.ts`、`useContestOperationsPage.ts`、`useContestAnnouncementsPage.ts` 自身的 route owner 不在这轮范围。
- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
