# 错误页按钮应使用通用 ui-btn

## 问题描述

错误页 `ErrorStatusShell` 原本用 `.error-status-action-primary / .error-status-action-secondary` 自己实现按钮边框、背景和 hover。夜间模式下 hover 边框过亮时，只能继续在错误页局部调色，说明它已经偏离通用按钮体系。

## 原因分析

`workspace-shell.css` 已经提供 `ui-btn` 作为通用按钮原语。错误页动作按钮属于普通恢复动作，不需要独立按钮体系。页面私有按钮类会让 primary / secondary / hover / focus / dark mode 的规则分散，后续其它错误页或空状态按钮容易重复出现同类问题。

## 解决方案

- 错误页动作按钮改为 `ui-btn ui-btn--primary` / `ui-btn ui-btn--secondary`。
- 错误页只通过 `--ui-btn-*` 变量设置尺寸和夜间 hover 边框强度。
- 禁止继续新增 `xxx-action-primary / xxx-action-secondary` 这类页面私有按钮体系。
- 发现其它页面仍有局部按钮体系时，先判断是否能直接切到 `ui-btn` 或已有共享变体；不能切时再补共享变体，而不是在页面内继续复制按钮状态。

## 收获

按钮问题不要只按视觉节点修。只要按钮承担的是通用动作，就应先回到共享按钮原语；页面只负责语义变量，不负责重新实现按钮状态机。
