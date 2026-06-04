# Reuse Decision

## Change type
component styling / visual polish

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightReviewSections.vue`
- `code/frontend/src/assets/styles/journal-notes.css`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## Similar implementations found
- `StudentInsightWriteupsSection.vue` 和 `StudentInsightManualReviewSection.vue` 都通过本地 class 组合共享 `metric-panel-card`，适合在组件内覆写 surface 变量，避免影响全局教师 summary 卡。
- 共享 `journal-notes.css` 是多个页面都在复用的 metric panel owner，这次不应直接改全局。

## Decision
extend_existing

## Reason
- 用户这次明确要去掉学生分析页内部元素的玻璃感，重点是 `writeups` tab 里的 KPI 卡、题目列表强调面和审核详情面。
- 最小正确改动是在 `student-analysis-review` 组件内把相关 surface 覆写为透明，同时保留边界线、层级和交互，不扩散到其他教师页。

## Files to modify
- `.harness/reuse-decisions/student-analysis-remove-inner-glass.md`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- `writeups` / `manual-review` 内部 KPI 卡与详情区不再有玻璃背景和阴影。
- 题目列表和审核详情仍保留分隔线与状态信息，但视觉上改为更平的 workspace 内页。
