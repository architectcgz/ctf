# Reuse Decision

## Change type
component

## Existing code searched
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue
- code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue
- code/frontend/src/entities/training-timeline/model/presentation.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/features/**/*

## Similar implementations found
- code/frontend/src/entities/training-timeline/model/presentation.test.ts
- code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.test.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.test.ts
- code/frontend/src/features/platform/challenges/__tests__/ChallengeManagePage.test.ts

## Decision
extend_existing

## Reason
本次不新增业务组件，也不重排测试目录。训练记录已经有 `TrainingTimelineContent.vue` 作为 loading / empty / loaded 的单一视觉 owner，因此测试瘦身应优先扩展这个组件自己的渲染测试，减少教师页源码字符串断言对内部 class 细节的重复锁定。

## Files to modify
- code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.test.ts
- code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## After implementation
- 这是当前任务的局部测试分层收口，不需要新增长期 reuse-index。
