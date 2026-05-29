# Contest Detail Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-contest-detail-route-owner-cleanup-plan.md`
- Scope：
  - `routeQueryTransport.ts`
  - `useContestDetailRoutePage.ts`
  - `ContestDetail.vue`
  - `ContestDetail.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“contest detail route owner cleanup”单独切片；这条同时触及 route param、query sync、AWD 默认页签和 tab owner，不适合和其它 contest surface 混做。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `contest-detail` 这条不需要新增 wrapper，也没必要为了清 allowlist 把 route state 平移回 route view。
- 更合理的 owner 分层是：
  - `useContestDetailRoutePage.ts` 继续持有 contest-specific route policy
  - `useRouteQueryTabs()` 接住 tab query transport
  - `routeQueryTransport.ts` 只暴露 `params / query / replaceQuery`
- AWD 默认页签现在会真实写回 query，这比旧的 `window.history` 同步更一致，也更容易被 router/test 观察到。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-contest-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeQueryTransport.ts code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts code/frontend/src/views/contests/ContestDetail.vue code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `routeQueryTransport.ts` 现在从“query only”扩成了“params + query transport”；如果后续继续增长，仍要守住“只做 transport，不做业务 normalize”这条边界。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
