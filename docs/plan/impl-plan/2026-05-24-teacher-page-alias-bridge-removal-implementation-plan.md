> 状态：Current
> 事实源：2026-05-24 前端架构 review、teacher route view 当前 alias 链路
> 替代：无

# Teacher Page Alias Bridge Removal Implementation Plan

## 目标

- 让 teacher route views 直接使用已经存在的中性 workspace feature owner。
- 删除 teacher feature 中只做 page-level alias 转发的历史桥接壳。

## 非目标

- 本轮不重命名仍承担真实共享逻辑的 teacher helper、widget 或 presentation 模块。
- 本轮不改 teacher / platform 页面可见结构、接口契约或路由行为。
- 本轮不处理 `useTeacherAwdReviewIndex` 这类仍承载真实共享逻辑的 owner。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/views/teacher/*.vue`
- `code/frontend/src/features/teacher-class-students/**`
- `code/frontend/src/features/teacher-student-analysis/**`
- `code/frontend/src/features/teacher-student-review-archive/**`
- `code/frontend/src/features/teacher-awd-review/**`
- `code/frontend/src/features/*-workspace/**`

## 当前结论

- teacher route views 已经有对应的中性 workspace owner：
  - `TeacherClassStudents.vue -> useClassStudentsPage`
  - `TeacherStudentAnalysis.vue -> useStudentAnalysisPage`
  - `TeacherStudentReviewArchive.vue -> useStudentReviewArchivePage`
  - `TeacherAWDReviewDetail.vue -> useAwdReviewDetailPage`
- 当前 teacher feature 目录里仍保留一层同名 alias 壳，把 route view 再转发到这些中性 workspace owner。
- 这些 alias 壳不再承担状态、路由或 API owner，只会延长依赖链。

## 任务切片

### Slice 1：teacher route view 直连中性 workspace owner

- 目标：
  - 更新 teacher route views 的 imports，直接指向中性 workspace owner。
  - 删除不再需要的 teacher page alias 壳，并同步收口相关 feature export。
- 预期改动：
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
- 验证：
  - `rg -n "useTeacherClassStudentsPage|useTeacherStudentAnalysisPage|useTeacherStudentAnalysisNavigation|useTeacherStudentReviewArchivePage|useTeacherAwdReviewDetail" code/frontend/src`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
  - `git diff --check -- <touched files>`
- Review focus：
  - teacher route views 是否已经直连中性 workspace owner。
  - 删除 alias 壳后，teacher feature 目录是否只保留仍有真实共享职责的 helper。

## 风险

- 部分源码断言型测试会校验 import 文本，需要同步更新。
- `teacher-*` feature public API 会缩小；若存在未搜索到的隐式引用，会在 typecheck 或 targeted tests 中暴露。

## 回退方式

- 如发现仍有存量引用，可从 Git 历史恢复单个 alias 文件，再按真实调用方拆更小切片。
