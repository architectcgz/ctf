# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts
- code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue
- code/frontend/src/views/platform/PlatformStudentReviewArchive.vue
- code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
- code/frontend/src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts

## Similar implementations found
- `teacher-dashboard-route-target-cleanup` 已示范“page model 暴露 route target，UI 通过 `AppRouteLink` 消费”的模式。
- `skill-profile-route-target-cleanup` 已示范“workspace shell 内 CTA / 列表跳转改为 route target，不动数据加载 owner”的做法。

## Decision
refactor_existing

## Reason
`useStudentReviewArchivePage.ts` 当前的 router 依赖只剩两条声明式导航：

- `openStudentAnalysis()`
- `goBack()`

真正需要继续留在 page model 里的 owner 是：

- 复盘归档数据加载
- 导出、轮询、下载与错误提示

最小正确改动是：

- 给 `student-review-archive-workspace` 补本地 route target helper
- `useStudentReviewArchivePage.ts` 去掉 `vue-router`，改为暴露 `analysisRoute` 和 `backRoute`
- `ReviewArchiveHero.vue` 通过共享 `AppRouteLink` 消费这两个 route target
- route view 继续只组合 feature page model 与 workspace widget

这样可以收掉：

- `features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts -> vue-router`

本轮不做：

- 不处理导出轮询、下载和 toast owner
- 不迁移 `ReviewArchiveHero.vue` / `ReviewArchiveWorkspace.vue` 的目录归属
- 不改学生复盘归档数据、证据展示或 teacher observation 展示

## Files to modify
- .harness/reuse-decisions/student-review-archive-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-student-review-archive-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-student-review-archive-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/student-review-archive-workspace/model/index.ts
- code/frontend/src/features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts
- code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts
- code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue
- code/frontend/src/views/platform/PlatformStudentReviewArchive.vue
- code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue
- code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue
- code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts

## After implementation
- `useStudentReviewArchivePage.ts` 不再 import `vue-router`
- 复盘归档页的“返回学生列表 / 返回学员分析”改成 route target + `AppRouteLink`
- `featureRouterImportAllowlist` 再减少 1 条
