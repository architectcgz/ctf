> 状态：Current
> 事实源：2026-05-24 前端架构 review、`featureBoundaries.test.ts` 当前失败项
> 替代：无

# Feature Public API Deep Import Cleanup Implementation Plan

## 目标

- 消除 `code/frontend/src/features/__tests__/featureBoundaries.test.ts` 中现存的 cross-feature `model/*` 深导入。
- 在不改变既有业务行为的前提下，把 helper / query / export flow 的跨 feature 依赖收口到 donor feature 的公共出口。

## 非目标

- 本轮不把 teacher 前缀 helper 继续改名为中性 feature。
- 本轮不重写相关 workflow 的实现逻辑，只调整公共导出边界和引用路径。
- 本轮不处理 `request.ts` 的全局错误跳转策略。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- 当前 cross-feature deep import 使用点：
  - `features/class-students-workspace/model/useClassStudentsPage.ts`
  - `features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
  - `features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
  - `features/teacher-class-report-export/model/useTeacherClassReportExport.ts`

## 当前结论

- 当前失败并不是“实现缺失”，而是几个 donor feature 还没有把已复用的 helper 暴露为公共 API。
- 这类债适合低风险收口：
  - donor feature 增加 index / model index 导出
  - 使用方改走 `@/features/<feature>` 或 `@/features/<feature>/model`
- 这样能先把边界测试变成真实约束，再决定后续是否把 teacher 命名进一步中性化。

## 任务切片

### Slice 1：public API 导出与调用方收口

- 目标：
  - 补齐 `teacher-class-insight-window`、`teacher-student-analysis`、`teacher-awd-review` 的公共导出。
  - 将 5 个调用方的 cross-feature deep import 改成公共出口。
- 预期改动：
  - `docs/plan/impl-plan/2026-05-24-feature-public-api-deep-import-cleanup-implementation-plan.md`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
  - `code/frontend/src/features/teacher-class-insight-window/index.ts`
  - `code/frontend/src/features/teacher-class-insight-window/model/index.ts`
  - `code/frontend/src/features/teacher-student-analysis/index.ts`
  - `code/frontend/src/features/teacher-awd-review/index.ts`
  - `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
  - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
  - `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
- 依赖：
  - 继续复用 donor feature 现有实现，不复制 helper 逻辑。
  - `teacher-student-review-archive` 已有 `index.ts` 暴露 `presentation` 和 query helper，直接复用现有公共出口。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/features/__tests__/featureBoundaries.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 是否仍存在 `@/features/*/model/*` 跨 feature 深导入。
  - donor feature 的公共出口是否只暴露已被复用的稳定 helper。
  - 调用方是否只发生 import 路径调整，没有引入行为变化。

## 风险

- `teacher-class-insight-window` 之前没有公共出口，本轮新加 `index.ts` 和 `model/index.ts` 时需要保证导出名与现有调用一致。
- `teacher-student-analysis` 和 `teacher-awd-review` 的 root index 之前只暴露 page hook；本轮扩大导出面后，必须避免误改已有默认使用者。

## 回退方式

- 如本轮导出调整引入回归，可回退新增的 donor feature 公共导出和调用方 import 变更，恢复为当前深导入路径。
- 相关 helper 的实现文件本轮不移动，因此回退成本仅限 import 和 export 层。
