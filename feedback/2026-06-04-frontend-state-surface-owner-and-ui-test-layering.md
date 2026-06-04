# 前端状态 surface owner 与 UI 测试分层需要收口

## 问题描述

训练记录页曾出现刷新加载时整块区域变成玻璃屏、加载完成后又恢复另一套样式的问题。排查后确认表象是 CSS surface 影响范围过大，根因是同一块业务区域的 loading 与 loaded 分别由不同组件承担：

- `StudentInsightTimelineSection` 渲染独立 loading 壳。
- `TrainingTimelineContent` 渲染 loaded / empty 内容结构。
- 视觉 surface、骨架、内容列表和指标卡没有同一个 owner。

这暴露出一个更通用的前端分层风险：页面壳、业务内容、状态 surface、异步状态 owner 没有始终被严格区分。复杂页面里如果继续让外层页面为 loading 单独换一整套 DOM，很容易再次出现“加载态像另一个页面、完成态又是另一套结构”的漂移。

同时，本次修复也暴露出 UI 测试侧的另一个问题：项目里存在较多 `?raw` 源码字符串断言。这类测试能有效防止关键 class 或边界回流，但如果数量过多、粒度过细，会把实现细节锁死，并让真实用户行为被“源码是否包含某段字符串”替代。

## 原因分析

- 复杂页面为了快速补 loading skeleton，容易在父级先写一套临时壳，后续 loaded 内容再进入另一个组件，形成双 owner。
- `workspace-glass-region`、`workspace-glass-metric-surface` 这类高影响视觉 class 缺少足够明确的挂载范围约束，挂到大区域时会放大结构问题。
- `SectionCard`、`StudentInsightStateSurface`、`StudentInsightLoadingSurface` 同时被用来承担 section 外壳、状态切换和视觉 surface，职责边界容易被调用方混用。
- `entities/*` 中部分 UI 组件已经接近 workspace 内容组件，既展示业务对象，又承担 header、分页、metric grid、页面 copy 等较重结构，需要持续检查是否越过实体展示边界。
- `?raw` 断言成本低，适合做架构护栏，但长期大量使用会让测试偏向“实现文本一致”，而不是验证组件状态、可见 UI 和交互行为。

## 解决方案

- 一个视觉区域只允许一个状态 owner：loading / empty / error / loaded 应由同一内容组件或同一 `StateSurface` 承载，外层页面只传入数据和状态。
- 页面壳与内容组件分离：页面壳负责 route、query、数据加载、权限和错误策略；内容组件负责展示结构和局部交互。
- 高影响 surface class 必须有明确挂载范围：整页、section、list shell、metric card、row surface 不混用，必要时用命名和测试把范围固定下来。
- `SectionCard` 只作为 section 外壳，`StateSurface` 只作为区域内状态切换，`LoadingSurface` 不再作为默认完整页面骨架壳。
- `entities/*` 只放稳定业务对象展示；如果组件开始承担页面 header、分页策略、panel rhythm 或 route/workflow 语义，应拆成实体展示组件与 feature/widget 壳。
- UI 测试分层瘦身：
  - 保留少量 `?raw` 架构边界测试，用来防止错误依赖、错误 owner、危险 surface class 回流。
  - 关键 loading / empty / loaded 用组件渲染测试验证真实 DOM 和可见状态。
  - 对容易出现“整块玻璃屏”“布局壳漂移”的页面，优先考虑少量视觉回归或截图测试，而不是堆大量源码字符串断言。
  - 删除或合并只重复检查同一 class 组合、且不能代表用户行为或架构边界的 UI 断言。

## 收获

- CSS 问题如果反复修不好，通常要回到 DOM owner 和状态 owner，而不是继续调选择器。
- 共享组件越通用，越需要明确“它不负责什么”；否则调用方会把布局、状态和视觉 surface 全塞进去。
- UI 测试不应只追求数量。测试组合应覆盖架构边界、状态行为和少量高风险视觉回归，而不是让源码字符串断言成为主要安全感来源。

## 沉淀状态

- 状态：仅项目保留
- Owner：前端页面重构 / 测试整理后续任务
- 链接：
  - `/home/azhi/workspace/projects/ctf/.harness/reuse-decisions/training-records-glass-scope.md`
  - `/home/azhi/workspace/projects/ctf/code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue`
  - `/home/azhi/workspace/projects/ctf/code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue`
