# Reuse Decision

## Change type
page / component

## Existing code searched
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue

## Similar implementations found
- `TeacherDashboardStudentInsightPanel.vue` 已经使用“标题左侧 + tags 右侧 + 说明文案下一行”的本地布局。
- `Trend / Review / Intervention` 三个 panel 都是同一页下的列表型条目，当前 chips 或 meta 独占下一行，适合按相同局部布局收口。

## Decision
refactor_existing

## Reason
这次改动只统一 `/academy/overview` 多个 panel 内条目标题与 chips 的相对位置，不改变 panel 数据、状态语义或空状态。最小正确改动是在各自的 panel owner 内引入本地 `title-line` 容器，直接复用 `insight` 已验证的布局模式，而不是抽共享组件或全局样式。

## Files to modify
- .harness/reuse-decisions/academy-overview-panel-chip-title-inline.md
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue

## After implementation
- `/academy/overview?panel=trend` 的重点班级 tags 与标题同行。
- `/academy/overview?panel=review` 的复盘摘要 tags 与标题同行。
- `/academy/overview?panel=intervention` 的介入对象 meta tags 与标题同行。
- 移动端或空间不足时，tags 仍可自然换行并回落到纵向堆叠。
