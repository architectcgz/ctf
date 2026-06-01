# Reuse Decision

## Change type
layout line alignment / shared top navigation height owner

## Existing code searched
- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/sidebar/SidebarDesktopPanel.vue`
- `code/frontend/src/shared/ui/layout/sidebar/SidebarPanelHeader.vue`
- `code/frontend/src/shared/ui/layout/sidebar/sidebarShell.css`

## Similar implementations found
- 左侧 `SidebarPanelHeader.vue` 直接把 `h-16` 和 `border-bottom` 挂在同一个 header 盒子上。
- 右侧 `TopNav.vue` 目前是外层 `header` 负责 `border-bottom`，内层 `.topnav-inner` 才是 `h-16`，两层 owner 分裂。
- 在项目的全局 `box-sizing: border-box` 前提下，左侧总高是 64px 含边线，右侧则会变成 64px 内容 + 1px 边线，所以分隔线天然低 1px。

## Decision
refactor_existing

## Reason
这次问题的 root cause 是共享顶栏高度 owner 没收口，而不是颜色、阴影或控件尺寸。

最小正确改动是：

- 保持 `TopNav` 的视觉样式不动。
- 让外层 `header` 自己成为 `h-16` owner。
- 把内层容器从 `h-16` 改成 `h-full`，这样底部分隔线会和左侧 sidebar header 落在同一条线上。

本轮不做：

- 不继续调顶栏按钮和用户卡片尺寸。
- 不改 sidebar 样式。
- 不改 workspace shell 或页面 tabs。

## Files to modify
- `.harness/reuse-decisions/topnav-divider-offset-align.md`
- `code/frontend/src/shared/ui/layout/TopNav.vue`

## After implementation
- 顶部 `header` 下方分隔线会和左侧 `STUDENT SPACE` 分隔线处于同一高度。
- 修复点收口在共享 `TopNav` owner，而不是局部页面补丁。
- 所有复用共享顶栏的页面都会同步得到同样的对齐结果。
