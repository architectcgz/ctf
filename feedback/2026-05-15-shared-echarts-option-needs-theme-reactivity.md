# 共享 ECharts option 需要显式依赖 theme ref

## 问题描述

学生仪表盘概览里的雷达图在亮暗切换后，canvas 上的轴标签颜色会停留在上一次主题的值。同一套共享图表封装里的折线图、柱状图和仪表图也有同样的风险。

根因不是 CSS 变量本身失效，而是 `option` 的 `computed` 只读了 `getComputedStyle(document.documentElement)`，没有把 `theme` ref 纳入响应式依赖。这样一来，`data-theme` 变了，图表 option 却不会重算，`vue-echarts` 继续拿旧的颜色配置。

## 原因分析

共享图表封装把“页面主题”转换成“ECharts 配置”。这一步属于运行时输入转换层，不能只依赖静态 CSS 变量读取，还必须显式订阅主题变化。

如果 `computed` 里没有 `theme.value` 这样的响应式读取，主题切换后即使 DOM 上的 `data-theme` 已经变了，图表 option 仍然会保持第一次计算出来的颜色。

## 解决方案

- 在 `LineChart.vue`、`BarChart.vue`、`GaugeChart.vue`、`RadarChart.vue` 的 option `computed` 中显式读取 `useTheme().theme.value`。
- 继续通过 `getComputedStyle` 读取 CSS 变量，但让 `computed` 在主题切换时自动重算。
- 给雷达图补回归测试，验证 `data-theme` 切换后 `axisName.color` 会跟着变化。
- 后续新增共享 ECharts 封装时，先检查 option 生成逻辑是否已经订阅主题 ref，而不是只确认 CSS 变量能读到值。

## 收获

主题切换是图表渲染输入的一部分，不只是页面外壳的视觉细节。只要 option 依赖 CSS 变量，就要确认生成 option 的那层真的能响应主题变化。

## 沉淀状态

- 状态：仅项目保留
- Owner：CTF 前端共享图表封装与 `EChartsMountGate` 回归测试
- 链接：`code/frontend/src/components/charts/LineChart.vue`、`code/frontend/src/components/charts/BarChart.vue`、`code/frontend/src/components/charts/GaugeChart.vue`、`code/frontend/src/components/charts/RadarChart.vue`、`code/frontend/src/components/charts/__tests__/EChartsMountGate.test.ts`
