> 状态：Current
> 事实源：`class-students-workspace`、`student-analysis-workspace`、`student-review-archive-workspace`、相关 teacher feature/widget 代码、architecture review 与 backlog
> 替代：无

# Admin Teacher Structure Coupling Alignment Implementation Plan

## 目标

- 继续收口 `/platform/*` 与 teacher 语义 owner 的结构耦合，优先把已经共享的 class students、student analysis、student review archive 三条 page workflow 改到中立 public owner。
- 让 teacher / platform route view 和共享 workspace 不再直接 import `teacher-*` 语义 feature / widget。
- 保持现有 API、页面模板结构和异步流程行为不变，只收口 public owner 与引用边界。

## 非目标

- 本轮不处理 `PlatformClassTrend`、`PlatformClassReview`、`PlatformClassInsights`、`PlatformClassIntervention` 仍共享的班级工作台别名页。
- 本轮不处理 `PlatformAwdReviewDetail`、`AWDReviewIndex`、`ChallengeWriteupManagePanel` 等仍直接依赖 `@/api/teaching` 的其他 admin / teacher 耦合面。
- 本轮不重命名 `api/contracts.ts` 里的 `Teacher*` 类型名，也不改后端接口路径。
- 本轮不删除旧的 `teacher-*` feature / widget 目录；它们先退成兼容桥，避免扩大文件迁移面。

## 输入依据

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
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `PlatformClassStudents.vue`、`PlatformStudentAnalysis.vue` 已经各自有独立 platform route view，当前主要残留不是“platform 还在直接路由到 teacher view”，而是共享 workflow 仍继续 import teacher 语义 owner。
- 这组残留主要集中在四个共享 owner：
  - `teacher-class-insight-window`
  - `teacher-student-analysis`
  - `teacher-student-review-archive`
  - `widgets/teacher-review-archive`
- 最小安全切片不是整目录搬迁，而是先补中立 public owner，再让共享 workspace 和 route view 统一改走新入口。

## 任务切片

### Slice 1：补共享 workflow 的中立 public owner

- 目标：
  - 为 class insight、student analysis review、student review archive data、review archive workspace 建立中立入口。
- 预期改动：
  - `code/frontend/src/features/class-insight-window/index.ts`
  - `code/frontend/src/features/class-insight-window/model/index.ts`
  - `code/frontend/src/features/student-analysis-review/index.ts`
  - `code/frontend/src/features/student-analysis-review/model/index.ts`
  - `code/frontend/src/features/student-review-archive/index.ts`
  - `code/frontend/src/features/student-review-archive/model/index.ts`
  - `code/frontend/src/widgets/review-archive-workspace/index.ts`
- review focus：
  - 只暴露现有稳定符号，不复制实现
  - 新入口命名能被 teacher / platform 同时接受

### Slice 2：让共享 workspace 与 route view 切到中立入口

- 目标：
  - 让共享 workflow owner 和 teacher / platform route view 统一改走中立 public owner。
- 预期改动：
  - `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
  - `code/frontend/src/features/teacher-class-report-export/model/useClassReportExport.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
  - `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
  - `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- review focus：
  - route view 仍只保留 page shell
  - page-level async owner 不漂移
  - 平台页不再直接引用 teacher 语义 owner

### Slice 3：同步测试与 backlog / review

- 目标：
  - 更新 source-string 断言、共享 workspace 断言和 backlog 状态说明。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts`
  - 如有必要，补充其他受影响的 raw source tests
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-admin-teacher-structure-coupling-alignment-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts`
- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退粒度按 public owner 切：可以整体回退新增的 neutral index 与对应 import 切换，不涉及接口、数据迁移或模板结构回退。

## 残余风险

- 旧 `teacher-*` 目录仍保留为兼容桥，因此 teacher 语义实现目录并没有在本轮物理消失；本轮优先解决的是 `/platform/*` 和共享 workspace 的直接依赖面。
- `PlatformClassWorkspaceSection`、`PlatformAwdReviewDetail` 和其他仍直接依赖 `@/api/teaching` 的 admin surface 还需要后续独立切片继续收口。
