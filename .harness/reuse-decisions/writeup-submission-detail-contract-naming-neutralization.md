# Reuse Decision

## Change type
+api / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/plan/impl-plan/2026-05-27-writeup-submission-contract-naming-neutralization-implementation-plan.md`
- `docs/reviews/frontend/2026-05-27-writeup-submission-contract-naming-neutralization-review.md`
- `docs/reviews/frontend/2026-05-27-manual-review-submission-contract-naming-neutralization-review.md`

## Similar implementations found
- 上一刀刚把 manual review 共享 DTO 收口到中性命名，说明当前可以继续沿“单个共享 contract 一刀一收”的方式推进。
- `TeacherSubmissionWriteupDetailData` 当前几乎只停留在 `api/contracts.ts` 和 backlog / plan / review 记录中，没有继续扩散到 active frontend workflow。

## Decision
refactor_existing

## Reason
- 这组 detail DTO 仍带 teacher 前缀，但实际并没有落到教师专属消费面；继续保留只会留下一块无意义的历史命名残片。
- 因为引用面很薄，本轮可以用更小切片完成：只改 contract 名称和事实文档，不混入 manual review、class 或 attack session 其他命名债。

## Files to modify
- `.harness/reuse-decisions/writeup-submission-detail-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-writeup-submission-detail-contract-naming-neutralization-implementation-plan.md`
- `code/frontend/src/api/contracts.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-writeup-submission-detail-contract-naming-neutralization-review.md`

## After implementation
- 这条 detail DTO 收口后，题解域剩余的 teacher 前缀 contract 残余将进一步收敛到 `TeacherClassItem`、`TeacherAWDReviewContestItemData` 和 `TeacherAttackSessionQuery`。
