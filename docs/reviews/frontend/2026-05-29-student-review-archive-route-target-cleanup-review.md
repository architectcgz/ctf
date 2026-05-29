# Student Review Archive Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-student-review-archive-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/student-review-archive-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-student-review-archive-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-student-review-archive-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components/navigation/AppRouteLink.vue`
  - `code/frontend/src/features/student-review-archive-workspace/model/index.ts`
  - `code/frontend/src/features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts`
  - `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
  - `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
  - `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
  - `code/frontend/src/router/routes/teacherRoutes.ts`
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue`
  - `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts`
- Classification check：同意按 `student-review-archive-workspace` feature 内“薄导航 route target cleanup”处理；原 page model 的 router 依赖只剩 hero 上的“返回学生列表 / 返回学员分析”，不应继续保留为 reviewed route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `studentReviewArchiveRoutes.ts` 现在集中提供角色感知的“返回学生列表 / 返回学员分析” route target contract。
- `useStudentReviewArchivePage.ts` 已退回纯复盘归档数据、导出轮询、下载和错误提示 owner，不再直接 import `vue-router`；`className / studentId` 改由 route props 传入。
- `TeacherStudentReviewArchive.vue` 与 `PlatformStudentReviewArchive.vue` 现在都通过 router props 把 `className / studentId` 传给 page model，route view 继续保持薄壳，不直接接 router API。
- `ReviewArchiveHero.vue` 已通过共享 `AppRouteLink.vue` 渲染 hero 上的两条返回导航；“导出复盘归档”仍由 button + emit 驱动，没有把导出 workflow 一起迁走。
- `TeacherStudentReviewArchive.test.ts` 已改成真实 router 导航断言，同时补上“page model 不再 import vue-router、hero 直接消费 AppRouteLink、teacher/admin 角色命中正确路由”的 raw-source 护栏。
- `featureRouterImportAllowlist` 已再减少 1 条：`features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts -> vue-router`。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-review-archive-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-student-review-archive-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-review-archive-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-review-archive-workspace/model/index.ts code/frontend/src/features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue code/frontend/src/views/platform/PlatformStudentReviewArchive.vue code/frontend/src/router/routes/teacherRoutes.ts code/frontend/src/router/routes/platformRoutes.ts code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- 本轮不处理导出轮询、下载和 toast owner；后续如果复盘归档页再增长，更合适继续沿导出 workflow 切片，而不是把 hero 导航 owner 再拉回 page model。
