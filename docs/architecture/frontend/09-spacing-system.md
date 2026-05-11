# 前端间距与主题节奏系统

> 状态：Current
> 事实源：`code/frontend/src/assets/styles/theme.css`、`code/frontend/src/style.css`、`code/frontend/src/assets/styles/*.css`、`code/frontend/package.json`
> 替代：无

## 定位

本文档只说明前端现在如何用全局 token、语义间距和共享样式壳保持页面节奏一致。

- 覆盖：`theme.css` 的间距刻度、workspace 语义变量、共享壳样式、目录页节奏和主题检查脚本。
- 不覆盖：单个页面的全部 CSS 细节。

## 当前设计

- `code/frontend/src/assets/styles/theme.css`
  - 负责：定义全局色彩、字号、间距刻度和 workspace 语义变量，例如 `--space-*`、`--space-workspace-*`、`--space-section-*`
  - 不负责：直接绑定某个具体页面结构

- `code/frontend/src/style.css`
  - 负责：目录页、路由转场、页面标题、副标题和通用列表容器的共享节奏
  - 不负责：替代各专题样式文件的局部视觉规则

- `code/frontend/src/assets/styles/workspace-shell.css`、`page-tabs.css`、`teacher-surface.css` 以及 `journal-*.css`
  - 负责：把语义间距 token 落到工作台壳、tab rail、教师工作区和 journal 系列表面样式
  - 不负责：允许页面继续写一套脱离 token 的平行间距系统

- `code/frontend/scripts/check-theme-tail.mjs`
  - 负责：阻止 `src/components`、`src/views`、`src/composables`、`src/utils` 中继续写硬编码主题尾部 token
  - 不负责：自动修正页面样式；它只做守门

## 1. 全局间距刻度

`theme.css` 当前提供的基础刻度以 4px 网格为主，保留若干半步：

- `--space-0`
- `--space-0-5`
- `--space-1`
- `--space-1-5`
- `--space-2`
- `--space-2-5`
- `--space-3`
- `--space-3-5`
- `--space-4`
- `--space-4-5`
- `--space-5`
- `--space-5-5`
- `--space-6`
- `--space-7`
- `--space-8`
- `--space-10`
- `--space-12`

使用原则：

- 新布局优先从这些 token 取值
- 不再把新的 `px` 常量散落到业务页面

## 2. 语义间距变量

`theme.css` 当前已经把一部分高频节奏上升为语义变量：

### 2.1 Workspace 壳层

- `--space-workspace-topbar-gap-x`
- `--space-workspace-topbar-gap-y`
- `--space-workspace-topbar-padding-top`
- `--space-workspace-side-padding`
- `--space-workspace-topbar-leading-gap-x`
- `--space-workspace-topbar-leading-gap-y`
- `--space-workspace-note-gap-x`
- `--space-workspace-note-gap-y`
- `--space-workspace-tabs-gap`
- `--space-workspace-tabs-offset-top`
- `--space-workspace-tab-padding-top`
- `--space-workspace-tab-padding-bottom`
- `--space-workspace-content-padding`

### 2.2 分区与分隔

- `--space-section-gap`
- `--space-section-gap-compact`
- `--space-divider-gap`
- `--space-divider-gap-compact`

这些语义变量的目的，是让页面结构改“语义层”，而不是逐页追着 `margin` 和 `padding` 改。

## 3. 当前共享绑定位置

### 3.1 `style.css`

当前负责这些共享节奏：

- `workspace-directory-section`
- `workspace-directory-loading`
- `workspace-directory-empty`
- `workspace-directory-list`
- `workspace-directory-pagination`
- 页面标题和副标题的全局 token
- route transition 位移和时长

目录页当前规则：

- `workspace-directory-section` 默认 `display: grid`
- section 自己控制 `gap`
- `list-heading` 默认不叠加额外底部 margin
- `workspace-directory-toolbar` 通过 `--workspace-directory-toolbar-gap-bottom: 0` 收口，避免重复间距

### 3.2 `workspace-shell.css`

当前绑定：

- `workspace-topbar`
- `topbar-leading`
- `top-note`
- `top-tabs`
- `top-tab`
- `content-pane`

这决定了工作台页面的上边距、tab rail 和正文 padding 都走语义 token。

### 3.3 `page-tabs.css`

当前绑定：

- `workspace-tab-heading`
- `workspace-tab-heading__title`
- `workspace-tab-copy`
- `top-tabs` 与 `top-tab` 的 journal 风格变体

### 3.4 `teacher-surface.css`

当前绑定：

- `teacher-surface-board`
- `teacher-surface-section`
- `teacher-surface-filter`
- `teacher-topbar`
- `teacher-summary`
- `teacher-actions`

说明：

- 教师工作区的主要垂直节奏已经通过 `--space-section-gap`、`--space-divider-gap` 等 token 收口
- 仍有少量半结构样式保留在该文件内，但不应再新起一套间距体系

## 4. 主题与明暗模式

`theme.css` 同时承担主题变量基线：

- 默认 `:root` 是 dark-first token
- `[data-theme='light']` 会覆盖基础色板
- `[data-brand='green' | 'cyan' | 'blue' | 'orange']` 会切主色

间距系统和主题系统当前是一起工作的：

- 页面节奏由 `--space-*` 与语义变量决定
- 颜色和表面层级由 `--color-*` 与 `--theme-*` 决定

## 5. 使用约束

必须遵守：

- 页面优先复用 `workspace-shell`、`content-pane`、`top-tabs`、`workspace-directory-*` 这些共享壳
- 新增布局优先使用 `var(--space-*)` 或现有语义变量
- 需要微调时，优先覆盖语义变量，而不是直接写更多裸值

避免继续扩散：

- 在业务组件里新增一串脱离 token 的固定 spacing 常量
- 为了解决单页问题，临时叠加一层全局类名
- 让目录 section 的 `gap` 和 toolbar 自带间距重复叠加

## 6. Guardrail

- 主题尾部硬编码检查：`code/frontend/scripts/check-theme-tail.mjs`
- 运行命令：`cd code/frontend && npm run check:theme-tail`
- 目录表格继续走主题 token：`code/frontend/src/components/common/__tests__/WorkspaceDataTable.test.ts`
- 布局壳和 full-bleed 内容壳：`code/frontend/src/components/layout/__tests__/AppLayout.test.ts`

