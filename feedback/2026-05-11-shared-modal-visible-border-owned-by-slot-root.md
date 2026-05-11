# 共享弹窗可见边框应由内容根承接

## 问题描述

在调整共享删除确认框时，外边框先加到了传给 `ModalTemplateShell` 的 `panel-class` 上。实际页面里用户定位到的可见节点是 slot 内容根 `section.delete-confirm-modal`，而不是子组件内部 panel，因此外边框在 `/platform/challenges` 删除弹窗中不明显。

同一轮还暴露出浅色模式下危险提示文字只使用 muted 文本色，语义和对比度都不够清楚。

## 原因分析

Vue scoped CSS 对子组件内部节点不天然生效。`panel-class` 虽然把类名传进了 `ModalTemplateShell` 内部，但父组件 scoped 样式不能把它当成本组件普通 DOM 稳定处理。

共享弹窗的视觉外壳如果依赖子组件内部节点，后续复用到其他入口时容易出现“源码有边框、实际可见节点没有边框”的偏差。

## 解决方案

- 子组件内部 panel 只负责布局尺寸，必要时使用 `:deep(.delete-confirm-modal__panel)` 明确作用范围。
- 可见背景、圆角、外边框、阴影放到 slot 内容根 `.delete-confirm-modal`。
- 危险提示使用 `--color-danger` 与主题 surface / border token 派生颜色，不使用页面私有颜色或硬编码色值。
- 静态测试检查共享删除确认框保留外边框变量、可见根节点样式和危险提示色。
- `ModalTemplateShell` 只作为模板内部结构基座，不能承接具体 surface 视觉；`ModalTemplates.test` 应阻止它新增 border / radius / shadow / overflow 这类面板视觉职责。

## `:deep` 使用边界

这次用 `:deep(.delete-confirm-modal__panel)` 修复 panel 尺寸和 overflow 是合理的，因为它只是在父组件内明确命中子组件内部的结构层。

但如果目标是“用户能看到的外边框、背景、阴影、圆角”，不应主要依赖 `:deep` 去修改 shell 内部 panel。可见 surface 应由当前组件的 slot root 承接，例如 `.delete-confirm-modal`。这样后续通过 XPath 或真实 DOM 检查时，样式归属和可见节点一致。

## 收获

共享 modal / drawer 的外观责任要和真实可见 DOM 对齐。后续如果组件通过 slot 组合，优先把可见边框、背景和阴影放在 slot 根节点；只有纯布局能力留给 shell panel。

浅色模式下的 warning / danger 文案不能只靠 muted 文本色表达风险，应使用语义 danger token 派生出更明确的文本、边框和弱背景。

## 沉淀状态

- 状态：仅项目保留
- Owner：CTF 前端共享弹窗样式、modal template 边界测试与 destructive confirm 静态测试
- 链接：`code/frontend/src/components/common/DeleteConfirmModal.vue`、`code/frontend/src/components/common/modal-templates/ModalTemplateShell.vue`、`code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`、`code/frontend/src/views/__tests__/destructiveConfirmThemeAlignment.test.ts`
