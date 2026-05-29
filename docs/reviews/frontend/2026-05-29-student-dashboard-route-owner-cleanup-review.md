# Student Dashboard Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-student-dashboard-route-owner-cleanup-plan.md`
- Scope：
  - `useStudentDashboardPage.ts`
  - `studentDashboardRoutes.ts`
  - `routeNavigationTransport.ts`
  - `useStudentDashboardPageBoundary.test.ts`
  - `DashboardView.test.ts`
- Classification check：同意按“student dashboard route owner cleanup”单独切片；这条同时触及 query tab owner、薄导航和 role redirect，不适合和 `contest-detail` 或 `audit-log` 混到一片里做。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `student-dashboard` 这条不需要为了清 allowlist 把所有 CTA 都翻成 link-first UI；当前 panel 仍是按钮交互，保留 callback contract 更符合最小切片。
- 真正该收口的是 page model 对 `vue-router` 的直接依赖：query tab owner 继续留在 page，薄导航与 role redirect 则改成 route target + transport。
- 新增的 `routeNavigationTransport.ts` 目前只承接 `push / replace`，没有吞进 dashboard-specific role 判定和 mounted policy，这个边界是合理的。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-dashboard-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-student-dashboard-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-dashboard-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeNavigationTransport.ts code/frontend/src/features/student-dashboard/model/studentDashboardRoutes.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `routeNavigationTransport.ts` 现在还是非常薄的 router transport；如果后续继续增长成 query normalize、role gate 或 redirect policy owner，就该回到 feature 本地再拆，而不是继续堆进 transport。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
