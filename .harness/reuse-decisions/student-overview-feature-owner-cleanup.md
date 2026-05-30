# Reuse Decision

## Change type
component / test / docs

## Existing code searched
- `code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue`
- `code/frontend/src/components/dashboard/student/StudentOverviewVariantSwitcher.vue`
- `code/frontend/src/components/dashboard/student/overviewProps.ts`
- `code/frontend/src/features/student-dashboard/ui/StudentOverviewPage.vue`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`
- `code/frontend/src/__tests__/studentJournalButtonStyles.test.ts`

## Similar implementations found
- `features/challenge-list/ui/ChallengeDirectoryPanel.vue`、`features/platform/awd-challenges/ui/AwdChallengeLibrarySection.vue` 都已经承接“历史 components 目录里、但运行时只服务单一 feature”的 owner 收口。
- `StudentOverviewPage.vue` 现在已经是 `features/student-dashboard/ui` 的页面入口，说明 overview 展示块与 props 契约继续留在 `components/dashboard/student` 只是迁移未收尾。

## Decision
refactor_existing

## Reason
- `StudentOverviewStyleEditorial.vue` 和 `overviewProps.ts` 的运行时 owner 明确属于 `features/student-dashboard/ui/StudentOverviewPage.vue`。
- `StudentOverviewVariantSwitcher.vue` 只是在旧目录里包了一层同义转发，没有独立 consumer，适合一并清理。
- 本轮只收 student dashboard overview 这一组，不扩展到 timeline、difficulty、category 等其他 student dashboard 子块。

## Files to modify
- `.harness/reuse-decisions/student-overview-feature-owner-cleanup.md`
- `code/frontend/src/features/student-dashboard/ui/StudentOverviewStyleEditorial.vue`
- `code/frontend/src/features/student-dashboard/ui/studentOverviewProps.ts`
- `code/frontend/src/features/student-dashboard/ui/StudentOverviewPage.vue`
- `code/frontend/src/features/student-dashboard/ui/index.ts`
- `code/frontend/src/features/student-dashboard/index.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue`
- `code/frontend/src/components/dashboard/student/StudentOverviewVariantSwitcher.vue`
- `code/frontend/src/components/dashboard/student/overviewProps.ts`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`
- `code/frontend/src/__tests__/studentJournalButtonStyles.test.ts`
- `docs/plan/impl-plan/2026-05-30-student-overview-feature-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-05-30-student-overview-feature-owner-cleanup-review.md`

## After implementation
- `components/dashboard/student/` 不再保留 overview 展示块与 props 契约。
- student overview 的展示 owner 与 props 契约将统一落在 `features/student-dashboard/ui`。
