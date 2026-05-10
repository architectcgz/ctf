# 2026-05-10 workspace page header 统一 review

## Review 结论

通过。首屏标题区已收口到共享 `workspace-page-header`，本次 touched surface 未继续保留局部 `workspace-hero` 布局块。

## 检查点

- 共享样式包含双列布局、下方分隔线、窄屏单列降级和可调变量。
- `/challenges`、学生目录页、教师管理页与平台资源类页面使用 `<header class="workspace-page-header">`。
- 页面局部只保留标题宽度、说明宽度、右侧操作区等差异样式。
- 测试新增结构约束，避免后续资源页重新引入 `<section class="workspace-hero">`。

## 残余风险

- 教师 dashboard 中仍有 `workspace-hero teacher-dashboard-hero tab-panel`，它承担 tab 面板显示状态，不属于本次首屏页头结构；后续如果要重命名该面板类，应单独处理，避免和 tab panel 行为混在同一次改动里。
