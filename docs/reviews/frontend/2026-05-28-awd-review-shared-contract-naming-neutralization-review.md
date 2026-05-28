# AWD Review Shared Contract Naming Neutralization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-review-shared-contract-naming-neutralization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-review-shared-contract-naming-neutralization.md`
    - `docs/plan/impl-plan/2026-05-28-awd-review-shared-contract-naming-neutralization-plan.md`
    - `docs/reviews/frontend/2026-05-28-awd-review-shared-contract-naming-neutralization-review.md`
    - `code/frontend/src/api/contracts.ts`
    - `code/frontend/src/api/teacher/awd-reviews.ts`
    - `code/frontend/src/api/teaching/awd-reviews.ts`
    - `code/frontend/src/api/awd-reviews.ts`
    - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
    - `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
    - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
    - `code/frontend/src/widgets/awd-review-workspace/model/presentation.ts`
    - `code/frontend/src/widgets/awd-review-workspace/model/presentation.test.ts`
    - `code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue`
    - `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
    - `code/frontend/src/api/__tests__/teacher.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 AWD review shared contract / shared widget owner 的命名收口，不涉及 route 或 endpoint 迁移。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `api/contracts.ts` 里的 AWD review shared response DTO 已统一从 `TeacherAWDReview*` 收口到中性 `AwdReview*`，`api/teacher/awd-reviews.ts`、`api/teaching/awd-reviews.ts` 的 page data、raw response 和 normalize helper 也同步跟上，不再把 shared contract 伪装成 teacher owner。
- `AwdReviewWorkspace.vue`、`AwdReviewIndexWorkspace.vue` 和 `widgets/awd-review-workspace/model/presentation.ts` 里的 summary types、summary builder 与 copy owner 已统一改成中性 `AwdReview*` / `AWD_REVIEW_*` 命名；shared widget 不再在 owner 层残留 teacher 语义。
- teacher endpoint function 名、route name 和导航配置保留不动，仍由 transport / route owner 承担 teacher 语义；本轮没有把 contract 命名收口扩展成权限或路由迁移，边界合理。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/widgets/awd-review-workspace/model/presentation.test.ts src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- AWD review 线仍保留 `TeacherAWDReviewIndex` / `TeacherAWDReviewDetail` 这类 route name，以及 `listTeacherAWDReviews()`、`getTeacherAWDReview()` 这类 teacher endpoint function 名；这是当前明确保留的 route / transport owner，不属于 shared contract 漂移。
- `api/admin/contests.ts` 仍通过 alias 复用 teacher 命名的 AWD review endpoint function；本轮已经把 shared contract 和 shared widget owner 清干净，但 admin API alias 是否继续细分，仍要放回更大的 admin / teacher API 分层语境判断。

## Touched known-debt status

- AWD review 这条更深层已知债在 touched surface 上，已经从“shared response DTO / shared widget presentation naming 仍带 teacher 前缀”收口到只剩 teacher route / endpoint 语义。
- 当前这条 P1 在 contract naming 维度的剩余重点，不再是 shared owner 命名，而是是否还需要继续推动 teacher endpoint function / admin alias 的更深层 API owner 语义拆分。
