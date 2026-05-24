> 状态：Current
> 事实源：2026-05-24 前端架构 review、当前代码结构、已落地平台路由拆分
> 替代：无

# Platform Route Owner Decoupling Implementation Plan

## 目标

- 按最小可审查切片，把 `/platform/*` 对 teacher route view 的直接依赖逐步收口。
- 当前活动切片处理 `PlatformAwdReviewDetail`，让管理员 AWD 复盘详情页拥有自己的 route view，并把 page workflow owner 上提到中性 feature。

## 非目标

- 本轮不改 teacher 侧 AWD 复盘详情的用户可见结构。
- 本轮不处理更大范围的 AWD review widget 中性化，只收口 route/page owner。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`

## 当前结论

- `PlatformClassStudents`、`PlatformStudentAnalysis`、`PlatformStudentReviewArchive`、平台班级工作台别名页都已经完成 owner 解耦。
- 当前剩余最小连续链路只剩 `PlatformAwdReviewDetail`。
- 该路由当前仍直接指向 `@/views/teacher/TeacherAWDReviewDetail.vue`，并复用 teacher page hook；其中 `openReviewIndex()` 仍硬编码返回 `TeacherAWDReviewIndex`，说明 admin 详情页 workflow owner 还没有真正中性化。

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

### Slice 3：平台班级工作台别名页 owner 解耦

- 目标：
  - 让 `PlatformClassTrend`、`PlatformClassReview`、`PlatformClassInsights`、`PlatformClassIntervention` 不再直接挂 teacher route view。
  - 把 alias route -> canonical page 的重定向 owner 上提到中性 feature。
- 预期改动：
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
  - `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
  - `code/frontend/src/features/class-workspace-redirect/**`
  - `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
  - `code/frontend/src/features/teacher-class-workspace/**`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
- 依赖：
  - 继续复用 `PlatformClassStudents` / `TeacherClassStudents` 作为 canonical workspace page。
  - panel 映射仍沿用现有 `trend / review / insight / action` 约定。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/router/__tests__/sharedRoutes.test.ts`
  - `npm run typecheck`
- Review focus：
  - 平台别名路由是否真正脱离 teacher route view。
  - admin 角色下 canonical redirect 是否回到 `PlatformClassStudents` 而不是 `TeacherClassStudents`。
  - teacher 侧旧别名页是否直接依赖中性 redirect owner，而不再保留 teacher-feature 兼容桥。

### Slice 4：平台 AWD 复盘详情页 owner 解耦

- 目标：
  - 让 `PlatformAwdReviewDetail` 不再直接挂 teacher route view。
  - 把 AWD 复盘详情页 workflow owner 上提到中性 feature，修正 admin 返回索引仍落到 teacher 命名空间的问题。
- 预期改动：
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
  - `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
  - `code/frontend/src/features/awd-review-detail-workspace/**`
  - `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewDetail.ts`
  - `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
- 依赖：
  - 继续复用 `TeacherAWDReviewWorkspace` widget 与 `useAwdReviewExportFlow`。
  - AWD 详情页返回索引与轮次切换继续沿用 `teachingWorkspaceRouting` 做 role-aware 路由解析。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/router/__tests__/sharedRoutes.test.ts`
  - `npm run typecheck`
- Review focus：
  - 平台 AWD 详情路由是否真正脱离 teacher route view。
  - admin 角色下返回列表、轮次切换是否继续留在 platform 命名空间。
  - teacher 侧旧详情页是否继续保持兼容行为。

## 风险

- `useTeacherStudentAnalysisPage` 体量较大，直接整体搬迁会放大 diff；因此本轮只移动 page owner 和导航 owner，保留内部 helper 现状。
- 复盘归档页虽然更小，但仍涉及导出轮询、文件下载和错误提示；因此本轮只把 page owner 上提到中性 feature，不改稳定的导出实现细节。
- 班级工作台别名页虽然逻辑更轻，但它同时承接 teacher / platform 两套路由名映射，canonical redirect 的目标路由如果判错，会直接把管理员链路跳回 teacher namespace。
- AWD 复盘详情页已经有一部分 role-aware 路由逻辑，但返回索引仍写死在 teacher 命名空间；如果只新增平台 view 而不迁移 page owner，管理员会继续被带回 `/academy/awd-reviews`。
- 学生分析页涉及 review workspace、writeup、manual review、report export，多处 watch / async flow 同时存在，不能因为 route owner 迁移破坏现有角色分流。

## 回退方式

- 如果平台页新 owner 引入回归，可回退本轮新增平台 route view 和中性 feature，并恢复 `platformRoutes.ts` 对 teacher view 的直接引用。
- teacher 侧兼容桥会保留到下一轮继续迁移前，不要求一次性删除旧入口。
