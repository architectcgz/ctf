# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts
- code/frontend/src/views/platform/AWDReviewIndex.vue
- code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts
- code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useNotificationDrawer.ts`、`useChallengeManagePresentation.ts`、`useAuth.ts` 这类共享 feature helper 已逐步改成纯 workflow/data owner，不再自己持有 `vue-router`。
- `useClassStudentsPage.ts`、`usePlatformUserManagePage.ts` 这类 route view owner 已承担 query / redirect / navigation，feature helper 退回纯业务状态 owner。

## Decision
refactor_existing

## Reason
`useAwdReviewIndex.ts` 目前被 teacher / platform 两个 route view 共同消费，但它内部直接做了三类导航：

- 打开 AWD 复盘详情
- 返回教师总览
- 返回平台概览

这让共享 feature model 直接认识不同角色的 route name，owner 过宽；但本仓库的 route view 也不允许直接持有 `useRouter`，所以导航不能简单回退到 `.vue` view。

最小正确改动是：

- 保留 `useAwdReviewIndex.ts` 的数据加载、筛选、分页和目录展示 owner
- 把 `openContest`、`openDashboard`、`openPlatformOverview` 从共享 feature model 移出
- 新增显式 route-aware page wrapper `useAwdReviewIndexPage(scope)` 承接 router.push
- `TeacherAWDReviewIndex.vue` 和 `AWDReviewIndex.vue` 继续保持薄 route shell
- 删除 `features/awd-review-workspace/model/useAwdReviewIndex.ts -> vue-router` allowlist

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
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts
- code/frontend/src/features/awd-review-workspace/model/index.ts
- code/frontend/src/features/awd-review-workspace/index.ts
- code/frontend/src/views/platform/AWDReviewIndex.vue
- code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts
- code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts

## After implementation
- `useAwdReviewIndex.ts` 不再 import `vue-router`
- teacher / platform AWD review index route shell 经由 `useAwdReviewIndexPage(scope)` 持有本角色导航 owner
- `featureRouterImportAllowlist` 收掉 `useAwdReviewIndex.ts`，仅保留新的 route-aware page wrapper
- AWD 复盘目录现有数据加载和跳转行为保持不变
