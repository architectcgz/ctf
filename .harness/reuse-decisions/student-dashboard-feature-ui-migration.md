# Reuse Decision

## Change type
frontend architecture / feature UI migration / student dashboard

## Existing code searched
- code/frontend/src/views/dashboard/DashboardView.vue
- code/frontend/src/components/dashboard/student/dashboardPanelRegistry.ts
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/dashboard/student/StudentOverviewPage.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/dashboard/student/StudentTimelinePage.vue
- code/frontend/src/components/dashboard/student/StudentOverviewVariantSwitcher.vue
- code/frontend/src/features/student-dashboard/index.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `PlatformOverviewPage.vue`、`TeacherDashboardPage.vue`、`UserGovernancePage.vue` 这些 page-sized UI 都已经按“route / page owner 留在 feature model，具体 page shell 迁入 `features/*/ui`”的模式收口。
- `DashboardView.vue` 现在已经只保留 page shell、tab 切换和动态挂载，没有直接持有学生仪表盘数据流，符合继续把 panel registry 与 page-sized UI 收回 feature 的前提。
- `StudentTimelinePage.vue` 目前仍被 `components/teacher/StudentInsightPanel.vue` 直接复用，这个 page-sized UI 还带着 teacher 共享消费面，不适合和本轮四个仅供 student dashboard route 使用的页面一起硬迁。

## Decision
refactor_existing

## Reason
这轮目标不是重做学生仪表盘，而是继续沿用已经验证过的 feature UI 迁移模式，先把只服务学生仪表盘 route 的四个 page-sized panel 和 `dashboardPanelRegistry.ts` 从 `components/dashboard/student` 收口到 `features/student-dashboard/ui`，减少 `legacyComponentPageAllowlist` 与 `componentFeatureImportAllowlist`。`StudentTimelinePage.vue` 因为还有 teacher 复用，先留到下一刀按共享 panel owner 单独处理。

## Files to modify
- .harness/reuse-decisions/student-dashboard-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-student-dashboard-feature-ui-migration-plan.md
- docs/reviews/frontend/2026-05-27-student-dashboard-feature-ui-migration-review.md
- code/frontend/src/features/student-dashboard/index.ts
- code/frontend/src/features/student-dashboard/ui/index.ts
- code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts
- code/frontend/src/features/student-dashboard/ui/StudentCategoryProgressPage.vue
- code/frontend/src/features/student-dashboard/ui/StudentDifficultyPage.vue
- code/frontend/src/features/student-dashboard/ui/StudentOverviewPage.vue
- code/frontend/src/features/student-dashboard/ui/StudentRecommendationPage.vue
- code/frontend/src/views/dashboard/DashboardView.vue
- code/frontend/src/components/dashboard/student/StudentOverviewVariantSwitcher.vue
- code/frontend/src/components.d.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/views/__tests__/studentOverviewEntrypoint.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/studentJournalButtonStyles.test.ts
- code/frontend/src/views/__tests__/studentJournalSoftStyles.test.ts
- code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/surfaceBackground.test.ts
- code/frontend/src/views/__tests__/rootHeroLayout.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 学生仪表盘的 4 个 page-sized panel 与 panel registry 会通过 `features/student-dashboard` public API 对外暴露，不再继续占用 `components/**Page.vue` 例外。
- `StudentTimelinePage.vue` 暂时保留在原位置，作为 teacher 学员洞察的现有共享 panel，避免在这一刀里把共享 owner 与 feature UI 迁移混在一起。
- `StudentOverviewVariantSwitcher.vue` 会改成直接桥接稳定视觉实现，不再依赖已迁出的 `StudentOverviewPage.vue` 旧路径。
