# Reuse Decision

This file is only for the current task and may be overwritten by the next protected change.

## Change type
component / page / layout

## Existing code searched
- code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue
- code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue
- code/frontend/src/api/contracts.ts
- code/frontend/src/api/teacher/writeups.ts
- code/frontend/src/style.css

## Similar implementations found
- `StudentInsightManualReviewSection.vue` already owns the manual review detail workflow and review actions.
- `StudentInsightWriteupsSection.vue` already owns the published writeup list, moderation actions, and writeup pagination.
- Shared `workspace-directory-list` / `workspace-directory-grid-row` styles provide the flat directory list pattern used by teacher/admin pages.

## Decision
extend_existing

## Reason
The user-facing problem is information architecture drift inside the writeups tab: published writeups and manual-review submissions were rendered as separate blocks. The existing data contracts remain separate, so this change extends the writeups section to display review status and review entry actions in the same list while keeping the existing manual-review detail workflow and API ownership unchanged.

## Files to modify
- code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
