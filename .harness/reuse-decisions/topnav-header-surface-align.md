# Reuse Decision

## Change type
frontend style alignment / shared topnav surface

## Existing code searched
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/views`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/topnav/TopNavUserCard.vue`
- `code/frontend/src/shared/ui/layout/topnav/TopNavBreadcrumbs.vue`
- `/tmp/ctf-frontend-old-assets/old-assets/AppLayout-PV13NnHB.css`

## Similar implementations found
- 页面整体顶栏由 `TopNav.vue + topNavShell.css` 统一 owner，不由各个 route view 或 scoreboard feature 自己控制。
- 用户实际关注的“路径 + 学生信息 + 退出按钮”背景来自 `.topnav-shell` / `.topnav-shell--admin`，内部按钮和用户卡再复用 `--topnav-surface*` token。

## Decision
refactor_existing

## Reason
这次不是新增顶栏组件，也不是做页面级局部 override，而是把共享 topnav 的背景表面调到更接近学生工作台左侧品牌区的近白观感。

最小正确改动是：

- 继续让 `TopNav.vue` 只负责结构，不承载局部样式。
- 在 `topNavShell.css` 内统一调整 `--topnav-surface`、`--topnav-surface-subtle` 和 `topnav-shell--admin` 的背景梯度。
- 保持按钮、用户卡、面包屑继续吃同一套 `--topnav-*` token，避免只亮顶栏外壳、内部控件仍发灰。

本轮不做：

- 不改 scoreboard 页面内部 header / list shell。
- 不改单个按钮、通知抽屉或用户卡布局。
- 不扩展到侧栏、workspace 内容区或其它共享 surface。

## Files to modify
- `.harness/reuse-decisions/topnav-header-surface-align.md`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`

## After implementation
- 学生侧共享顶栏会比原先更接近近白表面，而不是整条发灰。
- 路径区、学生信息卡和退出按钮仍保持原有层级，只是跟随新的 topnav surface token 一起提亮。
- 后续如果还要继续提亮，只需要继续收口在 `topNavShell.css`，不必再去页面级别排查。
