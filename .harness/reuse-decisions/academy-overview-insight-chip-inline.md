# Reuse Decision

## Change type
page / component

## Existing code searched
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue
- code/frontend/src/features/teacher/dashboard

## Similar implementations found
- `TeacherDashboardStudentInsightPanel.vue` 已经是 `/academy/overview?panel=insight` 学生洞察列表的唯一 UI owner。
- 当前 chips 由 `student-insight-row__chips` 独立占一行，没有共享布局层需要复用或扩展。

## Decision
refactor_existing

## Reason
这次只调整学生洞察条目里 detail 和 chips 的相对布局，不改变条目结构、tone、title、status 或数据构建。最小正确改动是在现有 panel owner 内增加一层 detail-line 布局容器，让文案占主列、chips 跟在右侧，并在窄屏下自然回落为换行布局。

## Files to modify
- .harness/reuse-decisions/academy-overview-insight-chip-inline.md
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue

## After implementation
- `/academy/overview?panel=insight` 的条目 chips 会显示在 detail 文案右侧。
- 移动端或空间不足时，chips 允许换到下一行，不影响现有列表结构和状态样式。
