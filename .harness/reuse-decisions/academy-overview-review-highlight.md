# Reuse Decision

## Change type
page / component

## Existing code searched
- code/frontend/src/router/routes/teacherRoutes.ts
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue
- code/frontend/src/features/teacher/dashboard

## Similar implementations found
- `TeacherDashboardReviewPanel.vue` 已经是 `/academy/overview?panel=review` 的 review 摘要唯一 UI owner。
- review 列表项的左侧 highlight 由 `review-highlight-item` 和 `review-highlight-item--{tone}` 组合控制，不存在需要复用的独立共享样式层。

## Decision
refactor_existing

## Reason
这次改动只需要移除 review 列表项左侧的视觉高亮，不改变数据来源、列表结构或 tone 语义。最小正确改动是直接在现有 panel owner 内删除对应样式和类绑定，避免把一个局部视觉调整扩散到 page model 或共享样式。

## Files to modify
- .harness/reuse-decisions/academy-overview-review-highlight.md
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue

## After implementation
- `/academy/overview?panel=review` 的摘要列表项不再渲染左侧 highlight。
- review 面板的数据构建、空状态和 chip 展示保持不变。
