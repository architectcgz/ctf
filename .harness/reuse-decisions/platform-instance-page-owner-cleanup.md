# Reuse Decision

## Change type
frontend refactor / platform instance page owner cleanup

## Existing code searched
- `code/frontend/src/pages/platform/InstanceManageRoutePage.vue`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform/instance-management/ui/InstanceManageHeroPanel.vue`
- `code/frontend/src/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `TeacherInstanceManagementPage.vue` 已经把教师实例页的 page shell 收口在 feature 内部，route page 只保留组合。
- `TeacherStudentAnalysisRoutePage.vue`、`PlatformStudentAnalysisRoutePage.vue`、`TeacherDashboardRoutePage.vue` 等最近几轮都改成 route page 退回薄壳，feature page 自己组合 page model 与内部 panel。
- 当前 `InstanceManageRoutePage.vue` 仍直接组合 `InstanceManageHeroPanel`、`InstanceManageWorkspacePanel` 和 `usePlatformInstanceManagementPage()`，是同批 platform 目录页里剩余的 route-page owner 残片。

## Decision
refactor_existing

## Reason
下一轮最小正确切片不是继续改实例目录 query / destroy workflow，而是把平台实例页的 page owner 对齐到已采用的 feature-page 模式：

- 在 `features/platform/instance-management/ui` 下新增 `PlatformInstanceManagementPage.vue`，让 feature 自己组合 page model、hero panel 和 workspace panel。
- `InstanceManageRoutePage.vue` 退回只渲染 feature page 的薄壳，不再直接 import page model 或内部 panel。
- 不改实例目录加载、销毁、筛选、分页、route target 或 managed-instance 共享 workflow owner。

## Files to modify
- `.harness/reuse-decisions/platform-instance-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-platform-instance-page-owner-cleanup-plan.md`
- `code/frontend/src/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue`
- `code/frontend/src/features/platform/instance-management/ui/index.ts`
- `code/frontend/src/pages/platform/InstanceManageRoutePage.vue`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 平台实例页会和教师实例页一样，由 feature page 持有 page shell owner。
- route page 只保留 route-level 渲染职责，不再直接 import feature 内部 panel 或 page model。
