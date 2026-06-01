# AWD Review API Implementation Owner Neutralization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-review-api-implementation-owner-neutralization-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/awd-review-api-implementation-owner-neutralization.md`
  - `docs/plan/impl-plan/2026-05-28-awd-review-api-implementation-owner-neutralization-plan.md`
  - `docs/reviews/frontend/2026-05-28-awd-review-api-implementation-owner-neutralization-review.md`
  - `code/frontend/src/api/teaching/awd-reviews.ts`
  - `code/frontend/src/api/admin/contests.ts`
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/api/__tests__/admin.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 AWD review 本地 API owner 更深层收口，不涉及后端路径迁移。
- Gate verdict：Implemented and re-validated

## Review focus

- `api/teaching/awd-reviews.ts` 是否从 teacher 命名共享实现切到中性实现 owner
- `api/admin/contests.ts` 是否停止 alias teacher 命名函数
- teacher / admin public API 与 feature 调用面是否保持稳定

## Findings

- 无新的未收口 finding。

## Material findings

- 无。

## Senior implementation assessment

- `api/teaching/awd-reviews.ts` 已把 AWD review 共享实现层函数名收口到中性 `listAwdReviews()`、`getAwdReview()`、`exportAwdReviewArchive()`、`exportAwdReviewReport()`；共享实现层不再把 teacher 语义写死在 owner 上。
- `api/admin/contests.ts` 已改为显式 platform wrapper，而不是继续 `listTeacherAWDReviews as listPlatformAWDReviews` 这类 alias teacher 命名函数；admin public owner 与 teaching 实现 owner 的边界更清楚。
- 这刀没有触碰后端 `/api/v1/teacher/awd/reviews*` path，也没有改 AWD review 页面和 feature 行为，边界合理。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/api/__tests__/admin.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- AWD review 仍保留 `/api/v1/teacher/awd/reviews*` 这组后端路径，以及 `TeacherAWDReviewIndex` / `TeacherAWDReviewDetail` route name；这是当前刻意保留的 transport / route 语义，不属于前端本地共享 owner 漂移。
- `api/teacher/awd-reviews.ts` 仍保留 teacher public owner 实现；本轮没有把 teacher / teaching 这两份文件进一步做物理去重。

## Touched known-debt status

- 本轮 touched 的已知结构债是 AWD review 前端本地 API owner 还残留 teacher 命名实现与 admin alias teacher 函数。
- 在本轮 touched surface 上，这条债务已经完成当前阶段收口；当前 AWD review 线上前端本地残留的 teacher 语义，已进一步缩到后端 teacher HTTP path 和 teacher route name。
