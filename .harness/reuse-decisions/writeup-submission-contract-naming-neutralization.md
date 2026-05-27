# Reuse Decision

## Change type
+api / feature / component / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 近几轮已经先把 platform 侧对 teacher route view、teacher API owner 的直连逐步摘掉；这说明当前更合适的下一刀不是继续复制 owner，而是开始收 contract 命名本身。
- `TeacherSubmissionWriteupItemData` 同时被 `useChallengeWriteupManagement` 和 `useSubmissionReviewFlows` 消费，是当前最适合先做中性化的一组共享 DTO。

## Decision
refactor_existing

## Reason
- `TeacherSubmissionWriteupItemData` 已经不是教师专属数据模型，platform 题解管理和 teacher 学生分析都在用这个 DTO 名。
- 最小正确切片是把这一个共享 DTO 先改成中性命名 `WriteupSubmissionItemData`，并同步 API client、teacher/platform feature、teacher 组件与相关文档。
- 本轮不顺手改 `TeacherClassItem`、`TeacherAttackSessionQuery`、`TeacherManualReviewSubmission*`，避免把多个共享领域命名债混进同一次提交。

## Files to modify
- `.harness/reuse-decisions/writeup-submission-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-writeup-submission-contract-naming-neutralization-implementation-plan.md`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-writeup-submission-contract-naming-neutralization-review.md`

## After implementation
- 如果这组 DTO 收口稳定，后续 `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 可以沿同样策略分 slice 继续去 teacher 化，而不是再从 owner 层兜命名噪音。
