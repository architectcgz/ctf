# Reuse Decision

## Change type
frontend refactor / class students route owner cleanup

## Existing code searched
- code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts
- code/frontend/src/features/class-students-workspace/model/useClassWorkspaceSection.ts
- code/frontend/src/views/teacher/TeacherClassStudents.vue
- code/frontend/src/views/platform/PlatformClassStudents.vue
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts
- code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/features/teacher-student-management/model/teacherStudentManagementRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeQueryTransport.ts`
- `composables/routeNavigationTransport.ts`
- `features/teacher-student-management/model/teacherStudentManagementRoutes.ts`

## Decision
refactor_existing

## Reason
`useClassStudentsPage.ts` 现在把三类 router 职责混在一起：

- 读取 `className` params 和 `from_date / to_date` query
- alias route 命中的 canonical redirect
- 打开班级管理、教学概览、学员分析 3 条薄导航

如果继续把 `vue-router` 留在 page model，`featureRouterImportAllowlist` 不会下降；但如果为了消条目把这些职责再拆成新的 route wrapper，又会制造一层只服务单页的中间态。更合理的收口方式是：

- route `name / params / query` 读侧下沉到共享 transport
- `push / replace` 下沉到共享 navigation transport
- 班级页自己的导航目标抽成本地 route target helper
- alias redirect、时间窗口 query owner、班级工作区加载 owner 继续留在 `useClassStudentsPage.ts`

这样既能真实减少 allowlist，又不会把班级工作区自己的 query/redirect 业务规则堆进 shared。

## Files to modify
- .harness/reuse-decisions/class-students-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-class-students-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-class-students-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/features/class-students-workspace/model/classStudentsRoutes.ts
- code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts
- code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts

## After implementation
- `useClassStudentsPage.ts` 不再 import `vue-router`
- 班级工作区的 alias redirect、insight window query owner、列表加载 owner 继续留在 page model
- 班级管理 / 教学概览 / 学员分析 走本地 route target helper + shared navigation transport
- `featureRouterImportAllowlist` 再收掉 `features/class-students-workspace/model/useClassStudentsPage.ts -> vue-router`
