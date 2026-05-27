# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPortraitPanel.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardStudentInsightPanel.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardTrendPanel.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardReviewPanel.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardInterventionPanel.vue
- code/frontend/src/views/teacher/TeacherDashboard.vue
- code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts
- code/frontend/src/features/teacher-dashboard/model/useDashboardMetrics.ts
- code/frontend/src/features/teacher-dashboard/index.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `features/challenge-writeup-editor/ui/*` 与 `features/platform-overview/ui/*` 已经证明，单一 feature 的 page-sized UI 可以直接挂在 `features/*/ui`，而不是继续通过 `components/*Page.vue` 中转。
- `TeacherDashboardPage.vue` 当前直接消费 `useDashboardMetrics()`，并且只被 `views/teacher/TeacherDashboard.vue` 使用，符合 `feature-owned UI` 的判定条件。
- 教师总览内部的 portrait / insight / trend / review / intervention 分区已经是稳定子面板，这次可以继续只迁 page shell owner，不顺手扩大成整组子组件搬迁。

## Decision
refactor_existing

## Reason
这次不是新增教师总览能力，而是继续沿用已经落地的 `feature-owned UI` 规则收口 legacy component page。最小正确改动是把 `TeacherDashboardPage.vue` 迁到 `features/teacher-dashboard/ui/`，让 route view 继续只组合 `useDashboardPage()` 与 feature ui，并移除它对应的 component->feature 与 legacy component page 例外。

## Files to modify
- .harness/reuse-decisions/teacher-dashboard-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-teacher-dashboard-feature-ui-migration-implementation-plan.md
- docs/reviews/frontend/2026-05-27-teacher-dashboard-feature-ui-migration-review.md
- code/frontend/src/features/teacher-dashboard/index.ts
- code/frontend/src/features/teacher-dashboard/ui/*
- code/frontend/src/views/teacher/TeacherDashboard.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
- code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果 `TeacherDashboardPage` 顺利收口，下一批同模式候选可继续看 `ContestOrchestrationPage.vue` 或其它仍直接依赖单一 feature model 的 legacy component page。
- 本轮只处理教师总览 page shell，不额外迁移 class management / student analysis 等 teacher 其它页面。
