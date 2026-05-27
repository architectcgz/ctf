# Reuse Decision

## Change type
+api / feature / component / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 题解投稿目录 DTO 已在上一刀从 `TeacherSubmissionWriteupItemData` 收口到 `WriteupSubmissionItemData`，说明这条题解评阅链可以继续沿“共享消费面改中性命名，teacher/platform API owner 保持原位”的方式推进。
- `TeacherManualReviewSubmissionItemData` / `TeacherManualReviewSubmissionDetailData` 当前通过共享 `student-analysis-workspace` 同时服务 teacher 与 platform 两个 route view，已经不是教师专属 contract。

## Decision
refactor_existing

## Reason
- platform 学员分析页和 teacher 学员分析页现在共用同一条 manual review workflow，但 contract 仍保留 teacher 前缀，会继续制造“platform 还在借 teacher 模型”的认知噪音。
- 这组 manual review DTO 的引用面集中在 writeups API client、submission review flow 和 student insight 组件，适合作为一个独立、可审阅的命名收口切片。
- `TeacherSubmissionWriteupDetailData` 当前几乎只停留在 contract 层，先不与 manual review 混切，避免把两个收益结构不同的命名债绑在同一提交里。

## Files to modify
- `.harness/reuse-decisions/manual-review-submission-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-manual-review-submission-contract-naming-neutralization-implementation-plan.md`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-manual-review-submission-contract-naming-neutralization-review.md`

## After implementation
- 如果这组 DTO 收口稳定，后续可以继续按同样策略处理 `TeacherSubmissionWriteupDetailData`、`TeacherClassItem`、`TeacherAWDReviewContestItemData` 和 `TeacherAttackSessionQuery`，逐步把共享 contract 的 teacher 语义从消费面摘干净。
