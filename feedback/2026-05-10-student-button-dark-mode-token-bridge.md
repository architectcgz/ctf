# Student 按钮夜间模式需要共享变体

## 问题描述

学生仪表盘难度面板里的行动按钮在区分主推档位和普通档位时，普通档位按钮最初挂在 `journal-btn-primary` 上，再通过页面局部规则改边框、背景和文字色。这样会让按钮语义和主题实现混在页面里，hover、focus 或夜间模式容易继承主按钮行为，导致按钮看起来没有完整适配夜间模式。

## 原因分析

这类按钮属于共享 `journal-soft-surface` 按钮体系的一部分。页面应该只选择按钮语义，例如 primary、secondary、outline；按钮的边框、背景、文字、hover、focus 和夜间模式应由共享样式层统一负责。若页面自己实现一套 secondary 配色，后续每个页面都会重复修暗色模式。

## 解决方案

- 在 `journal-soft-surfaces.css` 中新增共享 `journal-btn-secondary` 变体。
- 页面按钮在 `journal-btn-primary` 与 `journal-btn-secondary` 之间选择，不再保留页面私有 secondary 配色类。
- primary / secondary / outline 按钮在 light / dark 下都必须保留可见边框；边框颜色要以 `--journal-control-border` 为参照，不把 `transparent` 作为唯一边界参照。
- 夜间模式 hover 边框只做轻微增强，primary 保持克制，secondary / outline 更低强度，避免鼠标经过时出现刺眼高亮。
- 禁止用页面私有 dark selector 单独修按钮；优先让共享按钮规则和主题变量统一生效。
- 对这类问题补 raw-source 测试，确保共享样式声明 secondary 变体，并禁止难度面板继续保留页面私有按钮配色。
- 架构事实源同步记录在 `docs/architecture/frontend/06-components.md`。

## 收获

按钮的“业务状态”和“主题实现”要分开：页面负责选择共享按钮变体，实际边框、背景、hover、focus 和夜间模式由共享按钮规则执行。
