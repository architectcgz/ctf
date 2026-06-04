# Reuse Decision

## Change type
layout / shared workspace chrome / surface alignment

## Existing code searched
- `code/frontend/src/views/...`
- `code/frontend/src/components/...`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/Sidebar.vue`
- `code/frontend/src/shared/ui/layout/sidebar/SidebarDesktopPanel.vue`
- `code/frontend/src/shared/ui/layout/sidebar/SidebarPanelHeader.vue`
- `code/frontend/src/shared/ui/layout/sidebar/sidebarShell.css`
- `code/frontend/src/features/...`
- `code/frontend/src/composables/...`
- `code/frontend/src/api/...`
- `code/frontend/src/assets/styles/theme.css`
- `code/frontend/src/shared/ui/layout/__tests__/TopNav.test.ts`

## Similar implementations found
- `topNavShell.css` 已经承接共享 workspace 顶栏的背景、边线和控件 surface token。
- `sidebarShell.css` 里的 `backoffice-sidebar` 是学生 / 教师 / 管理后台共用的左侧壳层背景 owner。
- `SidebarPanelHeader.vue` 本身没有额外背景色，顶部观感实际来自 `backoffice-sidebar` 的共享背景。

## Decision
extend_existing

## Reason
这次需求不是新增学生页私有样式，而是让共享顶栏 header 的背景层和共享 sidebar 顶部保持一致。

最小正确改法是继续复用现有 workspace shell token，只把 `topnav-shell--workspace` 的背景 owner 收口到与 sidebar 顶部观感一致的 surface 层，不在 `student/dashboard` 页面单独写 override，也不复制一套新的颜色值。

## Files to modify
- `.harness/reuse-decisions/student-dashboard-topnav-sidebar-surface.md`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/__tests__/TopNav.test.ts`

## After implementation
- 学生 dashboard 顶栏会直接跟随共享 workspace 顶栏背景策略。
- 后续如果还要统一教师 / 管理后台的同一层背景，可以继续在共享 `TopNav` shell 上调整，而不是回到页面局部补丁。
