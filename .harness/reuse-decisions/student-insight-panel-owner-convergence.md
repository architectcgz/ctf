# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`

## Decision
refactor_existing

## Reason
- `StudentInsightPanel.vue` 已经在同目录下把题解、人工审核、复盘工作台抽成独立 section 组件，说明当前 owner 方向不是新建平行页面或 feature，而是继续沿用现有 `student-insight` 分区做收口。
- `StudentAnalysisPage.vue` 已经通过单个 `StudentInsightPanel` 挂载承接 tab 切换，页面 owner 与 query / 数据加载 owner 都在上层，当前 debt 集中在 `StudentInsightPanel.vue` 仍同时承担 overview、recommendations 与局部样式组合。
- 因此本轮应复用现有 `student-insight` 目录和 `SectionCard` / `AppEmpty` / `ChallengeCategoryDifficultyPills` 这些现成展示壳，继续把剩余 section 从 `StudentInsightPanel.vue` 内部抽出，而不是新建第二套 workspace 或把逻辑回退到 `StudentAnalysisPage.vue`。

## Files to modify
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- 如果 `student-insight` 目录的 section 分层成为后续可复用模式，再考虑把线索补进 `.harness/reuse-index/`。
