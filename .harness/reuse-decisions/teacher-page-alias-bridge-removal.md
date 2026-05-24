# Reuse Decision

## Change type
feature / view / test / docs

## Existing code searched
- `code/frontend/src/views/teacher/*.vue`
- `code/frontend/src/features/teacher-class-students/**`
- `code/frontend/src/features/teacher-student-analysis/**`
- `code/frontend/src/features/teacher-student-review-archive/**`
- `code/frontend/src/features/teacher-awd-review/**`
- `code/frontend/src/features/*-workspace/**`

## Similar implementations found
- `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewDetail.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`

## Decision
refactor_existing

## Reason
这些 teacher feature 文件已经不再承担真实 workflow，只是把 teacher 路由页再转发到中性 workspace owner。继续保留它们，会让 `views/teacher -> features/teacher-* -> features/*-workspace` 这条链路凭空多一层壳，既增加搜索噪音，也掩盖真正的共享 owner。更低风险的收口方式是让 teacher 路由页直接使用中性 workspace feature，并删除这些 page-level alias 壳；而 teacher feature 目录里仍承担共享逻辑的 helper 继续保留，不在本轮扩大范围。

## Files to modify
- `.harness/reuse-decisions/teacher-page-alias-bridge-removal.md`
- `docs/plan/impl-plan/2026-05-24-teacher-page-alias-bridge-removal-implementation-plan.md`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/features/teacher-class-students/**`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`
- `code/frontend/src/features/teacher-student-analysis/index.ts`
- `code/frontend/src/features/teacher-student-analysis/model/index.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-student-review-archive/model/index.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewDetail.ts`
- `code/frontend/src/features/teacher-awd-review/index.ts`
- `code/frontend/src/features/teacher-awd-review/model/index.ts`

## After implementation
- 后续只要 `teacher-*` / `platform-*` feature 退化成 page-level alias 壳，就按“视图直连中性 workspace owner + 删除 alias 文件”的模式继续收口，不再把 role 命名当成共享 feature 的事实 owner。
