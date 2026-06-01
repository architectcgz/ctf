# Contest Operations Hub Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-contest-operations-hub-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/contest-operations-hub-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-contest-operations-hub-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-contest-operations-hub-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts`
  - `code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts`
  - `code/frontend/src/features/platform-contests/model/index.ts`
  - `code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue`
  - `code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue`
  - `code/frontend/src/views/platform/ContestOperationsHub.vue`
  - `code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
- Classification check：同意按单条 feature route target cleanup 处理；`useContestOperationsHubPage.ts` 的返回与进入动作都只是薄导航，不应继续保留 `vue-router` 依赖。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useContestOperationsHubPage.ts` 继续保留 AWD 赛事目录数据、分页、preferred contest 和错误态 owner 是合理的，本轮不应把这些逻辑往 view 里回退。
- hero、空态和目录行里的入口都只是薄导航，收口成 route target contract 后，`ContestOperationsHub.vue` 的 owner 边界会更清楚。
- 本轮关键不是减少几行 handler，而是删除一条没有必要继续存在的 `feature -> vue-router` allowlist。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-operations-hub-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-operations-hub-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-operations-hub-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts code/frontend/src/features/platform-contests/model/index.ts code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue code/frontend/src/views/platform/ContestOperationsHub.vue code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `useContestOperationsPage.ts` 自身的 route owner 不在这轮范围。
- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
