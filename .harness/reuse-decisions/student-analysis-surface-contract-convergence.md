# Reuse Decision

## Change type
frontend refactor / loading-empty-content surface contract convergence

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightReviewSections.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## Similar implementations found
- `StudentInsightRecommendationsSection.vue` 已经收口成 “同一列表容器承接 loading / empty / list” 的三态模式。
- `TrainingTimelinePanel.vue` 自己承接 timeline 的真实 content surface，适合保留 content owner，只统一 loading surface contract。
- `workspace-directory-list` 已经提供目录列表壳体变量，适合作为 recommendations 这类列表 surface 的内容语义层，而不是继续在每个 panel 里复制 glass 背景。

## Decision
refactor_existing

## Reason
- 当前问题不是单个 tab 样式错位，而是 student-analysis 这组 section 的 surface owner 和三态 contract 没有收口。
- overview / timeline / writeups / evidence / manual-review 各自维护 glass shell、shimmer、empty / loading owner，导致同类页面一改再漂。
- 最小正确改动不是再补一批局部 `*-glass` class，而是在 `student-analysis` 语义边界内抽一个共享 surface contract，让 workspace / review 两侧都消费同一 owner。

## Files to modify
- `.harness/reuse-decisions/student-analysis-surface-contract-convergence.md`
- `docs/plan/impl-plan/2026-06-04-student-analysis-surface-contract-convergence-plan.md`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightLoadingSurface.vue`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightStateSurface.vue`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- student-analysis 这组 section 的 glass shell、skeleton shimmer、empty/loading/content owner 会收口到同一组共享 primitive。
- recommendations / writeups / evidence / manual-review 会显式声明自己的 state surface，不再在各文件里复制整套 glass 背景。
- timeline 保留 `TrainingTimelinePanel.vue` 的 content owner，但 loading surface 改走同一共享 contract，避免组合组件里继续内联局部 glass 模板。
