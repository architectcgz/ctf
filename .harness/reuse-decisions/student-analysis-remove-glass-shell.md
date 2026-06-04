# Reuse Decision

## Change type
page shell / visual polish

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/assets/styles/surface-shell-background.css`

## Similar implementations found
- `StudentAnalysisPage.vue` 已经是学生分析页的 shell owner，当前外层玻璃背景就是通过 `workspace-shell` 这层共享 surface 承接的。
- 共享 `workspace-shell` 被多个教师/平台工作台复用，这次不适合改全局样式，而是应该在学生分析页本地覆写。

## Decision
extend_existing

## Reason
- 用户明确要求“去掉玻璃屏”，目标是学生分析页 route shell 的视觉形态，不是内部卡片或推荐列表。
- 最小正确改动是在 `StudentAnalysisPage.vue` 本地去掉外层 shell 的背景、边框和阴影，保留原有 tabs、内容布局和内部 panel 结构。

## Files to modify
- `.harness/reuse-decisions/student-analysis-remove-glass-shell.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- 学员分析页外层不再显示玻璃壳背景。
- 内部推荐任务、概览、题解和证据等内容块维持现有布局与交互。
