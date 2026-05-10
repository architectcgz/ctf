# workspace page header 应作为共享结构

## 问题描述

学生端 `/challenges` 使用 `challenge-topbar`，平台资源页使用 `workspace-hero`，两者本质都是页面首屏标题区。由于结构和样式分散，部分页面漏掉下方分隔线，后续也容易继续新增第三套局部 header。

## 原因分析

之前只按页面局部样式补齐分隔线，没有把“首屏标题区”识别成共享工作区原语。这样看起来改动小，但违背项目既有风格复用原则，也让测试只能检查单页结果，不能约束新页面。

## 解决方案

- 首屏标题、说明、右侧操作区统一使用 `workspace-page-header`。
- 共享样式集中在 `workspace-shell.css`，包含双列布局、下方分隔线、响应式降级和可调 CSS 变量。
- 页面只保留标题宽度、说明宽度、action rail 等真正局部的差异。
- 测试需要同时约束共享样式存在，以及资源页不再新增 `<section class="workspace-hero">` 和 `.workspace-hero { ... }` 局部布局块。

## 收获

Good taste：先识别职责相同的 UI 原语，再抽到共享层，用测试固定结构契约。

Bad taste：给每个页面单独补 `padding-bottom + border-bottom`，短期像最小改动，长期会持续产生页面漂移。
