# Reuse Decision

## Change type
dark theme surface alignment / shared workspace chrome

## Existing code searched
- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/TopNavUserCard.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/sidebar/sidebarShell.css`
- `code/frontend/src/assets/styles/theme.css`
- `code/frontend/src/assets/styles/journal-user-shell.css`

## Similar implementations found
- `topNavShell.css` 的 `.topnav-shell` 基础层已经有一套可用于 dark theme 的低对比 surface token。
- 当前 `.topnav-shell--workspace` 把偏亮的白色混合渐变直接写成无条件背景，导致 dark theme 也吃到 light surface。
- `sidebarShell.css` 里的 `backoffice-sidebar` 和 `topNavShell.css` 的 `.topnav-shell` 目前各自维护一套 workspace 壳层背景，dark token 与背景公式长期漂移。
- `journal-user-shell.css` 在 dark theme 下会把学生页面 surface 压回暗面，因此共享顶栏也应跟随 sidebar 这套暗色 surface 系统，而不是继续使用单独推导的亮色变体。
- `theme.css` 已经是全局颜色 token owner，适合承接共享 workspace frame 的 surface / line / background token。

## Decision
refactor_existing

## Reason
这次问题不是单个选择器数值偏了，而是共享 workspace 顶栏和侧栏分别维护背景公式，dark theme 下长期漂移，导致顶栏继续比侧栏更亮。

最小正确改动是：

- 在 `theme.css` 新增共享 `workspace frame` token，统一定义 surface / line / text / background。
- 让 `TopNav` 和 `Sidebar` 都直接消费这套共享 token，不再各自推导 dark surface。
- 保留各自结构性边框、阴影和控件层级，但背景 owner 收口成单点。

本轮不做：

- 不改 breadcrumb、按钮和用户卡片尺寸。
- 不改正文 shell 的 hero / page 背景系统。
- 不改路由与角色逻辑。

## Files to modify
- `.harness/reuse-decisions/topnav-dark-surface-restore.md`
- `code/frontend/src/assets/styles/theme.css`
- `code/frontend/src/shared/ui/layout/sidebar/sidebarShell.css`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`

## After implementation
- 顶栏与侧栏会吃同一个 workspace frame 背景 owner，不再各自维护两套配方。
- dark theme 下顶栏不会再单独发白，视觉上会回到和左侧侧栏一致的暗面体系。
- 后续如果再调共享壳层背景，只需要改一处 token。
