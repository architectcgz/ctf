# AWD Review Index Route Target Cleanup 复核

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
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts`
  - `code/frontend/src/features/awd-review-workspace/model/awdReviewIndexRoutes.ts`
  - `code/frontend/src/features/awd-review-workspace/model/index.ts`
  - `code/frontend/src/features/awd-review-workspace/index.ts`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue`
  - `code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue`
  - `code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue`
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- Classification check：同意按单条 feature route target cleanup 处理；`useAwdReviewIndexPage.ts` 继续保留 `useRouter()` 只是在拖住 allowlist，没有承担额外 workflow owner，适合继续收口成 route target contract。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useAwdReviewIndex.ts` 保持数据加载、筛选和分页 owner 是合理的，这轮不需要再动。
- `useAwdReviewIndexPage.ts` 只剩两个薄导航目标，继续停留在 `useRouter()` 已经没有结构收益；改成 route target contract 后，view / widget 的导航边界会更清楚。
- 本轮收掉的是一条真正的共享 feature router 例外，并把导航进一步压成显式 route target，而不是只在 allowlist 里挪位置。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-review-index-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-review-index-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts code/frontend/src/features/awd-review-workspace/model/awdReviewIndexRoutes.ts code/frontend/src/features/awd-review-workspace/model/index.ts code/frontend/src/features/awd-review-workspace/index.ts code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue code/frontend/src/views/platform/AWDReviewIndex.vue code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- AWD review 线后续仍可能继续处理 export flow、detail page 或更深层 route owner，但不属于这轮范围。
