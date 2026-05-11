# Reuse Decision

在新增或修改 `page / component / hook / service / store / api / form / table / modal / layout / schema` 之前，先更新本文件。

## Change type
- component
- style

## Existing code searched
- code/frontend/src/components/dashboard/student
- code/frontend/src/components/errors
- code/frontend/src/views/errors
- code/frontend/src/views
- code/frontend/src/features
- code/frontend/src/widgets

## Similar implementations found
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue
- code/frontend/src/assets/styles/journal-soft-surfaces.css
- code/frontend/src/views/instances/InstanceList.vue
- code/frontend/src/views/notifications/NotificationList.vue
- code/frontend/src/components/common/InstancePanel.vue

## Decision
- extend_existing

## Reason
- StudentDifficultyPage and StudentRecommendationPage already follow the shared journal-soft-surface pattern.
- The right move is to extend journal-soft-surfaces.css with a shared secondary button state instead of keeping page-local button variants.
- ErrorStatusShell already serves all error pages, so its actions should reuse the existing ui-btn primary/secondary pattern instead of keeping a private button implementation.

## Files to modify
- code/frontend/src/assets/styles/journal-soft-surfaces.css
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/errors/ErrorStatusShell.vue
- code/frontend/src/views/__tests__/studentJournalButtonStyles.test.ts
- code/frontend/src/views/__tests__/studentJournalSoftStyles.test.ts
- code/frontend/src/views/errors/__tests__/AdditionalErrorViews.test.ts
- code/frontend/src/views/errors/__tests__/ForbiddenView.test.ts
- code/frontend/src/views/errors/__tests__/NotFoundView.test.ts
- code/frontend/src/views/errors/__tests__/errorViewVisualParity.test.ts
