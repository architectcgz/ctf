# Reuse Decision

## Change type
frontend refactor / route target cleanup

## Existing code searched
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts
- code/frontend/src/views/platform/AWDReviewIndex.vue
- code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts
- code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useChallengePackageFormatPage.ts`、`useNotificationListPage.ts`、`useStudentReviewArchivePage.ts` 这类薄 page wrapper 已继续收口成纯 route target contract，再由 view / widget 直接消费 `AppRouteLink`。
- `teacherDashboardRoutes.ts`、`skillProfileRoutes.ts`、`studentReviewArchiveRoutes.ts` 这类 feature 内 route target 文件，已经是当前仓库处理 `feature -> vue-router` 薄导航债的主模式。

## Decision
refactor_existing

## Reason
`useAwdReviewIndex.ts` 现在已经是纯数据 / 筛选 / 分页 owner，但 `useAwdReviewIndexPage.ts` 仍直接持有 `useRouter()`，只负责两类薄导航：

- 返回教师总览 / 平台概览
- 进入 AWD 复盘详情

这类只拼 route target 的 page wrapper 继续停留在 `feature -> vue-router` allowlist 没有收益。当前仓库同类债已经改成“feature 提供 route target contract，view/widget 直接用 `AppRouteLink`”，因此 AWD review index 也应对齐。

最小正确改动是：

- 保留 `useAwdReviewIndex.ts` 的数据加载、筛选、分页和目录展示 owner
- 把 `useAwdReviewIndexPage(scope)` 继续收口成纯 route target wrapper
- 新增 `awdReviewIndexRoutes.ts` 承接角色感知 route target
- `TeacherAWDReviewIndex.vue`、`AWDReviewIndex.vue`、teacher workspace、platform hero / directory panel 直接消费 `AppRouteLink`
- 删除 `features/awd-review-workspace/model/useAwdReviewIndexPage.ts -> vue-router` allowlist

本轮不做：

- 不调整 AWD 复盘目录的视觉结构
- 不继续处理 `useAwdReviewExportFlow.ts` 或 AWD detail 页的 router 依赖
- 不改变 teacher / platform 现有导航目标

## Files to modify
- .harness/reuse-decisions/awd-review-index-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-awd-review-index-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts
- code/frontend/src/features/awd-review-workspace/model/awdReviewIndexRoutes.ts
- code/frontend/src/features/awd-review-workspace/model/index.ts
- code/frontend/src/features/awd-review-workspace/index.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue
- code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue
- code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue
- code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue
- code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue
- code/frontend/src/views/platform/AWDReviewIndex.vue
- code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts
- code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts

## After implementation
- `useAwdReviewIndexPage.ts` 不再 import `vue-router`
- teacher / platform AWD review index 的返回与详情入口改为显式 route target contract
- `featureRouterImportAllowlist` 收掉 `useAwdReviewIndexPage.ts`
- AWD 复盘目录现有数据加载和跳转行为保持不变
