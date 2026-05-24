> 状态：Current
> 事实源：2026-05-24 前端架构 review、当前代码结构、已落地平台路由拆分
> 替代：无

# Platform Route Owner Decoupling Implementation Plan

## 目标

- 按最小可审查切片，把 `/platform/*` 对 teacher route view 的直接依赖逐步收口。
- 当前活动切片只处理 `PlatformStudentAnalysis`，让管理员学生分析页拥有自己的 route view，并把共享 page workflow 上提到中性 feature owner。

## 非目标

- 本轮不处理 `PlatformStudentReviewArchive`。
- 本轮不处理 `PlatformClassTrend`、`PlatformClassReview`、`PlatformClassInsights`、`PlatformClassIntervention`。
- 本轮不改学生分析页的用户可见流程、接口契约或视觉结构。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`

## 当前结论

- `PlatformClassStudents` 已经完成第一刀 owner 解耦，平台路由不再直接挂 teacher route view。
- `PlatformStudentAnalysis` 仍直接指向 `@/views/teacher/TeacherStudentAnalysis.vue`，是下一条最小连续链路。
- 共享 workflow 的真正 owner 在 `features/teacher-student-analysis/model/*`；如果只新增一个平台 view 但继续直接引用 teacher feature，结构债仍然原样保留。

## 任务切片

### Slice 1：平台学生分析页 owner 解耦

- 目标：
  - 新增平台 route view。
  - 把 page-level workflow owner 上提到中性 feature，例如 `student-analysis-workspace`。
  - teacher 侧保留兼容桥，避免一次性搬空全部 teacher 相关 helper。
- 预期改动：
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
  - `code/frontend/src/features/student-analysis-workspace/**`
  - `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
  - `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
- 依赖：
  - 复用 `teacher-student-analysis` 现有 helper，不复制异步工作流。
  - 继续使用 `teachingWorkspaceRouting` 做 role-aware 跳转。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/router/__tests__/sharedRoutes.test.ts`
  - `npm run typecheck`
- Review focus：
  - 平台 route owner 是否真正脱离 teacher view。
  - 新中性 feature 是否成为实际 page workflow owner，而不是只做再包装。
  - admin 角色下跳转 `PlatformStudentReviewArchive`、`PlatformClassStudents` 是否保持正确。

### Slice 2：平台学生复盘归档 owner 解耦

- 目标：
  - 让 `PlatformStudentReviewArchive` 不再直接挂 teacher route view。
- 预期改动：
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts`
  - `code/frontend/src/features/student-review-archive-workspace/**`
  - `code/frontend/src/features/teacher-student-review-archive/model/useTeacherStudentReviewArchivePage.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
- 依赖：
  - 继续复用 `teacher-student-review-archive` 中已稳定的数据查询、导出消息和 workspace widget，不复制导出轮询链路。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/router/__tests__/sharedRoutes.test.ts`
  - `npm run typecheck`
- Review focus：
  - 平台复盘归档页是否摆脱 teacher route view。
  - 导出归档、返回分析页、返回班级页在 admin 角色下是否仍落到平台路由。

## 风险

- `useTeacherStudentAnalysisPage` 体量较大，直接整体搬迁会放大 diff；因此本轮只移动 page owner 和导航 owner，保留内部 helper 现状。
- 复盘归档页虽然更小，但仍涉及导出轮询、文件下载和错误提示；因此本轮只把 page owner 上提到中性 feature，不改稳定的导出实现细节。
- 学生分析页涉及 review workspace、writeup、manual review、report export，多处 watch / async flow 同时存在，不能因为 route owner 迁移破坏现有角色分流。

## 回退方式

- 如果平台页新 owner 引入回归，可回退本轮新增平台 route view 和中性 feature，并恢复 `platformRoutes.ts` 对 teacher view 的直接引用。
- teacher 侧兼容桥会保留到下一轮继续迁移前，不要求一次性删除旧入口。
