# Student Management Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-student-management-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/student-management-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-student-management-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-student-management-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-student-management/model/platformStudentManagementRoutes.ts`
  - `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
  - `code/frontend/src/features/teacher-student-management/model/teacherStudentManagementRoutes.ts`
  - `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
  - `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
  - `code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue`
  - `code/frontend/src/views/platform/StudentManage.vue`
  - `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
  - `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- Classification check：同意按“platform / teacher 学生目录的薄导航 route target cleanup”处理；两条 page model 的 router 依赖都只是目录内单次跳转，不应继续保留为 route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `usePlatformStudentManagementPage.ts` 现在只保留平台学生目录的数据、筛选、分页与错误态 owner；“查看学员”路由改成 `platformStudentAnalysisRoute()` contract，并随目录行一起下发给 workspace panel。
- `useStudentManagementPage.ts` 现在只保留教师学生目录的数据、筛选、分页与 `reportDialogVisible` owner；“班级管理 / 学员分析”路由分别改成 `teacherClassManagementRoute()` 与 `teacherStudentAnalysisRoute()` contract。
- 平台目录 `StudentManageWorkspacePanel.vue` 与教师目录 `StudentManagementPage.vue` 都已改为通过共享 `AppRouteLink.vue` 消费 route target，没有把教师端“导出班级报告”这种非路由 workflow 一起迁走。
- `featureRouterImportAllowlist` 已从学生目录这一组净减少 2 条。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-management-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-student-management-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-management-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-student-management/model/platformStudentManagementRoutes.ts code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts code/frontend/src/features/teacher-student-management/model/teacherStudentManagementRoutes.ts code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue code/frontend/src/views/platform/StudentManage.vue code/frontend/src/views/teacher/TeacherStudentManagement.vue code/frontend/src/views/platform/__tests__/StudentManage.test.ts code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `platform-instance-management`、`teacher-instances` 和其它仍在 allowlist 里的 route-aware page owner 不在这轮范围内。
