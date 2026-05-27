# Reuse Decision

## Change type
+feature / widget / page / composition

## Existing code searched
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/teacher-class-insight-window/**`
- `code/frontend/src/features/teacher-student-analysis/**`
- `code/frontend/src/features/teacher-student-review-archive/**`
- `code/frontend/src/widgets/teacher-review-archive/**`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `class-students-workspace`、`student-analysis-workspace`、`student-review-archive-workspace` 已经是 teacher / platform 共用的中立 page workflow owner，说明这轮不需要复制一套 admin 专属页面逻辑。
- `api/admin/teaching.ts` 刚刚证明仓库接受“保留底层实现不动，先通过更准确的 public owner 隔离调用面”的收口方式。
- `docs/plan/impl-plan/2026-05-24-feature-public-api-deep-import-cleanup-implementation-plan.md` 已经采用过“补齐中立公共出口，而不是继续让调用方深依赖 donor feature 内部实现”的路径。

## Decision
refactor_existing

## Reason
- 这轮目标不是重写 teacher / admin 页面，也不是把所有 teacher 命名目录一次性改名，而是先停止 `/platform/*` 和共享 workspace 继续直连 teacher 语义 owner。
- 当前最小正确路径是：为已经共享的 class insight、student analysis review、student review archive、review archive widget 补中立 public owner，再让 teacher / platform route view 与共享 workspace 统一改走这些中立入口。
- 这样可以把风险限制在 import owner 和边界命名层，不改 API contract、不改页面模板结构、不改异步工作流实现，也不复制并行 hook。

## Files to modify
- `.harness/reuse-decisions/admin-teacher-structure-coupling-alignment.md`
- `docs/plan/impl-plan/2026-05-27-admin-teacher-structure-coupling-alignment-implementation-plan.md`
- `code/frontend/src/features/class-insight-window/index.ts`
- `code/frontend/src/features/class-insight-window/model/index.ts`
- `code/frontend/src/features/student-analysis-review/index.ts`
- `code/frontend/src/features/student-analysis-review/model/index.ts`
- `code/frontend/src/features/student-review-archive/index.ts`
- `code/frontend/src/features/student-review-archive/model/index.ts`
- `code/frontend/src/widgets/review-archive-workspace/index.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
- `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- 相关 platform / teacher 测试
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-admin-teacher-structure-coupling-alignment-review.md`

## After implementation
- 如果这组“共享 workflow 先补中立 public owner，再让 route view 统一改走中立入口”的模式稳定，可以后续继续复用到 `PlatformClassWorkspaceSection` 和 `PlatformAwdReviewDetail` 仍残留的 teacher 语义 owner 收口。
