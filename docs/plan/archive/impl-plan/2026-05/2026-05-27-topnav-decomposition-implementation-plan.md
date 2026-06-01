> 状态：Current
> 事实源：`TopNav.vue` 的 layout owner 与拆分边界
> 替代：无

# TopNav Decomposition Implementation Plan

## 目标

- 把 `code/frontend/src/components/layout/TopNav.vue` 从单文件大壳拆成“layout owner + stable topnav sections”结构。
- 保留 `TopNav.vue` 对 route breadcrumb、brand picker open state、theme/brand action、通知 trigger slot 和 logout 的 owner。

## 非目标

- 本轮不改 `AppLayout.vue` 与 `TopNav.vue` 的 props / emits contract。
- 本轮不改 `useWorkspaceShellNavigation.ts`、`useTheme.ts`、`useBackofficeBreadcrumbDetail.ts` 的行为 owner。
- 本轮不重做 topnav 视觉，不新增新的 header 功能。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/layout/TopNav.vue`
- `code/frontend/src/components/layout/AppLayout.vue`
- `code/frontend/src/components/layout/NotificationDrawer.vue`
- `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TopNav.vue` 应继续留在 `components/layout/`，因为它是全局 workspace header，不是单一 feature 的 UI。
- 当前最值得拆的部分是稳定视图区域：移动侧栏 toggle、breadcrumb、brand picker、通知 trigger、用户卡片。
- route-derived detail label、brand picker open state、theme/brand action 与通知 trigger slot 仍应留在父组件，不应被拆散成二级 owner。

## 设计边界

### `components/layout/TopNav.vue` 继续负责

- `useWorkspaceShellNavigation()` 与 breadcrumb 派生
- `useTheme()` 的 brand/theme owner
- brand picker open/close 与 outside click / Escape cleanup
- `NotificationDrawer` slot 装配
- logout owner

### `components/layout/topnav/*` 本轮负责

- 移动端侧栏 toggle 视图
- breadcrumb 视图壳
- brand picker 展示壳
- 通知 trigger 展示壳
- 用户卡片展示壳

### 本轮不负责

- 调整 `Sidebar` / `AppLayout` 的状态 owner
- 改动通知抽屉能力或主题持久化 contract
- 新增新的 header 导航行为

## 任务切片

### Slice 1：抽出稳定 topnav sections

- 目标：
  - 新增 `components/layout/topnav/*`
  - `TopNav.vue` 收回到 owner + 子视图装配
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/TopNav.test.ts`
- Review focus：
  - route / brand picker / notification owner 是否仍在父组件
  - 子组件是否只承担视图，不重新直接依赖 router/store

### Slice 2：同步 raw-source 护栏与 backlog

- 目标：
  - 更新 `TopNav.test.ts` 与 `sharedThemeTokenAdoption.test.ts` 的聚合源码断言
  - 更新 backlog 中 `TopNav.vue` 进展
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - topnav shell、breadcrumb、brand picker、通知按钮和主题 token 护栏是否仍覆盖到位

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `TopNav.vue` 拆完后，布局层这条 `P2` backlog 会更多转向更深层 owner 清理，而不再是单纯的大组件壳体问题。
