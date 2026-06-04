# Reuse Decision

## Change type
frontend shared style contract convergence

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`

## Similar implementations found
- `studentInsightSurface.css` 已经是 student-analysis 顶部 glass / loading surface 的共享 owner，说明 overview hero 的 summary 基础 contract 应继续落在这个共享样式入口。
- `StudentAnalysisOverviewHeroPanel.vue` 当前本地 `.summary-strip` 同时承接了 header attached summary 的 `padding: 0`、列数和响应式折叠，属于与 `class-students` 同类的 consumer 局部 contract。
- `teacherDetailSurfaceAlignment.test.ts` 当前仍允许这些 contract 留在组件本地样式，缺少 shared owner 护栏。

## Decision
extend_existing

## Reason
- 这轮不是重做 student-analysis hero 视觉，而是把顶部 summary 的基础展示 contract 收回已有 shared owner。
- 最小正确路径是扩 `studentInsightSurface.css`，让 header summary 的间距和响应式列数由 shared CSS 承接，consumer 只保留业务语义和卡片视觉。

## Files to modify
- `.harness/reuse-decisions/student-analysis-overview-summary-convergence.md`
- `docs/plan/impl-plan/2026-06-04-student-analysis-overview-summary-convergence-plan.md`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- student-analysis 顶部 summary 不再依赖本地 `.summary-strip` 承接 header attached summary contract。
- summary 列数与窄屏折叠回到 shared owner，hero 组件只保留 summary card 自身的视觉变量。
- raw-source 护栏测试会指向 `studentInsightSurface.css`，不再继续容忍 consumer 局部 contract。
