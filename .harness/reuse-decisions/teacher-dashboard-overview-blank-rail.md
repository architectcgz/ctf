# Reuse Decision

## Change type
component / page

## Existing code searched
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
- code/frontend/src/components/teacher/teacher-workspace-subpanel.css
- code/frontend/src/assets/styles/teacher-surface.css

## Similar implementations found
- `TeacherDashboardPage.vue` 里的 overview hero 已经采用标准 `workspace-hero` 双列结构，右侧 `hero-rail` 就是当前空白占位的 owner。
- `teacher-workspace-subpanel.css` 已经为 `workspace-subpanel` 提供共享卡片壳层，说明这次不需要新增新的空态卡片组件或局部容器。
- `TeacherDashboard.test.ts` 已经覆盖 `hero-rail workspace-subpanel` 的存在和样式来源，适合在原测试里收口“保留外壳但移除内容”的断言。

## Decision
refactor_existing

## Reason
这次不是新增一个空状态，而是把 overview hero 右侧 rail 的内容清空，同时保留现有布局和留白。直接复用 `hero-rail workspace-subpanel` 现有壳层，比删除整列或新增占位节点更符合现有 workspace hero 合同，改动也最小。

## Files to modify
- .harness/reuse-decisions/teacher-dashboard-overview-blank-rail.md
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
