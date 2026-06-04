# Reuse Decision

## Change type
frontend shared style contract convergence

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSections.css`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

## Similar implementations found
- `studentInsightSections.css` 已经承接了 student-analysis review section card、state surface 和 KPI card 的基础 owner，说明 KPI grid 的列数 contract 也应继续落在这里。
- `StudentInsightWriteupsSection.vue` 仍在本地 `.writeup-kpi-grid` 和媒体查询里承接 3 列与折叠规则。
- `StudentInsightManualReviewSection.vue`、`StudentInsightAttackSessionsSection.vue` 仍通过 `md:grid-cols-*` 或 `teacher-summary-grid` 混合表达列数 contract，owner 不一致。

## Decision
extend_existing

## Reason
- 这轮不是重做 review section 视觉，而是把三处 review KPI grid 的基础列数和响应式折叠 contract 收回共享 owner。
- 最小正确路径是扩 `studentInsightSections.css`，用显式 shared class 承接 3 列 / 4 列 KPI grid，而不是继续让单个 consumer 用局部媒体查询或 utility class 兜底。

## Files to modify
- `.harness/reuse-decisions/student-analysis-review-kpi-grid-convergence.md`
- `docs/plan/impl-plan/2026-06-04-student-analysis-review-kpi-grid-convergence-plan.md`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSections.css`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

## After implementation
- student-analysis review 三处 KPI grid 不再各自持有列数和窄屏折叠规则。
- `studentInsightSections.css` 会成为 review KPI grid 列数 contract 的单点 owner。
- raw-source 护栏测试会指向 shared class，而不是继续容忍 `writeup-kpi-grid`、`md:grid-cols-*`、`teacher-summary-grid` 的混合表达。
