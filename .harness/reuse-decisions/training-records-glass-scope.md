# Reuse Decision

## Change type
component / layout

## Existing code searched
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue
- code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightLoadingSurface.vue
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
- code/frontend/src/assets/styles/workspace-glass.css
- code/frontend/src/assets/styles/journal-notes.css
- code/frontend/src/features/teaching/student-analysis-workspace

## Similar implementations found
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue
- code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightLoadingSurface.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue
- code/frontend/src/assets/styles/workspace-glass.css
- code/frontend/src/assets/styles/journal-notes.css

## Decision
extend_existing

## Reason
本次重新设计训练记录面板的 surface 结构，不新增业务组件、不新增全局 surface token。排查后确认训练记录页的问题不只是 `workspace-glass-region` / `workspace-glass-metric-surface` 挂载过宽，而是 loading 与 loaded 分别由 `StudentInsightTimelineSection` 和 `TrainingTimelineContent` 两套 DOM/样式 owner 承担，导致刷新加载时看到整块玻璃屏，加载后又切到另一套结构。正确收口方式是让 `StudentInsightTimelineSection` 只做入口透传，由 `TrainingTimelineContent` 统一承载 loading / empty / loaded，指标卡骨架和列表骨架只出现在对应元素区域。

## Files to modify
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## After implementation
- 这是当前任务的局部设计收口，不需要新增长期 reuse-index。
