# AWD Review Index Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/awd-review-index-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-awd-review-index-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts`
  - `code/frontend/src/features/awd-review-workspace/model/index.ts`
  - `code/frontend/src/features/awd-review-workspace/index.ts`
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- Classification check：同意按单条 feature router owner cleanup 处理；`useAwdReviewIndex.ts` 作为 teacher / platform 共用的 feature model，不应继续直接持有 `vue-router`，而 route view 也不应直接持有导航。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useAwdReviewIndex.ts` 负责 AWD 复盘目录的数据加载、筛选和分页是合理的，但继续把 teacher / platform 两套导航也吞进去，owner 明显过宽。
- route view 直接持有 `useRouter` 同样违反仓库边界，所以这轮改成 `useAwdReviewIndexPage(scope)` route-aware page wrapper 是更合适的落点。
- 本轮收掉的是一条真正的共享 feature router 例外，并把导航压到更明确的 page wrapper，而不是只在 allowlist 里挪位置。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-review-index-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-review-index-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts code/frontend/src/features/awd-review-workspace/model/index.ts code/frontend/src/features/awd-review-workspace/index.ts code/frontend/src/views/platform/AWDReviewIndex.vue code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- AWD review 线后续仍可能继续处理 export flow、detail page 或更深层 route owner，但不属于这轮范围。
