# 2026-05-10 workspace page header 统一方案

## 背景

部分工作区页面使用学生端 `/challenges` 的顶部 header 结构，另一些平台资源页使用 `workspace-hero` 局部结构。两者承担同一职责：页面首屏标题、说明和右侧操作区，但分隔线、响应式和间距由不同局部样式维护，容易出现页面漏掉下方分隔线。

## 范围

- 新增共享 `.workspace-page-header`，统一首屏页头的双列布局、下方分隔线和窄屏单列降级。
- 将同类资源页、学生目录页、教师管理页和平台入口页首屏标题区从 `workspace-hero` 或局部 header 样式迁到 `workspace-page-header`。
- 保留各页面自己的标题宽度、说明宽度、右侧操作区和状态摘要样式。
- 用测试约束后续页面不得继续为首屏页头重复声明 `workspace-hero` 布局和分隔线。

## 判定标准

- `/challenges`、学生目录页、教师管理页和平台资源类页面的首屏标题区都包含 `workspace-page-header`。
- `workspace-page-header` 的分隔线在共享样式中声明。
- 已迁移页面不再出现 `<section class="workspace-hero">` 和 `.workspace-hero { ... }` 局部布局块。
- 相关前端测试和类型检查通过。
