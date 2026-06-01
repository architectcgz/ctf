# Class Students Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-class-students-route-owner-cleanup-plan.md`
- Scope：
  - `routeQueryTransport.ts`
  - `classStudentsRoutes.ts`
  - `useClassStudentsPage.ts`
  - `TeacherClassStudents.test.ts`
  - `PlatformClassStudents.test.ts`
  - `TeacherClassWorkspaceSection.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“class students route owner cleanup”单独切片；这条同时包含 params/query 读取、alias redirect 和薄导航，但都仍属于 page owner 自己的路由边界，不需要再造新 wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useClassStudentsPage.ts` 继续保留 alias redirect、时间窗口 query owner、班级工作区加载和 stale request owner 是合理的；这些都不是 shared transport 应该吞进去的业务规则。
- `routeQueryTransport.ts` 增加 `name` 仍然保持在纯 transport 边界内，没有混入 class students 的业务判断。
- 班级管理 / 教学概览 / 学员分析三条导航改成本地 `classStudentsRoutes.ts`，比继续让 page model 直接拼 `resolve*RouteName` 更清楚，也比再造 route wrapper 更小。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-students-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-class-students-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-class-students-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeQueryTransport.ts code/frontend/src/features/class-students-workspace/model/classStudentsRoutes.ts code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `routeQueryTransport.ts` 后续如果继续长出更多 route 元信息，仍要守住 transport 边界，不要把 role-aware route policy 平移进 shared。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
