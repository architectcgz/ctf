# Reuse Decision

## Change type
+api / feature / widget / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexFilters.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 题解投稿目录、manual review、writeup detail、班级目录和 attack session query 已经按“共享 contract 一组一刀”的方式完成中性化命名，说明这类薄 DTO / query 残片适合继续按最小切片推进。
- AWD review 目录当前已经同时服务 teacher / platform 两个前端 owner，teacher / admin wrapper 只是角色选择层，不再代表目录项 DTO 仍是教师专属语义。

## Decision
refactor_existing

## Reason
- `TeacherAWDReviewContestItemData` 当前停留在共享 AWD review index contract 层，被 feature、widget 和 teacher / platform 两边共同消费；保留 teacher 前缀会继续误导目录 DTO owner。
- 这次只改赛事目录项 DTO 名称和相关消费面，不顺手动 archive / round / team / service 等更深的 teacher response contract，能控制 blast radius。

## Files to modify
- `.harness/reuse-decisions/awd-review-contest-item-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-awd-review-contest-item-contract-naming-neutralization-implementation-plan.md`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexFilters.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-awd-review-contest-item-contract-naming-neutralization-review.md`

## After implementation
- 这组 DTO 收口后，当前 backlog 里已显式记录的更深层 teacher 前缀共享 contract 命名残片将完成本阶段清理。
