# Reuse Decision

## Change type
frontend refactor / feature ui owner cleanup

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentReviewWorkspace.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentInsightShared.ts`
- `code/frontend/src/features/teaching/student-analysis-review/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `student-analysis-review` 已经承接 `useReviewWorkspace()`、`useSubmissionReviewFlows()`、`useReviewArchiveExportFlow()` 和 `InterventionPanel.vue` 这组教学复盘 workflow owner。
- `student-analysis-workspace` 当前更适合保留 page shell、tab 组合与学员洞察总体装配，而不是继续持有 review/writeup/manual review/evidence 细分 UI。

## Decision
refactor_existing

## Reason
当前 `student-analysis-workspace` 内部混放了两类职责：

- page shell / tab / 学员洞察总装配
- review workspace、题解列表、人工审核、证据链这些已经由 `student-analysis-review` model owner 驱动的 UI

最小正确收口方式是：

- 把 `StudentReviewWorkspace.vue`、`StudentInsightAttackSessionsSection.vue`、`StudentInsightWriteupsSection.vue`、`StudentInsightManualReviewSection.vue` 和 `studentInsightShared.ts` 收回 `features/teaching/student-analysis-review/ui`
- 让 `StudentInsightPanel.vue` 只通过 `student-analysis-review` 的 public API 组合这组 review UI
- 同步迁移邻近测试 owner，并更新 raw-source 护栏与 backlog 进展

本轮不做：

- 不改 `useStudentAnalysisPage()`、`useReviewWorkspace()`、`useSubmissionReviewFlows()` 的 workflow
- 不把 `StudentInsightOverviewSection.vue`、`StudentInsightRecommendationsSection.vue` 上提到 `entities/*`
- 不继续拆 `StudentAnalysisPage.vue` 的 tab shell

## Files to modify
- `.harness/reuse-decisions/student-analysis-review-ui-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-student-analysis-review-ui-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-student-analysis-review-ui-owner-cleanup-review.md`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentReviewWorkspacePresentation.test.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `student-analysis-review` 会同时承接 review / writeup / manual review / evidence 的 model 和 UI owner。
- `student-analysis-workspace` 会退回 page shell、洞察总装配和非 review 区块 owner。
- 后续如果继续拆 `student-analysis`，可以直接围绕 workspace 与 review 两个 feature 的边界推进，而不再先做文件落点纠偏。
