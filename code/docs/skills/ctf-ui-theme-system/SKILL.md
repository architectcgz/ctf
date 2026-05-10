---
name: ctf-ui-theme-system
description: CTF 前端视觉系统规范（项目内副本），用于页面视觉一致性与交互语义对齐。
---

# CTF UI Theme System

## 用途
- 统一页面排版、色彩 token、列表布局、按钮层级和交互语义。
- 防止平台页、教师页、学生页在同一仓库里出现风格漂移。

## 关键约束
- 页面只显示真实产品文案，不渲染实现说明/设计意图/占位注释。
- 列表页优先平铺目录结构，避免卡片堆叠。
- 样式优先 token，不写硬编码色值，不用 `!important`。
- 先保证信息层级和可读性，再做视觉装饰。

## 说明
- 详细版规范可参考仓库根目录的 `harness/prompts/ctf-ui-theme-system-skill.md`（已迁入 harness 的唯一入口）。
