# 业务专用抽屉不应继承通用弹窗视觉模板

## Status

not-impl

## 问题描述

通知中心右侧抽屉曾复用 `SlideOverDrawer` / `ModalTemplateShell` 这类通用弹窗模板。后续为了 1:1 复刻通知中心设计，不断通过 `panelStyle`、`overlayClass`、`:deep()` 和局部变量修正高度、背景、footer、浅深色背景，导致问题排查被拉长。

最终问题不是单个 CSS 值错误，而是抽象边界错误：通知中心属于导航级系统抽屉，具有自己的宽度、高度、footer 布局、遮罩和视觉节奏，不适合继承通用弹窗/抽屉视觉模板。

## 原因分析

`ModalTemplateShell` 当前同时承担了两类职责：

- 行为能力：Teleport、Escape 关闭、点击遮罩关闭、body scroll lock、`aria-modal`。
- 视觉与布局注入：`panelClass`、`overlayClass`、`panelStyle`、`panelTag`、`frosted` 和一组布局变量。

当业务组件需要明显不同的视觉结构时，继续继承这个模板会把样式所有权拆散到父组件、通用 shell、全局 CSS 和业务组件内部。组件表面复用，实际变成对通用模板打补丁，降低灵活性，也增加调试成本。

## 解决方案

后续处理前端弹窗/抽屉时默认遵守：

- 只复用通用行为，不强行复用通用视觉。
- `ModalTemplateShell` 适合普通确认弹窗、标准居中弹窗、标准表单抽屉等结构稳定的场景。
- 通知中心、导航系统层、复杂业务面板、需要 1:1 特定设计稿的抽屉，应拥有自己的 DOM 和 CSS。
- 如需复用能力，优先抽出 headless 行为层或 composable，例如 Teleport、Escape、scroll lock、backdrop close，而不是继承 `ModalTemplateShell` 的视觉结构。
- 不再用 `panelStyle` / `overlayClass` 作为业务专用抽屉的主要适配方式；出现这类需求时，应先判断是不是抽象边界选错。

## 收获

行为复用和视觉复用必须分开判断。业务专用抽屉如果已经有独立设计语言，就应该直接拥有结构和样式；通用模板只适合稳定、可预测的标准弹窗形态。
