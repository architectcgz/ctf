# Student Dashboard Data Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-student-dashboard-data-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/student-dashboard-data-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-student-dashboard-data-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-student-dashboard-data-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts`
  - `code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts`
  - `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- Classification check：预期按单条 feature router owner cleanup 处理；`useStudentDashboardData.ts` 不属于合理的 route-aware feature owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useStudentDashboardPage.ts` 继续持有 `useRouter()` 合理，因为它本来就是 page owner。
- `useStudentDashboardData.ts` 应该只保留 data loading / display derivation owner，不应直接认识 `Router`。
- 这条 allowlist 如果继续保留，会给 data/helper 层混入 router 开后门；本轮应当在 touched surface 内收掉。
- 当前实现已经把 teacher/admin redirect 收回 page owner，同时保留 data 层的 redirect signal 与 load guard，避免非 student 角色继续命中 dashboard API。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/dashboard/__tests__/DashboardView.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-dashboard-data-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-student-dashboard-data-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-dashboard-data-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有大量条目，需要后续继续按“page owner / non-page owner”逐条判定。
