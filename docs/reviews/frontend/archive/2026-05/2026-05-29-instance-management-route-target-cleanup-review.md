# Instance Management Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-instance-management-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/instance-management-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-instance-management-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-instance-management-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-instance-management/model/platformInstanceManagementRoutes.ts`
  - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
  - `code/frontend/src/features/teacher-instances/model/teacherInstanceManagementRoutes.ts`
  - `code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts`
  - `code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue`
  - `code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue`
  - `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue`
  - `code/frontend/src/views/platform/InstanceManage.vue`
  - `code/frontend/src/views/teacher/InstanceManagement.vue`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
  - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- Classification check：同意按“platform / teacher 实例目录的薄导航 route target cleanup”处理；两条 page model 的 router 依赖都只是目录内单次跳转，不应继续保留为 route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `usePlatformInstanceManagementPage.ts` 现在只保留平台实例目录的数据、筛选、分页、销毁与刷新 owner；“返回概览”与“所属用户 -> 学员分析”路由改成 `platformOverviewRoute` 与 `platformInstanceStudentAnalysisRoute()` contract。
- `useInstanceManagementPage.ts` 现在只保留教师实例目录的筛选、分页、销毁与初始化 owner；“返回教学概览”路由改成 `teacherInstanceDashboardRoute()` contract。
- 平台实例目录的 `InstanceManageHeroPanel.vue` / `InstanceManageWorkspacePanel.vue` 和教师实例目录的 `TeacherInstanceHeroPanel.vue` 都已改为通过共享 `AppRouteLink.vue` 消费 route target，没有把实例销毁、刷新、筛选和分页 owner 一起迁走。
- `featureRouterImportAllowlist` 已从实例目录这一组净减少 2 条。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/instance-management-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-instance-management-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-instance-management-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-instance-management/model/platformInstanceManagementRoutes.ts code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts code/frontend/src/features/teacher-instances/model/teacherInstanceManagementRoutes.ts code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue code/frontend/src/views/platform/InstanceManage.vue code/frontend/src/views/teacher/InstanceManagement.vue code/frontend/src/views/platform/__tests__/InstanceManage.test.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `useInstances.ts` 和其它仍在 allowlist 里的 route-aware page owner 不在这轮范围内。
