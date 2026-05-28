# Reuse Decision

## Change type
frontend api owner / shared transport wrapper alignment

## Existing code searched
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/awd-reviews.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/plan/impl-plan/2026-05-27-admin-awd-review-api-owner-alignment-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-28-awd-review-role-aware-access-owner-normalization-plan.md`
- `docs/plan/impl-plan/2026-05-28-awd-review-shared-contract-naming-neutralization-plan.md`
- `docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`
- `docs/reviews/frontend/2026-05-28-awd-review-role-aware-access-owner-normalization-review.md`
- `docs/reviews/frontend/2026-05-28-awd-review-shared-contract-naming-neutralization-review.md`

## Similar implementations found
- `api/teaching/*` 已经承担共享教学域实现 owner，`api/teacher/*` 只负责对 teacher public owner 暴露语义化函数。
- AWD review 已经有 `api/awd-reviews.ts` 作为 role-aware facade，说明本地 API owner 已经从 route / feature 面收回到单点 owner。
- 当前 `api/admin/contests.ts` 仍直接 alias `listTeacherAWDReviews as listPlatformAWDReviews`，说明剩余残片已经下探到共享实现层函数命名，而不是页面或 DTO owner。

## Decision
refactor_existing

## Reason
- 本轮不新增后端 `/admin/awd/reviews*` 路径，也不修改 `/api/v1/teacher/awd/reviews*` 的 HTTP contract。
- 最小正确改动是把 `api/teaching/awd-reviews.ts` 的共享实现符号收口成中性 `AwdReview*`，再让 `api/admin` 在 public owner 层显式暴露 platform 命名；`api/teacher` 继续保留 teacher public owner 不动。
- 这样可以停止 admin public API 继续显式依赖 teacher 命名函数，同时保留既有前端调用面和后端权限语义。

## Files to modify
- `.harness/reuse-decisions/awd-review-api-implementation-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-28-awd-review-api-implementation-owner-neutralization-plan.md`
- `docs/reviews/frontend/2026-05-28-awd-review-api-implementation-owner-neutralization-review.md`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `api/teaching/awd-reviews.ts` 不再把 teacher 语义写死在共享实现符号里。
- `api/teacher` 继续保留 `listTeacherAWDReviews()` 等 teacher public API，不需要同步改动。
- `api/admin` 继续保留 `listPlatformAWDReviews()` 等 platform public API，但不再通过 alias teacher-named function 暴露。
- AWD review 这条 admin / teacher 结构耦合在前端本地 API owner 维度会进一步缩到后端既有 teacher HTTP path 与 teacher route name。
