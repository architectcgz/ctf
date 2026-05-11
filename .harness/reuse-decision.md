# Reuse Decision

This file is only for the current task and may be overwritten.
Durable reuse knowledge belongs in `harness/reuse/index.yaml`; append-only summaries belong in `harness/reuse/history.md`.

## Change type
- component
- list
- layout

## Existing code searched
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/common/AppCard.vue
- code/frontend/src/style.css
- code/frontend/src/components/teacher/student-management/StudentManagementPage.vue
- code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewContestRow.vue
- code/frontend/src/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue
- code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## Similar implementations found
- code/frontend/src/style.css
- code/frontend/src/components/teacher/student-management/StudentManagementPage.vue
- code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewContestRow.vue

## Decision
- reuse_existing

## Reason
- 推荐训练任务本质是可点击条目列表，不需要继续使用 `AppCard variant="action"` 的左侧强调边框。
- 项目已有 `workspace-directory-list`、`workspace-directory-grid-row`、`workspace-directory-row-title`、`workspace-directory-row-subtitle`、`workspace-directory-row-btn` 这组通用目录列表样式，适合复用到推荐列表。

## Files to modify
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## After implementation
- No new durable reuse entry was added; this follows the existing workspace directory list pattern.
