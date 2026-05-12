# Reuse Decision

## Change type
component / layout

## Existing code searched
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/Sidebar.vue
- code/frontend/src/components/layout/AppLayout.vue
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/components/layout/__tests__/AppLayout.test.ts
- code/frontend/src/assets/styles/theme.css

## Similar implementations found
- `AppLayout.vue` 已经是 `sidebarCollapsed` 的唯一状态 owner，并把这个状态传给 `TopNav` 和 `Sidebar`。
- `Sidebar.vue` 已经在桌面折叠态通过 `collapsed` 切换宽度，不需要再新增新的折叠状态源。
- `TopNav.vue` 现有 `topnav-inner-shell` 已经承担 header 内容宽度控制，适合在这个壳层上追加 collapsed modifier，而不是拆出第二套 header 结构。

## Decision
refactor_existing

## Reason
这次不是新增 header 组件，而是让现有全局 topnav 在桌面侧栏折叠后更好地利用释放出来的横向空间。直接复用 `AppLayout` 传下来的 `sidebarCollapsed` 和 `TopNav` 现有壳层宽度控制，能把左侧面包屑区和右侧操作区一起向两侧外扩，改动最小，也不会引入新的状态分叉。

## Files to modify
- .harness/reuse-decisions/topnav-collapsed-header-balance.md
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
