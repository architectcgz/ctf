# Reuse Decision

## Change type
dark theme surface alignment / shared app layout background

## Existing code searched
- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/shared/ui/layout/AppLayout.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/sidebar/sidebarShell.css`

## Similar implementations found
- `AppLayout.vue` 会在顶部铺一层全局 `app-layout-top-glow`，它不区分角色页和学生页。
- `sidebarShell.css` 的夜间暗面本身没有额外顶部泛光，因此颜色更稳。
- `topnavShell.css` 即使改成和 sidebar 一样的 dark surface，如果底下继续叠全局 glow，视觉上仍会偏亮。

## Decision
refactor_existing

## Reason
这次真正的剩余亮度来自共享布局底层泛光，不是 topnav 自身 token 继续算错。

最小正确改动是：

- 保留 light theme 的顶部 glow。
- 在 dark theme 下把 `app-layout-top-glow` 关闭，让顶栏与侧栏按各自暗面背景直接呈现。

本轮不做：

- 不改顶栏结构和 sidebar 结构。
- 不改单页局部样式。
- 不修改底部阴影和侧边 rail。

## Files to modify
- `.harness/reuse-decisions/app-layout-top-glow-dark-align.md`
- `code/frontend/src/shared/ui/layout/AppLayout.vue`

## After implementation
- dark theme 下顶栏不会再被全局 glow 额外提亮。
- sidebar 与 topnav 的暗面会更接近同一层级。
- light theme 的顶部氛围层保持不变。
