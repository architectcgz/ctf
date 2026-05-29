# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts
- code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts
- code/frontend/src/views/platform/StudentManage.vue
- code/frontend/src/views/teacher/TeacherStudentManagement.vue
- code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue
- code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue
- code/frontend/src/views/platform/__tests__/StudentManage.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/features/platform-class-management/model/platformClassManagementRoutes.ts
- code/frontend/src/features/teacher-class-management/model/teacherClassManagementRoutes.ts

## Similar implementations found
- `class-management-route-target-cleanup` 已示范“page model 只产出 route target contract，UI 通过共享 `AppRouteLink` 消费”的模式。
- `platformOverviewRoutes.ts` 已示范同一 feature 内将薄导航 owner 收口为本地 route helper。

## Decision
refactor_existing

## Reason
这轮的两条 allowlist 也属于“数据 owner + 顺手 push”的薄导航 wrapper：

- `usePlatformStudentManagementPage.ts` 只剩 `openStudent(studentId)` 一个单次跳转。
- `useStudentManagementPage.ts` 里 `openClassManagement()` 与 `openStudent(studentId)` 都只是单次跳转，本地 `reportDialogVisible` 才是它真正该保留的 workflow owner。

最小正确改动是：

- 给 `platform-student-management`、`teacher-student-management` 各自补本地 route target helper
- 两个 page model 去掉 `vue-router`，改成返回 route target contract
- 平台 / 教师学生目录 UI 改为通过共享 `AppRouteLink` 消费 route target
- 保留教师端“导出班级报告”弹窗 owner，不把非路由 workflow 一起迁走

这样可以一次收掉：

- `features/platform-student-management/model/usePlatformStudentManagementPage.ts -> vue-router`
- `features/teacher-student-management/model/useStudentManagementPage.ts -> vue-router`

本轮不做：

- 不处理 `useStudentAnalysisPage.ts` 或 `useClassManagementPage.ts` 等其它 route-aware page owner
- 不改学生目录的数据拉取、筛选、分页和报告导出流程
- 不继续做学生目录的大组件拆分

## Files to modify
- .harness/reuse-decisions/student-management-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-student-management-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-student-management-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-student-management/model/platformStudentManagementRoutes.ts
- code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts
- code/frontend/src/features/teacher-student-management/model/teacherStudentManagementRoutes.ts
- code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts
- code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue
- code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue
- code/frontend/src/views/platform/StudentManage.vue
- code/frontend/src/views/teacher/TeacherStudentManagement.vue
- code/frontend/src/views/platform/__tests__/StudentManage.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts

## After implementation
- `usePlatformStudentManagementPage.ts` 不再 import `vue-router`
- `useStudentManagementPage.ts` 不再 import `vue-router`
- 平台 / 教师学生目录都改成 route target contract + 共享 `AppRouteLink`
- `featureRouterImportAllowlist` 再减少 2 条
