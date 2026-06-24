# 不要用 embedded 切换布局 owner

## 问题描述

学生 dashboard 与教师学员分析页都出现过同一类实现：组件暴露一个 `embedded` 布尔开关，同时切换完整页壳、section 节奏、padding、divider 和内容区域外框。

这类写法表面上是在做“嵌入态”，实际是在让一个组件同时承担两套布局 owner 语义。时间一长，调用方和组件内部都会开始堆条件分支，后续任何样式或结构调整都需要双份判断。

## 原因分析

- 初始实现把“同一份内容可能出现在不同页面”简化成了一个布尔开关。
- `embedded` 没有被限制在局部视觉差异，而是蔓延到了 page shell / section shell 级别。
- 一旦 header、summary、divider、list shell、content padding 都跟着切，组件职责就已经混合。

## 解决方案

- 共享内容单独抽成内容 owner，只负责稳定的 header、summary、body、list、pagination、事件分组等。
- 每个调用场景各自持有自己的壳：
  - 完整页壳
  - tab panel 内嵌壳
  - section 壳
- `embedded` 只允许保留在“局部视觉嵌入态”且不会改变 owner 语义的场景；如果它开始切换 page shell、section shell、内容 root 或主要 spacing contract，就必须回到显式拆壳。

## 收获

- “一个组件服务多个入口”不等于“一个组件自己切多套布局 owner”。
- 判断标准不该是有没有复用，而该是 owner 是否单一。
- 结构清楚之后，测试护栏也会更稳定：内容断言盯共享内容组件，壳层断言盯各自调用方。

## 沉淀状态

- 状态：archived
- Owner：`ctf/AGENTS.md`
- 链接：
  - `/home/azhi/workspace/projects/ctf/AGENTS.md`
  - `/home/azhi/workspace/projects/ctf/docs/plan/archive/impl-plan/2026-06/2026-06-04-embedded-layout-owner-retirement-plan.md`
  - `/home/azhi/workspace/projects/ctf/.harness/reuse-decisions/embedded-layout-owner-retirement.md`
