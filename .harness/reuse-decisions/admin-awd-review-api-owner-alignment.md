# Reuse Decision

## Change type
api / feature / page / test / docs

## Existing code searched
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/admin/index.ts`
- `code/frontend/src/api/teacher/index.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`

## Similar implementations found
- `api/admin/teaching.ts` 已经证明这个仓库接受“底层 HTTP 实现继续留在 `api/teaching/*`，上层通过 `api/<role>/*` 薄 wrapper 收口角色语义 owner”的模式。
- `student-analysis-workspace`、`student-review-archive-workspace` 最近已经通过 role-aware owner 收口过共享 workflow，说明这轮不需要复制平台专属 AWD review hook。
- `useAwdReviewIndex`、`useAwdReviewDetailPage`、`useAwdReviewExportFlow` 已经是共享 feature owner，最小改动是让它们按角色挑选 API owner，而不是再造一条 admin 平行实现。

## Decision
extend_existing

## Reason
- 当前 AWD 复盘的 route view 和 workspace owner 已经中立化，但共享 feature 仍直接 import `@/api/teaching` 下的 `Teacher*` 函数，导致 `/platform/*` 仍通过共享 feature 间接依赖 teacher 语义 API owner。
- 最小正确方案不是再拆页面，而是复用现有 `api/admin/contests.ts` 作为 admin AWD review wrapper 落点，并让共享 AWD review feature 按当前角色选择 `api/admin` 或 `api/teacher` owner。
- 这样可以把收口范围限制在 AWD review query / export owner，不新增平行 workspace、不改路由、不改 DTO contract，也不引入新的 API wrapper 目录文件。

## Files to modify
- `.harness/reuse-decisions/admin-awd-review-api-owner-alignment.md`
- `docs/plan/impl-plan/2026-05-27-admin-awd-review-api-owner-alignment-implementation-plan.md`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`

## After implementation
- 如果这层 AWD review owner 收口稳定，后续同类 admin / teacher 共享能力可以继续沿用“共享 feature 通过角色选择 `api/admin` / `api/teacher` wrapper”的模式，而不是直接让平台页或共享 feature 绑定 `api/teaching` 的 teacher 命名函数。
