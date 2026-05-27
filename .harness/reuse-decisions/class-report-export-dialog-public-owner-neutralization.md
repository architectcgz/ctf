# Reuse Decision

## Change type
+component / view / docs / test

## Existing code searched
- `code/frontend/src/components/teacher/reports/index.ts`
- `code/frontend/src/components/teacher/reports/ClassReportExportDialog.vue`
- `code/frontend/src/views/teacher/ClassManagement.vue`
- `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `class-management`、`student-analysis-workspace`、`review-archive-workspace` 等共享页面已经通过中立 public owner 暴露给 teacher / platform 复用，说明共享班级报告导出对话框也适合继续按 public owner 收口。
- `ClassReportExportDialog` 当前同时被 teacher / platform 路由页使用，但公开入口仍挂在 teacher 目录，只是历史 owner 残留。

## Decision
refactor_existing

## Reason
- `ClassReportExportDialog` 是共享对话框，不是教师专属页面壳；继续从 `components/teacher/reports` 暴露，会让 platform 页面在结构上继续依赖 teacher 组件命名空间。
- 这次只加一个中立 public owner 并迁移相关 import / 测试，不移动实际 `.vue` 文件，也不改导出流程和 feature owner，能控制 blast radius。

## Files to modify
- `.harness/reuse-decisions/class-report-export-dialog-public-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-class-report-export-dialog-public-owner-neutralization-implementation-plan.md`
- `code/frontend/src/components/reports/index.ts`
- `code/frontend/src/views/teacher/ClassManagement.vue`
- `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-class-report-export-dialog-public-owner-neutralization-review.md`

## After implementation
- 这组收口后，teacher / platform 班级与学员分析路由页对共享班级报告导出对话框的 public import 将不再挂在 teacher 组件命名空间。
