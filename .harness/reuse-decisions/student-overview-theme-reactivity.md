# Reuse Decision

## Change type
page / component / test

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Similar implementations found
- `StudentAnalysisPage.vue` 已经通过 `--workspace-brand`、`--workspace-brand-ink`、`--workspace-brand-soft` 统一提供教师侧学员分析页的主题品牌变量。
- `StudentAnalysisOverviewHeroPanel.vue` 的三张 summary card 已经共用 `summary-card` 这一套 metric-panel 变量桥接，问题只在第二张卡把 accent 固定成了 `--color-success`。

## Decision
refactor_existing

## Reason
这次不是新增主题系统，也不是抽共享卡片组件。最小正确改动是在现有 `summary-card` 样式桥里，把第二张卡从固定 `success` 语义改为消费当前 workspace 的主题品牌变量，并补一条测试，确保该卡不再回退成写死绿色。

## Files to modify
- `.harness/reuse-decisions/student-overview-theme-reactivity.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- `/academy/classes/:className/students/:studentId` 的 `#student-overview` 第二张 summary card 不再固定绿色。
- 该卡片颜色跟随当前 teacher workspace 主题品牌变量变化。
- 测试会明确阻止把 `summary-card--completion` 重新绑回 `--color-success`。
