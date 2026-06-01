> 状态：Current
> 事实源：`api/contracts.ts`、AWD review index workspace、frontend backlog
> 替代：无

# AWD Review Contest Item Contract Naming Neutralization Implementation Plan

## 目标

- 把共享 AWD review 赛事目录项 DTO `TeacherAWDReviewContestItemData` 收口成中性命名 `AwdReviewContestItemData`。
- 保持 teacher / platform AWD 复盘目录的筛选、分页、跳转和 API 行为不变。

## 非目标

- 本轮不改 `TeacherAWDReviewArchiveData`、`TeacherAWDReviewRoundItemData`、`TeacherAWDReviewTeamItemData` 等 teacher response contract。
- 本轮不改 `listTeacherAWDReviews()`、`listPlatformAWDReviews()` 的 HTTP path、query 参数或 summary 结构。
- 本轮不调整 AWD review index 的 UI copy、筛选 owner 或权限逻辑。

## 输入依据

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

## 当前结论

- `TeacherAWDReviewContestItemData` 当前已经落在共享 AWD review index contract 层，被 teacher / platform 两边一起消费，继续带 teacher 前缀不符合实际 owner 语义。
- 这条债的变化面主要停留在 DTO 名称和类型引用，不需要扩大到 archive 详情 contract；因此可以保持为一刀较小的命名收口。

## 任务切片

### Slice 1：收口 AWD review contest item contract 名称

- 目标：
  - 在 `api/contracts.ts` 提供中性 `AwdReviewContestItemData`，并同步 AWD review archive 对这条 contest 字段的引用。
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/awd-reviews.ts`
  - `code/frontend/src/api/teaching/awd-reviews.ts`
  - `docs/contracts/api-contract-v1.md`
- review focus：
  - AWD review list / archive 的字段结构、normalize 结果和分页 summary 不变

### Slice 2：同步 AWD review index feature / widget / test 消费面

- 目标：
  - 让 AWD review index 的 feature model、widget props 和相关测试全部切到 `AwdReviewContestItemData`。
- 预期改动：
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexFilters.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts`
- review focus：
  - status filter、row props、reviewRows 派生和 widget typing 不回归

### Slice 3：同步 backlog 与 review 证据

- 目标：
  - 在 backlog 和 review 里记录 AWD review 目录项 DTO 已完成收口，以及当前这组共享 contract 命名债的状态变化。
- 预期改动：
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-awd-review-contest-item-contract-naming-neutralization-review.md`

## 验证

- `npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts src/api/__tests__/teacher.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按 DTO 命名切：撤销 `AwdReviewContestItemData` 与相关消费面引用即可，不涉及 AWD review 请求行为或 UI 流程回退。

## 残余风险

- 共享 AWD review 目录项 DTO 收口后，teacher 前缀剩余主要停留在明确 teacher owner 的 archive / response contract，需要后续按实际共享情况再判断是否继续切片。
