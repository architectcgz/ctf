# Reuse Decision

## Change type
+api / feature / component / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/teacher-workspace/model/useWorkspace.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
- `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 题解域已经按最小切片完成 `WriteupSubmissionItemData`、manual review DTO、`WriteupSubmissionDetailData` 的中性化命名，说明当前可以继续沿“共享 contract 一组一刀”的方式推进。
- `TeacherClassItem` 当前同时被 teacher、platform、admin 通知发布和 teacher workspace 共享消费，已经不是教师专属 contract。

## Decision
refactor_existing

## Reason
- `TeacherClassItem` 是当前更深层残余命名里最清楚的一组共享 DTO：它只承载班级目录项语义，不应继续挂 teacher 前缀。
- 这次只收“班级目录项”本身，不顺手改 `TeacherInstanceItem`、`TeacherStudentItem` 或其他班级洞察 DTO，能把 blast radius 控制在可审阅范围内。

## Files to modify
- `.harness/reuse-decisions/class-directory-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-class-directory-contract-naming-neutralization-implementation-plan.md`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/teacher-workspace/model/useWorkspace.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
- `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-class-directory-contract-naming-neutralization-review.md`

## After implementation
- 这组班级目录 contract 收口后，当前更深层 teacher 前缀共享 contract 将主要收敛到 `TeacherAWDReviewContestItemData` 和 `TeacherAttackSessionQuery`。
