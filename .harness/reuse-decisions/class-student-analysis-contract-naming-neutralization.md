# Reuse Decision

## Change type
frontend contract / api / feature / component / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/features/class-students-workspace/**`
- `code/frontend/src/features/student-analysis-workspace/**`
- `code/frontend/src/features/student-directory/**`
- `code/frontend/src/features/teacher-student-analysis/**`
- `code/frontend/src/features/teacher-class-report-export/**`
- `code/frontend/src/features/teacher-dashboard/**`
- `code/frontend/src/features/skill-profile/**`
- `code/frontend/src/components/teacher/**`
- `code/frontend/src/widgets/teacher-student-review-workspace/**`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/contracts/api-contract-v1.md`
- `.harness/reuse-decisions/class-directory-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-class-directory-contract-naming-neutralization-implementation-plan.md`
- `docs/reviews/frontend/2026-05-27-class-directory-contract-naming-neutralization-review.md`

## Similar implementations found
- 班级目录项 `TeacherClassItem -> ClassDirectoryItem` 已按“共享 contract 一组一刀”完成中性化，说明当前可以继续沿最小命名切片推进。
- `AttackSessionQuery` 已经先于 response DTO 完成中性化，说明 student analysis 这条共享链路仍有一组 teacher 前缀残片等待收口。
- `class-students-workspace`、`student-analysis-workspace`、platform route view 与 teacher route view 当前都通过同一组共享 feature 消费这些 DTO，更适合在 contract 层统一改名，而不是继续让 platform surface 共用 teacher 前缀。

## Decision
refactor_existing

## Reason
- 当前剩余 `P1` 中最直接、最可审的残留，是共享班级 / 学员分析 contract 仍保留 `Teacher*` 命名。
- 最小正确改动是把 shared class insight、student directory、student evidence、attack session 以及关联 query/payload 一并中性化，再同步 teaching API client 与共享 feature 消费面。
- 本轮不改 HTTP path、teacher / platform route path、teacher public wrapper 行为，也不触碰 teacher-only overview contract。

## Files to modify
- `.harness/reuse-decisions/class-student-analysis-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-28-class-student-analysis-contract-naming-neutralization-plan.md`
- `docs/reviews/frontend/2026-05-28-class-student-analysis-contract-naming-neutralization-review.md`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/features/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentListQuery.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useInterventionRecommendations.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/teacher-dashboard/model/teacherDashboardInsightBuilders.ts`
- `code/frontend/src/features/teacher-dashboard/model/teacherDashboardOverviewBuilders.ts`
- `code/frontend/src/features/teacher-dashboard/model/useDashboardMetrics.ts`
- `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
- `code/frontend/src/features/teacher-workspace/model/useWorkspace.ts`
- `code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/features/teacher-class-insight-window/model/window.ts`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/components/teacher/ClassInsightsPanel.vue`
- `code/frontend/src/components/teacher/ClassReviewPanel.vue`
- `code/frontend/src/components/teacher/ClassTrendPanel.vue`
- `code/frontend/src/components/teacher/InterventionPanel.vue`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsDirectoryPanel.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsOverviewPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `code/frontend/src/widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue`
- `code/frontend/src/widgets/teacher-student-review-workspace/model/presentation.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- shared class/student analysis feature 不再继续消费 `TeacherClass*`、`TeacherStudentItem`、`TeacherEvidence*`、`TeacherAttackSession*` 这组 teacher 前缀 contract。
- teacher / platform route view 与共享 feature 在 contract 层的残余 teacher 语义会进一步缩到 teacher-only overview DTO、teacher public wrapper 命名与后端 teacher transport path。
