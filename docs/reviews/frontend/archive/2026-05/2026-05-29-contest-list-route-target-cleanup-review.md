# Contest List Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-contest-list-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/contest-list-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-contest-list-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-contest-list-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
  - `code/frontend/src/features/contest-detail/model/contestListRoutes.ts`
  - `code/frontend/src/features/contest-detail/model/index.ts`
  - `code/frontend/src/features/contest-detail/index.ts`
  - `code/frontend/src/views/contests/ContestList.vue`
  - `code/frontend/src/views/contests/__tests__/ContestList.test.ts`
- Classification check：同意按单条 feature route target cleanup 处理；`useContestListPage.ts` 的薄导航不再值得继续保留 `vue-router` 依赖。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useContestListPage.ts` 保留竞赛目录查询、筛选和分页 owner 是合理的，本轮不应把这些数据逻辑往 view 里回退。
- 竞赛详情入口是典型薄导航，收口成 route target contract 后，`ContestList.vue` 的导航边界会更清楚。
- 本轮的关键不是“少一层函数”，而是删除一条没有必要继续存在的 `feature -> vue-router` allowlist。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestList.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-list-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-list-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-list-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/contest-detail/model/useContestListPage.ts code/frontend/src/features/contest-detail/model/contestListRoutes.ts code/frontend/src/features/contest-detail/model/index.ts code/frontend/src/features/contest-detail/index.ts code/frontend/src/views/contests/ContestList.vue code/frontend/src/views/contests/__tests__/ContestList.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `ContestDetail` 自身的 route/query owner 仍在 `useContestDetailRoutePage()`，不属于这轮范围。
- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
