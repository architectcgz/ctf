# Reuse Decision

## Change type
frontend shared style contract convergence

## Existing code searched
- `code/frontend/src/assets/styles/teacher-surface.css`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

## Similar implementations found
- `teacher-surface.css` 已经承接 teacher summary / report summary 的基础 metric-panel contract，适合作为 teacher 页通用 summary surface 的唯一共享 owner。
- `AwdReviewSummaryPanel.vue` 已经把 teacher summary 封成稳定复用块，说明 teacher summary family 优先应继续走显式共享 contract，而不是回到各页本地复制 class。
- `TeacherDashboardPage.vue`、`TeacherDashboardPortraitPanel.vue`、`TeacherDashboardTrendPanel.vue` 目前仍在各自持有相同的 summary grid / note card 基础样式，属于 feature 内局部重复。

## Decision
extend_existing

## Reason
- 这轮不是重做 teacher 页面视觉，而是继续把“teacher summary / dashboard summary / student-analysis KPI 测试护栏”这些已确认会反复漂移的基础 contract 收到明确 owner。
- 对 teacher summary family，最小正确路径是继续复用 `teacher-surface.css` 作为跨页面基础 owner，只把 dashboard feature 内重复的 summary note contract 抽到 feature 共享文件。
- 对 student-analysis，最小正确路径是把仍指向旧类名的测试护栏改到真实 owner，不再让测试继续放大已完成的收口债。

## Files to modify
- `.harness/reuse-decisions/teacher-summary-contract-convergence.md`
- `docs/plan/impl-plan/2026-06-04-teacher-summary-contract-convergence-plan.md`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/teacherDashboardSummary.css`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`

## After implementation
- `student-analysis` 旧测试类名漂移被清掉，护栏重新指向当前共享 KPI contract。
- `teacher-dashboard` 概览 / 画像 / 趋势三处 summary strip 不再各自维护相同的 grid / note card 基础样式，dashboard feature 内有明确共享 owner。
- teacher summary family 不新增第二份全局 shared contract，继续以 `teacher-surface.css` 为跨页面基础 owner。
