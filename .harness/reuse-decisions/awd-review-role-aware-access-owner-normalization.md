# Reuse Decision

## Change type
frontend architecture / role-aware access owner normalization

## Existing code searched
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts
- code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts
- code/frontend/src/views/platform/PlatformAwdReviewDetail.vue
- code/frontend/src/views/platform/AWDReviewIndex.vue
- code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts
- code/frontend/src/api/admin/contests.ts
- code/frontend/src/api/teacher/awd-reviews.ts
- code/frontend/src/api/teaching/awd-reviews.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- docs/reviews/architecture/2026-05-24-frontend-architecture-review.md

## Similar implementations found
- `api/admin/contests.ts` 当前已经通过 alias 复用 `api/teaching/awd-reviews.ts`，说明 AWD review 的底层 contract 实际上已经是共享实现，只是 feature 层还在自己分支挑 `admin` / `teacher` 入口。
- `useClassWorkspaceSection()` 这类中立 feature 已经把 role route 差异收口成单点 owner，而不是让每个 route view 各自维护 redirect 细节。
- `resolveAwdReviewDetailRouteName()` / `resolveAwdReviewIndexRouteName()` 已经把 AWD review route 差异收在中立 util，说明这条线已经接受“共享 feature + role-aware owner”模式。

## Decision
refactor_existing

## Reason
`useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 当前都在 feature 内直接双引 `@/api/admin` 和 `@/api/teacher`，再按 `authStore.user?.role` 做分支。这样共享 feature 的 role-aware access owner 被拆散在多个文件里，admin / teacher 结构耦合仍然停留在 feature 调用面。最小正确改动是：

- 新增一层中立 AWD review access owner，统一承接 list / detail / archive export / report export 的 role-aware API 选择
- `useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 改为只依赖这层中立 access owner
- 同步更新 teacher / platform AWD review 路由测试与 backlog 记录

本轮不继续改 AWD review 的 DTO 命名，也不拆 `widgets/awd-review-workspace` 里的 `Teacher*` 命名展示类型。

## Files to modify
- .harness/reuse-decisions/awd-review-role-aware-access-owner-normalization.md
- docs/plan/impl-plan/2026-05-28-awd-review-role-aware-access-owner-normalization-plan.md
- docs/reviews/frontend/2026-05-28-awd-review-role-aware-access-owner-normalization-review.md
- code/frontend/src/api/awd-reviews.ts
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts
- code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts
- code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts
- code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- AWD review 共享 feature 会把 role-aware API 选择收口成单点 owner。
- `useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 不再各自双引 `@/api/admin` / `@/api/teacher`。
- admin / teacher AWD review route 继续复用同一个 feature，但 role access owner 会更清楚。
