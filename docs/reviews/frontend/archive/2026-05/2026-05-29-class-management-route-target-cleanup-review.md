# Class Management Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-class-management-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/class-management-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-class-management-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-class-management-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components/navigation/AppRouteLink.vue`
  - `code/frontend/src/features/platform-class-management/model/platformClassManagementRoutes.ts`
  - `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
  - `code/frontend/src/components/teacher/class-management/TeacherClassManagementHeaderActions.vue`
  - `code/frontend/src/components/teacher/class-management/TeacherClassManagementRowLink.vue`
  - `code/frontend/src/features/teacher-class-management/model/teacherClassManagementRoutes.ts`
  - `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
  - `code/frontend/src/views/platform/ClassManage.vue`
  - `code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue`
  - `code/frontend/src/features/teacher-class-management/ui/ClassManagementPage.vue`
  - `code/frontend/src/views/teacher/ClassManagement.vue`
  - `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
  - `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- Classification check：同意按“platform / teacher 班级目录的薄导航 route target cleanup”处理；两条 page model 的 router 依赖都只是单次跳转，不应继续保留为 route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `usePlatformClassManagementPage.ts` 现在只保留班级目录数据、分页和错误态 owner；“查看班级”路由改成 `platformClassStudentsRoute()` contract。
- `useClassManagementPage.ts` 现在只保留教师班级目录数据、分页和 `reportDialogVisible` owner；“教学概览 / 进入班级”路由改成 `teacherDashboardRoute` 与 `teacherClassStudentsRoute()` contract。
- `AppRouteLink.vue` 现在集中承接声明式路由链接；`ClassManageWorkspacePanel.vue` 与教师目录动作子组件改为消费这层 bridge，而不是各自直接 import `vue-router`。教师端“导出班级报告”仍由 button + emit 驱动，没有把非路由 workflow 顺手下沉。
- `featureRouterImportAllowlist` 已从班级目录这一组净减少 2 条。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ClassManage.test.ts src/views/teacher/__tests__/ClassManagement.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-management-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-class-management-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-class-management-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components/navigation/AppRouteLink.vue code/frontend/src/features/platform-class-management/model/platformClassManagementRoutes.ts code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue code/frontend/src/components/teacher/class-management/TeacherClassManagementHeaderActions.vue code/frontend/src/components/teacher/class-management/TeacherClassManagementRowLink.vue code/frontend/src/features/teacher-class-management/model/teacherClassManagementRoutes.ts code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts code/frontend/src/features/teacher-class-management/ui/ClassManagementPage.vue code/frontend/src/views/platform/ClassManage.vue code/frontend/src/views/teacher/ClassManagement.vue code/frontend/src/views/platform/__tests__/ClassManage.test.ts code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- 班级工作区和教师总览等其它 route-aware page owner 不在这轮范围内。
