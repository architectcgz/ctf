# 共享抽屉间距应通过组件契约配置

## 问题描述

通知中心右侧抽屉列表顶部间距过大。最初修复时在业务组件里用更高优先级选择器覆盖 `SlideOverDrawer` 内部 `.modal-template-drawer__body` 的默认 padding，视觉问题被修掉了，但实现方式不够稳。

## 原因分析

通知中心是独立业务抽屉，但它复用了公共 `SlideOverDrawer` 外壳。header、body、footer 这类骨架间距属于共享组件契约，不完全属于业务组件私有样式。

如果业务组件用 `:deep` 或 `:global` 直接覆盖公共组件内部结构，会带来几个问题：

- 调用方需要知道公共组件内部 class，耦合过深。
- scoped 样式注入顺序可能影响覆盖结果。
- 其他抽屉遇到类似需求时容易复制同类穿透覆盖。
- 公共组件默认值和调用方意图没有显式表达。

## 解决方案

当右侧抽屉需要调整公共骨架的布局参数时，优先让 `SlideOverDrawer` 暴露正式配置入口，例如 body padding、header padding、footer padding 等，而不是在业务组件里覆盖内部 class。

本次已将正文内边距收敛为 `SlideOverDrawer` 的 `bodyPadding` prop，并通过 CSS 变量落到内部样式：

```vue
<SlideOverDrawer body-padding="var(--space-0)" />
```

公共组件保留默认值，具体抽屉显式声明自己的正文节奏。这样通知中心可以去掉 `:global(.notification-shell ...)` 这类临时覆盖，同时不影响其他抽屉。

## 收获

共享组件的可变布局参数应优先进入组件契约。业务组件可以定义自己的内容样式，但不应长期依赖公共组件内部 DOM/class 细节来修正骨架间距。

后续遇到类似问题时，先判断样式来源：

- 内容区域内部节奏：留在业务组件。
- header/body/footer、默认 padding、滚动容器、overlay、宽度等骨架参数：优先补共享组件 prop 或 CSS 变量。
- 临时穿透覆盖只能作为定位或短期修复，不应作为最终实现。
