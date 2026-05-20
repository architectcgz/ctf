# Reuse Decision

## Change type
- component
- hook
- page
- test

## Existing code searched
- code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue
- code/frontend/src/features/platform-overview/model/usePlatformOverviewWorkspace.ts
- code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/views/platform/PlatformOverview.vue

## Similar implementations found
- `PlatformOverviewPage.vue` already centralizes admin dashboard resource display in one page-level component.
- `usePlatformOverviewWorkspace.ts` already owns dashboard data shaping and local display formatting.
- `TeacherDashboardPage.vue` uses the same workspace shell conventions, so the admin page should keep the existing owner boundaries instead of introducing a new formatting component.

## Decision
extend_existing

## Reason
This task is a display-only refinement on an existing platform overview surface. The correct reuse path is to extend the current formatting helper and page bindings rather than create a new CPU widget or split the dashboard into another component. The existing hook already owns dashboard presentation helpers, and the page already owns the container hotspot rendering, so this change stays within the current ownership boundary.

## Files to modify
- code/frontend/src/features/platform-overview/model/usePlatformOverviewWorkspace.ts
- code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue
- code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts
