# Reuse Decision

## Change type
frontend architecture / shared panel owner normalization

## Existing code searched
- code/frontend/src/components/dashboard/student/StudentTimelinePage.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts
- code/frontend/src/features/student-analysis-workspace/ui/StudentAnalysisPage.vue
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/views/__tests__/metricPanelSurfaceOwnership.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 学生仪表盘前一轮已经把 student-only 的 4 个 page-sized panel 迁入 `features/student-dashboard/ui`，只剩 `StudentTimelinePage.vue` 因为 teacher 学员洞察复用而暂留。
- `StudentInsightPanel.vue` 当前直接消费 `StudentTimelinePage.vue`，说明这个 UI 已经具备跨 feature 的共享属性，不再适合作为 `student-dashboard` feature-owned page 继续存在。
- 项目里类似复用展示块通常收口到中立 `components/*` 入口，再由多个 feature / workspace 组合，而不是继续让一个 feature 独占 page 名义 owner。

## Decision
refactor_existing

## Reason
最小正确改动不是把 `StudentTimelinePage.vue` 迁入某个单一 feature，而是把它改成中立的训练时间线 panel，停止使用 `components/**Page.vue` 的页面命名，并让 student dashboard 和 teacher 学员洞察都通过这个共享 panel 入口读取它。这样可以清掉最后一条 student dashboard `legacy component page`，同时不引入新的 feature 间所有权歧义。

## Files to modify
- .harness/reuse-decisions/training-timeline-panel-owner-normalization.md
- docs/plan/impl-plan/2026-05-27-training-timeline-panel-owner-normalization-plan.md
- docs/reviews/frontend/2026-05-27-training-timeline-panel-owner-normalization-review.md
- code/frontend/src/components/training/TrainingTimelinePanel.vue
- code/frontend/src/components/dashboard/student/utils.ts
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts
- code/frontend/src/components.d.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/studentJournalSoftStyles.test.ts
- code/frontend/src/views/__tests__/surfaceBackground.test.ts
- code/frontend/src/views/__tests__/rootHeroLayout.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/metricPanelSurfaceOwnership.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `TrainingTimelinePanel.vue` 会成为中立共享入口，继续保留现有训练时间线的分页、分组和 metric-panel 展示语义。
- student dashboard 与 teacher 学员洞察都会消费同一个共享 panel，而不再通过 `StudentTimelinePage.vue` 这个误导性的 page owner 名称复用。
- `legacyComponentPageAllowlist` 里 student dashboard 这条存量会被清空。
