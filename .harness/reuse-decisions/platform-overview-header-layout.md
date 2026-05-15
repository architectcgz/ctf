# Reuse Decision

## Change type

layout / page / component / test

## Existing code searched

- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/views/platform/PlatformOverview.vue`
- `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/style.css`
- `code/frontend/src/components/platform/images/ImageManageHeroPanel.vue`
- `code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue`
- `code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue`

## Similar implementations found

- `ImageManageHeroPanel` already uses the shared `workspace-page-header` shell with a left intro and a right action group.
- `InstanceManageHeroPanel` already places page actions in the header-right rail without introducing a page-specific header primitive.
- `ContestOrchestrationPage` already uses shared header button primitives and a compact action block.

## Decision

extend_existing

## Reason

This change only realigns an existing admin workspace header. The page already uses the shared `workspace-page-header` contract, `header-actions`, and `header-btn` primitives, so the correct fix is to extend that pattern locally with an `overview-page-header` modifier and a 2x2 action grid. Creating a new page header primitive would duplicate an already approved workspace layout pattern for a one-page spacing issue.

## Files to modify

- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`

## After implementation

- Keep the page on the shared workspace header system.
- Preserve the 2x2 action layout and top-left title alignment as page-local refinements.
