> 状态：Current
> 事实源：`api/contracts.ts`、fronted backlog
> 替代：无

# Writeup Submission Detail Contract Naming Neutralization Implementation Plan

## 目标

- 把 `TeacherSubmissionWriteupDetailData` 收口成中性命名 `WriteupSubmissionDetailData`。
- 不改任何行为面、API owner、HTTP path 或页面交互，只清理题解 detail contract 的残余 teacher 语义。

## 非目标

- 本轮不改 manual review DTO；这组 contract 已在上一刀收口完成。
- 本轮不改 `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery`。
- 本轮不新增 writeup detail 的实际消费面或新 API client。

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-writeup-submission-contract-naming-neutralization-review.md`
- `docs/reviews/frontend/2026-05-27-manual-review-submission-contract-naming-neutralization-review.md`

## 当前结论

- `TeacherSubmissionWriteupDetailData` 当前只停留在 contract 层，是一块薄但明确的历史命名残片。
- 这条债不需要再等行为面重构，直接按最小切片改名即可，能进一步降低题解域 contract 的 teacher 语义噪音。

## 任务切片

### Slice 1：收口 writeup detail contract 名称

- 目标：
  - 在 `api/contracts.ts` 提供中性 `WriteupSubmissionDetailData`，移除 `TeacherSubmissionWriteupDetailData`。
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `docs/contracts/api-contract-v1.md`
- review focus：
  - 不误改 `SubmissionWriteupData` 或 `WriteupSubmissionItemData`

### Slice 2：同步 backlog 与 review 证据

- 目标：
  - 在 backlog / review 中记录 detail DTO 已完成收口后的剩余命名债列表。
- 预期改动：
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-writeup-submission-detail-contract-naming-neutralization-review.md`

## 验证

- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退只需撤销 `WriteupSubmissionDetailData` 的命名和相关文档记录，不涉及运行时行为回退。

## 残余风险

- `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract 命名残余还在。
