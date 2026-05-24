# Reuse Decision

## Change type
hook / page / layout

## Existing code searched
- `code/frontend/src/views/platform`
- `code/frontend/src/views/teacher`
- `code/frontend/src/features/platform-users`
- `code/frontend/src/features/class-students-workspace`
- `code/frontend/src/features/student-analysis-workspace`
- `code/frontend/src/features/student-review-archive-workspace`
- `code/frontend/src/features/teacher-class-management`
- `code/frontend/src/features/teacher-class-students`
- `code/frontend/src/features/teacher-student-analysis`
- `code/frontend/src/features/teacher-student-review-archive`
- `code/frontend/src/features/teacher-instances`
- `code/frontend/src/features/student-directory`
- `code/frontend/src/components/platform`
- `code/frontend/src/composables`
- `harness/prompts`

## Similar implementations found
- `code/frontend/src/features/platform-users/model/usePlatformUsers.ts`
- `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-class-management/model/useTeacherClassManagementPage.ts`
- `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
- `code/frontend/src/features/teacher-class-workspace/model/useTeacherClassWorkspaceSection.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewDetail.ts`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue`
- `code/frontend/src/features/teacher-instances/model/useTeacherInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-student-management/model/useTeacherStudentManagementPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchive.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`
- `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentListQuery.ts`
- `harness/prompts/architecture-diagram-generation.md`
- `harness/prompts/coding-agent-system-prompt.md`

## Decision
refactor_existing

## Reason
这次不新增并行实现。`platform-users` 已经承担了用户治理之外的班级、学生目录和实例管理职责，适合把现有页面模型迁到更细的 feature owner 下，而不是继续往该桶里加代码。学生目录查询能力继续复用 `student-directory` 里的既有 query owner；平台侧的 `usePlatformStudentManagementPage` 只保留页面级筛选、分页和路由跳转编排，不再新造一条与 `useTeacherStudentManagementPage`、`useTeacherStudentAnalysisPage`、`useStudentListQuery` 平行的学生查询栈。继续收口 `/platform/classes/:className`、`/platform/classes/:className/students/:studentId` 与 `/platform/classes/:className/students/:studentId/review-archive` 时，也不复制一份 `TeacherClassStudents.vue`、`TeacherStudentAnalysis.vue`、`TeacherStudentReviewArchive.vue` 或对应 page hook；改为把实际 page workflow 上提到中性的 `class-students-workspace`、`student-analysis-workspace`、`student-review-archive-workspace` feature，再让 teacher / platform 各自保留 route view owner。平台班级工作台别名页也沿用同一模式：不保留 `platformRoutes.ts -> TeacherClassWorkspaceSection.vue` 的跨 namespace 依赖，不复制一份新的 panel redirect 逻辑，而是把 alias route 到 canonical workspace 的重定向 owner 上提到 `class-workspace-redirect`，让 teacher 侧 `useTeacherClassWorkspaceSection` 只保留兼容桥，平台侧 route view 直接复用 `PlatformClassStudents.vue` 作为 canonical workspace 页。`PlatformAwdReviewDetail` 也继续按相同思路处理：不保留 `platformRoutes.ts -> TeacherAWDReviewDetail.vue` 的直连，不额外复制一套 AWD detail widget 或导出流程，而是把页面级路由、轮次切换和返回索引 owner 上提到中性的 `awd-review-detail-workspace`，继续复用 `TeacherAWDReviewWorkspace` 和 `useTeacherAwdReviewExportFlow`，让 teacher 侧 detail hook 退成兼容桥。`useStudentReviewArchivePage` 也不是新造一套导出/归档实现：它继续复用 `useTeacherStudentReviewArchive` 作为数据 owner，复用现有 review-archive 导出消息约定，并沿用与 `useReviewArchiveExportFlow`、`useTeacherAwdReviewExportFlow` 一致的“页面 owner 编排 + 导出 helper 复用”模式，只把 route 级导航和页面流程从 teacher 命名空间上提到中性 feature。review 提示词也放到现有 `harness/prompts/` 目录，复用已有项目 prompt 资产形态，不新建额外 prompt 目录。

## Files to modify
- `harness/prompts/AGENTS.md`
- `harness/prompts/frontend-architecture-review.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/views/platform/ClassManage.vue`
- `code/frontend/src/views/platform/StudentManage.vue`
- `code/frontend/src/views/platform/InstanceManage.vue`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/features/platform-users/index.ts`
- `code/frontend/src/features/platform-users/model/index.ts`
- `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform-class-management/index.ts`
- `code/frontend/src/features/platform-class-management/model/index.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/index.ts`
- `code/frontend/src/features/platform-student-management/model/index.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/index.ts`
- `code/frontend/src/features/platform-instance-management/model/index.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/features/class-students-workspace/index.ts`
- `code/frontend/src/features/class-students-workspace/model/index.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `docs/plan/impl-plan/2026-05-24-platform-route-owner-decoupling-implementation-plan.md`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/features/student-analysis-workspace/index.ts`
- `code/frontend/src/features/student-analysis-workspace/model/index.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`
- `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
- `code/frontend/src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts`
- `code/frontend/src/features/student-review-archive-workspace/index.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/index.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchivePage.ts`
- `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `code/frontend/src/features/class-workspace-redirect/index.ts`
- `code/frontend/src/features/class-workspace-redirect/model/index.ts`
- `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `code/frontend/src/features/teacher-class-workspace/model/useTeacherClassWorkspaceSection.ts`
- `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/features/awd-review-detail-workspace/index.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/index.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewDetail.ts`

## After implementation
- 如果后续 review 流程继续稳定复用，再考虑补 `feedback/` 记录 prompt 资产化经验；本次先以 `harness/prompts/` 作为唯一事实源。
