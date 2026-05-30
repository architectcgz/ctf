> 状态：Current
> 事实源：`TopNav.vue` 当前 owner、layout shell bridge、topnav raw-source / 主题 token 护栏测试
> 替代：无

# TopNav Owner Cleanup Plan

## 目标

- 把 `TopNav.vue` 从“route/theme/session bridge + 本地 view state + breadcrumb detail 推导 + shell CSS”收口成明确的 layout shell owner
- 为 topnav 补齐本地 view-state composable 与独立 shell CSS 文件
- 保持顶部导航的 brand picker、breadcrumb、notification trigger、user card 和 logout 语义不变

## 非目标

- 本轮不调整 `useTheme()`、`useLayoutSessionActionsBridge()`、`useWorkspaceShellNavigation()`、`useBackofficeBreadcrumbDetail()` 的 public contract
- 本轮不修改 `TopNavBrandPicker.vue`、`TopNavBreadcrumbs.vue`、`TopNavNotificationTrigger.vue` 等子组件接口
- 本轮不改变 breadcrumb 的业务文案规则或支持路径

## 输入依据

- `code/frontend/src/components/layout/TopNav.vue`
- `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/composables/useBackofficeBreadcrumbDetail.ts`
- `code/frontend/src/composables/useWorkspaceShellNavigation.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 当前主问题不在顶栏子组件数量，而在 `TopNav.vue` 仍同时持有 mobile 判定、brand picker lifecycle、breadcrumb detail 推导和整段 shell CSS。
- 这些状态都属于 topnav 本地 view-model，不应继续堆在父 SFC，也不需要回灌到更底层通用 composable；更适合落到 `TopNav` 局部 composable。
- shell 样式已足够稳定，继续留在父 SFC 只会让 `TopNav.vue` 长期维持非 owner 的大体量。

## 设计边界

### `TopNav.vue` 本轮继续负责

- `sidebarCollapsed / notificationStatus` props contract
- `toggleSidebar` emits contract
- notification trigger slot 组合
- 顶栏模板组合与子组件编排

### `useTopNavViewState.ts` 本轮负责

- `isMobile`
- `brandPickerRef`
- `brandPickerOpen`
- `toggleBrandPicker / closeBrandPicker / selectBrand`
- `currentBrandLabel`
- `userDisplayName / userInitial / roleCaption`
- `backofficeBreadcrumb`
- `navigateBreadcrumb`
- detail label 推导
- `resize` / pointerdown / keydown lifecycle cleanup

### `topNavShell.css` 本轮负责

- `.topnav-shell*`
- `.topnav-inner-shell*`
- `.topnav-tool-cluster*`
- `.topnav-brand-picker*`
- `.topnav-icon-button*`
- `.topnav-breadcrumb*`
- `.topnav-user-*`
- light / dark theme token 与响应式 shell 样式

## 任务切片

### Slice 1：抽出 topnav 本地 view-state owner

- 目标：
  - 新增 `useTopNavViewState.ts`
  - 把 mobile 判定、brand picker lifecycle、breadcrumb detail 推导和展示派生迁出父组件
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts`
- Review focus：
  - route/theme/session bridge 是否仍保持清晰 owner
  - breadcrumb detail label 语义和 brand picker close 行为是否没有漂移

### Slice 2：抽出 topnav shell CSS

- 目标：
  - 新增 `topNavShell.css`
  - `TopNav.vue` 改为只保留模板组合与 imports
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - raw-source / theme-token 护栏是否仍覆盖 shell CSS
  - light / dark token、collapsed 宽度补偿和移动端 spacing 是否没有漂移

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上“route/theme/session bridge + 本地 view-state + breadcrumb detail 推导 + shell CSS”混写是否真的收口
  - 测试是否已经切到聚合源码视角

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `TopNav` 涉及的 breadcrumb 详情标题规则较多；本轮会保留现有语义，但后续如果再增长，更适合考虑单独的 breadcrumb detail support，而不是再回灌到父组件。
- 本轮完成后，layout P2 在三大布局组件上的大组件 owner 收口基本完成，剩余问题更多会转成局部 contract 或表现层优化，不再是同级别的大文件混写。
