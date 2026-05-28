# AWD Review Role-Aware Access Owner Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-review-role-aware-access-owner-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-review-role-aware-access-owner-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-awd-review-role-aware-access-owner-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-awd-review-role-aware-access-owner-normalization-review.md`
    - `code/frontend/src/api/awd-reviews.ts`
    - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
    - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
    - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
    - `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 AWD review 共享 feature 的 role-aware access owner 收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- 已新增 `code/frontend/src/api/awd-reviews.ts` 作为 AWD review 共享 feature 的中立 role-aware access owner，`list / detail / archive export / report export` 的角色分派不再分散在 3 个 feature model 里各自判断。
- `useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 已统一改为只依赖 `@/api/awd-reviews`，共享 feature model 不再直接双引 `@/api/admin` 和 `@/api/teacher`。
- 这次实现把 owner 收口保持在 API facade 层，页面路由、导出轮询、错误提示和筛选分页状态仍留在各自 feature model，不会把 page/workflow owner 反向抽空。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这轮只收口 AWD review 共享 feature 的 access owner，不继续重命名 `TeacherAWDReview*` 深层 DTO / widget 合同；如果后续继续推进 admin / teacher 结构耦合，这批命名残片仍需要单独再切一刀。
- `api/admin/contests.ts` 当前仍以 alias 方式转发 AWD review API；本轮已经把 feature model 侧的 owner 散落收住，但 admin API barrel 本身是否继续细分，还要结合 contest 线整体 API 分层再判断。

## Touched known-debt status

- AWD review 共享 feature 在 touched surface 内，已从“feature model 各自散落 role-aware API 分支”收口到单点 `api/awd-reviews.ts` owner。
- 这条 P1 当前已从 access owner 漂移收敛到更深层的 DTO / contract naming 残片，不再继续停留在 shared feature model 的双 API import。
