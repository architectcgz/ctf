# Reuse Decision

## Change type
frontend contract naming / shared awd review owner normalization

## Existing code searched
- code/frontend/src/api/contracts.ts
- code/frontend/src/api/teacher/awd-reviews.ts
- code/frontend/src/api/teaching/awd-reviews.ts
- code/frontend/src/api/awd-reviews.ts
- code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue
- code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue
- code/frontend/src/widgets/awd-review-workspace/model/presentation.ts
- code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue
- code/frontend/src/widgets/awd-review-workspace/model/presentation.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts
- code/frontend/src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts
- code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts
- code/frontend/src/api/__tests__/teacher.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- docs/plan/impl-plan/2026-05-27-awd-review-contest-item-contract-naming-neutralization-implementation-plan.md
- docs/reviews/frontend/2026-05-27-awd-review-contest-item-contract-naming-neutralization-review.md
- docs/plan/impl-plan/2026-05-28-awd-review-role-aware-access-owner-normalization-plan.md

## Similar implementations found
- `AwdReviewContestItemData` 已经是上一轮收口后的中性 AWD review contest DTO，说明这条命名线已经接受“共享 owner 用中性 contract，teacher 只留在 endpoint / route 语义”的模式。
- `ClassDirectoryItem`、`ManualReviewSubmissionItemData`、`WriteupSubmissionDetailData` 等前序切片都沿用了“共享 contract 去 teacher 化，但不顺手改 teacher route / endpoint 名称”的策略。
- `AwdReviewWorkspace`、`AwdReviewIndexWorkspace` 当前已经是 teacher / platform 共用 widget，继续保留 `TeacherAWDReviewArchiveData`、`TeacherAwdReviewSummaryStats`、`TEACHER_AWD_REVIEW_WORKSPACE_COPY` 这类命名，不符合实际 owner。

## Decision
refactor_existing

## Reason
当前 AWD review 更深层残片不再是 route / feature owner，而是共享 contract 与 shared widget presentation naming：

- `api/contracts.ts` 里 `TeacherAWDReviewArchiveData`、`TeacherAWDReviewRoundItemData`、`TeacherAWDReviewTeamItemData` 等 response DTO 已被 teacher / platform 共用
- `api/teacher/awd-reviews.ts`、`api/teaching/awd-reviews.ts` 里的 page data / normalize helper 继续沿用 teacher 前缀
- `widgets/awd-review-workspace/model/presentation.ts` 里的 `TeacherAwdReview*` summary types 与 `TEACHER_AWD_REVIEW_*_COPY` 也已经落在 shared widget owner

最小正确改动是：

- 把这组 shared AWD review contract 名称统一收口成中性 `AwdReview*`
- teacher / teaching API 保留 endpoint function 名，但返回类型与内部 normalize / raw contract 改为中性命名
- AWD review shared widget、feature model、team drawer、测试同步切到中性命名

本轮不改：

- `listTeacherAWDReviews()`、`getTeacherAWDReview()`、`exportTeacherAWDReviewArchive()`、`exportTeacherAWDReviewReport()` 这些 teacher endpoint function 名
- `TeacherAWDReviewIndex` / `TeacherAWDReviewDetail` 这类 route / view 名称
- teacher 端 UI copy 文案本身的内容含义，只收口共享 owner 的命名

## Files to modify
- .harness/reuse-decisions/awd-review-shared-contract-naming-neutralization.md
- docs/plan/impl-plan/2026-05-28-awd-review-shared-contract-naming-neutralization-plan.md
- docs/reviews/frontend/2026-05-28-awd-review-shared-contract-naming-neutralization-review.md
- code/frontend/src/api/contracts.ts
- code/frontend/src/api/teacher/awd-reviews.ts
- code/frontend/src/api/teaching/awd-reviews.ts
- code/frontend/src/api/awd-reviews.ts
- code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue
- code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue
- code/frontend/src/widgets/awd-review-workspace/model/presentation.ts
- code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue
- code/frontend/src/widgets/awd-review-workspace/model/presentation.test.ts
- code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts
- code/frontend/src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts
- code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts
- code/frontend/src/api/__tests__/teacher.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- AWD review 的共享 response DTO、shared widget presentation helper 和 copy owner 会统一落到中性 `AwdReview*` 命名。
- teacher / platform 继续共享同一套 AWD review feature / widget，但命名不会再把共享 owner 伪装成 teacher 专属。
- teacher endpoint function 名和 route name 保持不变，因此不会把这刀扩展成权限或路由迁移。
