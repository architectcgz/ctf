> 状态：Current
> 事实源：`Sidebar.vue` 当前 owner、`useWorkspaceShellNavigation()` layout shell bridge、侧栏 raw-source / 主题 token 护栏测试
> 替代：无

# Sidebar Owner Cleanup Plan

## 目标

- 把 `Sidebar.vue` 从“workspace shell bridge + nav view-model + navigate workflow + shell CSS”收口成明确的 layout shell owner
- 为侧栏补齐本地 nav view state composable 与独立 shell CSS 文件
- 保持 student / academy / platform 三条 workspace 导航的现有可见模块、active 判定和移动端关闭语义不变

## 非目标

- 本轮不调整 `useWorkspaceShellNavigation()` 的 contract
- 本轮不改 `backofficeNavigation` 配置、角色权限或模块命名
- 本轮不继续拆 `SidebarDesktopPanel.vue`、`SidebarMobilePanel.vue`、`SidebarNavTree.vue`

## 输入依据

- `code/frontend/src/components/layout/Sidebar.vue`
- `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/composables/useWorkspaceShellNavigation.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `Sidebar.vue` 的 route/auth workspace 识别已经统一经由 `useWorkspaceShellNavigation()`，当前主要问题不再是 bridge，而是父组件仍同时持有 nav 展开态、active/highlight 规则、navigate 短路和整段 sidebar shell CSS。
- 这批状态都属于 layout 内本地 view-model，不应回灌到 `useWorkspaceShellNavigation()`；更适合落到 sidebar 局部 composable。
- shell 样式已经足够稳定，继续留在父 SFC 只会让 `Sidebar.vue` 维持一个非 owner 的大体量。

## 设计边界

### `Sidebar.vue` 本轮继续负责

- `collapsed / mobileOpen` props contract
- `closeMobile / toggleCollapse` emits contract
- `SidebarDesktopPanel` / `SidebarMobilePanel` 组合

### `useSidebarNavigationViewState.ts` 本轮负责

- `brandKicker`
- `backofficeItems`
- `expandedMenus`
- item / child active、expanded、highlight 判定
- `navigate()` 的 same-path / same-query 短路
- mobile navigate 后的 `closeMobile`

### `sidebarShell.css` 本轮负责

- `.sidebar-shell`
- `.backoffice-sidebar*`
- `.sidebar-brand*`
- `.sidebar-group-title`
- `.sidebar-item*`
- light / dark theme token 变量与桌面/移动侧栏壳体样式

## 任务切片

### Slice 1：抽出 sidebar 本地 nav view-model owner

- 目标：
  - 新增 `useSidebarNavigationViewState.ts`
  - 把 `expandedMenus`、item 判定和 `navigate()` 从父组件迁出
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts`
- Review focus：
  - `useWorkspaceShellNavigation()` 是否仍是唯一 workspace shell bridge
  - active / expanded / same-route 短路语义是否没有漂移

### Slice 2：抽出 sidebar shell CSS

- 目标：
  - 新增 `sidebarShell.css`
  - `Sidebar.vue` 改为只保留模板组合与 imports
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - raw-source / theme-token 护栏是否仍覆盖 sidebar shell CSS
  - light / dark token 与 desktop/mobile shell 结构是否没有漂移

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上“workspace shell bridge + nav 判定 + navigate + shell CSS”混写是否真的收口
  - 测试是否已切到聚合源码视角

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `SidebarNavTree.vue` 及桌面/移动 panel 本身仍保留展示层 props contract；本轮只收口父层 owner，不扩大到整个 sidebar API 重排。
- `TopNav.vue` 仍是 layout P2 最主要的剩余大组件；本轮完成后更适合继续转向 topnav 的 route/breadcrumb/user-card owner，而不是继续在 sidebar 内扩大范围。
