# Teacher Dashboard Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-teacher-dashboard-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/teacher-dashboard-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-teacher-dashboard-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-teacher-dashboard-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components/navigation/AppRouteLink.vue`
  - `code/frontend/src/features/teacher-dashboard/model/index.ts`
  - `code/frontend/src/features/teacher-dashboard/model/teacherDashboardRoutes.ts`
  - `code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts`
  - `code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue`
  - `code/frontend/src/views/teacher/TeacherDashboard.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- Classification check：同意按 `teacher-dashboard` feature 内“薄导航 route target cleanup”处理；`useDashboardPage.ts` 的 router 依赖只剩一次“进入班级管理”，不应继续保留为 reviewed route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `teacherDashboardRoutes.ts` 现在集中提供教师概览的“班级管理” route target contract，并复用既有 `resolveClassManagementRouteName()` 角色判定。
- `useDashboardPage.ts` 已退回纯教师概览数据加载、错误态和 retry owner，不再直接 import `vue-router`；`classManagementRoute` 改为 computed route target。
- `TeacherDashboardPage.vue` 已恢复真实可见的“班级管理”入口，并通过共享 `AppRouteLink.vue` 消费 route target；`retry` 仍由 button + emit 驱动，没有把非路由 workflow 混到 link contract 里。
- `TeacherDashboard.test.ts` 已从 mock `router.push()` 改成真实 router 导航断言，同时补上“page model 不再 import vue-router、page shell 直接消费 AppRouteLink”的 raw-source 护栏。
- `featureRouterImportAllowlist` 已再减少 1 条：`features/teacher-dashboard/model/useDashboardPage.ts -> vue-router`。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/teacher-dashboard-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-teacher-dashboard-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-teacher-dashboard-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components/navigation/AppRouteLink.vue code/frontend/src/features/teacher-dashboard/model/index.ts code/frontend/src/features/teacher-dashboard/model/teacherDashboardRoutes.ts code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue code/frontend/src/views/teacher/TeacherDashboard.vue code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `teacher-dashboard` 的 tab query sync 继续留在 `useUrlSyncedTabs()` 本地 window owner，这轮不处理。
