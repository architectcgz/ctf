# workspace-grid 只用于真实布局

## 问题描述

部分单列页面使用了 `workspace-shell > workspace-grid > content-pane`，但共享样式里的 `workspace-grid` 只声明单列 `grid-template-columns: 1fr`。

这让 `workspace-grid` 和 `content-pane` 的职责重叠：`content-pane` 已经负责内容区域的 padding、起始间距和最小宽度，单列 `workspace-grid` 没有额外布局价值，反而迫使共享间距规则同时兼容 `workspace-shell > content-pane` 与 `workspace-shell > workspace-grid > content-pane` 两种结构。

## 原因分析

早期页面为了预留复杂布局，把所有工作区都套进 `workspace-grid`。后续大部分列表、资源、仪表盘页面已经稳定为单列内容，`workspace-grid` 退化成无意义包裹，但测试仍把它当作默认骨架的一部分。

结果是调整非 top-tabs 内容起始间距时，需要为退化结构补额外选择器，说明结构边界已经不清晰。

## 解决方案

- `content-pane` 是普通页面内容区域的唯一 owner，负责 padding、起始间距和内容主轴节奏。
- `workspace-grid` 只在存在真实布局职责时使用，例如主区/侧栏、多列响应式布局或需要在 grid 层控制区域关系的页面。
- 单列列表页、资源页、仪表盘页不要包 `workspace-grid`。
- 如果页面只是 `workspace-shell > workspace-grid > main.content-pane`，应直接改成 `workspace-shell > main.content-pane`。
- 共享样式不再为退化的 `workspace-grid > content-pane` 继续补兼容选择器；发现需要补时，优先回到结构层移除无意义包裹。

## Good Taste

- 结构元素必须有可说明的布局职责。
- 普通单列页面直接使用 `content-pane`，让间距规则只有一个 owner。
- 保留 `ChallengeDetail` 这类真实多列布局中的 `workspace-grid`。

## Bad Taste

- 为“可能以后会多列”给所有页面套 `workspace-grid`。
- 因为多了一层退化 DOM，再在共享 CSS 里追加兼容选择器。
- 把 `workspace-grid` 当作默认壳层骨架，而不是布局能力。

## 收获

共享样式的选择器复杂度通常是在提示结构职责不清。与其继续为每种 DOM 嵌套补选择器，不如先判断这层 DOM 是否还有真实职责。
