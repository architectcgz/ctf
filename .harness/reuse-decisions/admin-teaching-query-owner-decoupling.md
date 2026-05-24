# Reuse Decision

## Change type
api / hook / page

## Existing code searched
- `code/frontend/src/api/teacher.ts`
- `code/frontend/src/api/teacher`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/features/platform-class-management`
- `code/frontend/src/features/platform-student-management`
- `code/frontend/src/features/platform-instance-management`
- `code/frontend/src/features/class-students-workspace`
- `code/frontend/src/features/student-analysis-workspace`
- `code/frontend/src/features/student-review-archive-workspace`
- `code/frontend/src/features/awd-review-detail-workspace`
- `code/frontend/src/features/teacher-student-analysis`
- `code/frontend/src/features/teacher-student-review-archive`
- `code/frontend/src/features/teacher-class-report-export`
- `code/frontend/src/features/teacher-awd-review`
- `code/frontend/src/views/platform/__tests__`

## Similar implementations found
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherSubmissionReviewFlows.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchive.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewIndex.ts`

## Decision
refactor_existing

## Reason
这次不新增后台专用并行接口，也不改后端路由。现有 platform 班级、学生、实例管理已经在页面 owner 上拆开，但 query helper 仍然直接挂在 `@/api/teacher` 命名空间下。继续扩 `code/frontend/src/api/teacher/classes.ts`、`code/frontend/src/api/teacher/students.ts` 或 `code/frontend/src/api/teacher/awd-reviews.ts` 虽然能复用现有实现，但会把 platform 侧和共享 workspace 侧的新查询继续沉积在 teacher namespace 里，和这轮“去 teacher owner 耦合”的目标相反。更低风险的收口方式是把这些已经被 admin/teacher 共用的查询实现迁到中性的 `@/api/teaching` owner，同时保留 `@/api/teacher` 作为 teacher 侧原始 owner，不再增加兼容 re-export。这样既能保留现有请求路径和 DTO 形状，也能让 platform feature 以及中性 workspace helper 不再显式依赖 teacher namespace。

## Files to modify
- `code/frontend/src/api/teaching.ts`
- `code/frontend/src/api/teaching/index.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentListQuery.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherSubmissionReviewFlows.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherInterventionRecommendations.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchive.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewIndex.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- `docs/plan/impl-plan/2026-05-24-admin-teaching-query-owner-decoupling-implementation-plan.md`

## After implementation
- 如果后续还要继续把 class/student review、AWD review detail 等共享 workflow 从 `@/api/teacher` 收口到中性 owner，再沿用 `@/api/teaching` 扩展，不再新增新的 role-specific wrapper 或兼容桥接层。
