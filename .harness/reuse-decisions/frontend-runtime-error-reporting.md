# Reuse Decision

## Change type
utility / feature / test

## Existing code searched
- `code/frontend/src/features`
- `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/utils`
- `code/frontend/src/composables`

## Similar implementations found
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`

## Decision
refactor_existing

## Reason
这次不引入埋点 SDK，也不把所有错误处理逻辑推回 request 层。当前问题是多个 page owner 和 shared workspace 直接散落 `console.error`，而登录页还保留了装饰性 `console.log`。更低风险的收口方式是新增一个轻量的前端错误输出 helper，只保留开发环境日志出口，再让当前这些页面 owner 统一调用它；用户可见反馈仍由各自的 toast / error state 负责。这样既不改变现有交互分支，也能先把运行时噪音和错误输出 owner 收成单点。

## Files to modify
- `code/frontend/src/utils/reportFrontendError.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentListQuery.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/teacher-dashboard/model/useTeacherDashboardPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewIndex.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchive.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`

## After implementation
- 后续如果项目需要接真实埋点或前端告警，再让 `reportFrontendError` 成为统一接入点，不再回到页面里直接写原始 `console.error`。
