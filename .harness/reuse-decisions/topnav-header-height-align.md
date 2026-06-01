# Reuse Decision

## Change type
layout visual alignment / shared top navigation

## Existing code searched
- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/sidebar/SidebarPanelHeader.vue`
- `code/frontend/src/shared/ui/layout/sidebar/sidebarShell.css`

## Similar implementations found
- `TopNav.vue` 和 `SidebarPanelHeader.vue` 的主体内容高度都已经是 `h-16`，实际结构高度并没有分叉。
- 视觉差异主要来自 `topNavShell.css` 里 `.topnav-shell--admin` 额外叠加的底部阴影，以及比侧栏更轻的底边颜色。
- 左侧 `STUDENT SPACE` 对照区只使用简单的 `border-bottom`，没有额外阴影参与分隔。

## Decision
refactor_existing

## Reason
这次问题不是某个页面 header 单独超高，而是共享顶部导航和共享侧栏品牌条在“分隔方式”上的默认规则不一致。

最小正确改动是直接修改共享 `topNavShell.css`：

- 保留 `TopNav` 现有的 `h-16` 布局高度不动。
- 去掉 `.topnav-shell--admin` 的底部投影，避免底边往下发灰，造成 header 比左侧更高的视觉错觉。
- 把顶部导航底部分隔线强度调整到和侧栏 header 同一档，统一成“简单细线分隔”。

本轮不做：

- 不改 breadcrumb、用户卡片、按钮的内部尺寸。
- 不改单页 route 或 scoreboard 私有样式。
- 不重做顶部导航背景层次，只修正共享壳的分隔视觉。

## Files to modify
- `.harness/reuse-decisions/topnav-header-height-align.md`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`

## After implementation
- 顶部导航的视觉高度会和左侧 `STUDENT SPACE` 更一致。
- 顶部与正文之间只靠简单边线分隔，不再有额外阴影把 header 视觉拉高。
- 所有复用共享 `TopNav` 的页面都会自动吃到这次对齐结果。
