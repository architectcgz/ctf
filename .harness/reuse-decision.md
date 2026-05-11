# Reuse Decision

This file is only for the current task and may be overwritten.
Durable reuse knowledge belongs in `harness/reuse/index.yaml`; append-only summaries belong in `harness/reuse/history.md`.

## Change type
- component
- layout

## Existing code searched
- code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.vue
- code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.test.ts
- code/frontend/src/assets/styles/journal-notes.css
- code/frontend/src/assets/styles/teacher-surface.css
- code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue

## Similar implementations found
- code/frontend/src/assets/styles/journal-notes.css
- code/frontend/src/assets/styles/teacher-surface.css

## Decision
- extend_existing

## Reason
- 证据复盘工作台已经使用 `journal` / `teacher-surface` / `metric-panel` 的共享暗色 surface token，不需要新增独立调色板。
- 当前亮边框来自组件局部 observation 状态边框和 KPI 网格缺少 teacher surface 变量桥接；扩展现有 `metric-panel` surface 变体并收敛局部边框即可。

## Files to modify
- code/frontend/src/assets/styles/journal-notes.css
- code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.vue
- code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.test.ts

## After implementation
- No new durable reuse entry was added; this uses the existing dark surface alignment rules and shared metric-panel surface pattern.
