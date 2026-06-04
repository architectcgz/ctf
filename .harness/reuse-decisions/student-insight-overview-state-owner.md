# Reuse Decision

## Change type
component

## Existing code searched
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue
- code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightLoadingSurface.vue
- code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css
- code/frontend/src/shared/ui/common/SectionCard.vue
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## Similar implementations found
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue
- code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightStateSurface.vue
- code/frontend/src/shared/ui/common/SectionCard.vue

## Decision
refactor_existing

## Reason
`StudentInsightOverviewSection.vue` 当前 loading 与 loaded 使用两套外壳：loading 直接渲染 `StudentInsightLoadingSurface`，loaded 才渲染 `SectionCard`。这和训练记录页曾经的问题一致，容易让加载态与完成态表现成两套区域结构。最小正确改动是在现有 overview section 内收口状态 owner，让两个固定 `SectionCard` 承载 loading / empty / loaded 内容，不新增业务组件、不改变父级数据流。

## Files to modify
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## After implementation
- 这是当前任务的局部状态 owner 收口，不需要新增长期 reuse-index。
