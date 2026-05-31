# shared 共享层 owner 收口计划

> 状态：Current
> 事实源：当前 `components/{common,layout,navigation,charts,errors}` consumer 扫描、前端架构守卫与 `widgets/layout-shell` 现状
> 替代：无

## 目标

- 把当前长期共享层从 `components/*` 收口到正式的 `shared/*` 分层。
- 明确 `shared/ui`、`shared/model`、`shared/lib` 的边界，不把非 UI 文件硬塞进 `shared/ui`。
- 收口 `layout` 反向依赖 `widgets/layout-shell` 的临时桥接，避免迁移后形成新的低层反向依赖。
- 同步更新前端架构事实源、自动组件注册入口和 raw-source / 架构守卫测试。

## 非目标

- 本轮不处理 `components/platform/*`、`components/teacher/*`、`components/contest*/*` 等业务 owner 目录。
- 本轮不改动 feature / entity / widget 的业务职责，只调整它们对共享层的依赖入口。
- 本轮不顺手重写共享组件视觉或交互，只做 owner 和分层归位。

## 方案依据

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/06-components.md`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/vite.config.ts`
- `code/frontend/src/components/{common,layout,navigation,charts,errors}/**`
- `code/frontend/src/widgets/layout-shell/**`

## 当前边界

- `components/common/*` 里同时存在可渲染组件、overlay 行为 hook、类型文件和菜单子块。
- `components/layout/*` 里同时存在布局壳、可渲染子面板、纯类型、view-state hook 和局部样式文件。
- `components/navigation/*` 里既有可渲染路由链接，也有纯导航契约类型。
- `components/charts/*` 里既有可渲染图表组件，也有 chart mount gate 与雷达图视觉 helper。
- `components/errors/ErrorStatusShell.vue` 是共享错误页表面，但目前仍留在 `components/errors`。
- `layout` 共享壳当前通过 `widgets/layout-shell/*` 获取通知与登出 bridge，形成共享层反向依赖 widget 的中间态。

## shared 分层约定

- `shared/ui/*`
  - 负责：可直接渲染的共享组件、布局壳、错误页表面、overlay 模板、chart 组件、相关样式壳。
  - 不负责：路由契约类型、纯 helper、view-state composition、bridge hook。
- `shared/model/*`
  - 负责：共享 UI 自己的稳定 view-state / local composition / 类型契约。
  - 不负责：跨 feature 业务状态机，或需要长期留在 feature / widget 的流程 owner。
- `shared/lib/*`
  - 负责：与 UI 无关的纯 helper、导航 target 类型、chart visual helper、mount gate 等。
  - 不负责：可渲染表面和依赖 feature/widget 的 bridge。

## 任务切片

### Slice 1：建立 `shared/*` 目录与架构边界

- 新增 `code/frontend/src/shared/ui`、`shared/model`、`shared/lib`
- 更新前端架构事实源与 `frontend-architecture-policy.json`
- 让架构守卫把 `shared/*` 识别为低层共享入口，而不是继续只认 `components/common`

### Slice 2：迁移长期共享表面与非 UI helper

- 迁移可渲染组件与样式：
  - `components/common/*`、`components/common/modal-templates/*` 中的 UI 表面 -> `shared/ui/*`
  - `components/layout/*`、`sidebar/*`、`topnav/*`、`notification-drawer/*` 中的 UI 表面与样式 -> `shared/ui/layout/*`
  - `components/navigation/AppRouteLink.vue`、`AppRouteRedirect.vue` -> `shared/ui/navigation/*`
  - `components/charts/*.vue` -> `shared/ui/charts/*`
  - `components/errors/ErrorStatusShell.vue` -> `shared/ui/errors/*`
- 迁移非 UI 文件：
  - `routeTarget.ts`、`echartsMountGate.ts`、`radarVisuals.ts`、`useOverlayBehavior.ts`
  - `instancePanel.types.ts`
  - `layout/**/types.ts`、`use*ViewState.ts`

### Slice 3：收口 layout bridge 与全量 import / test 引用

- 把 `widgets/layout-shell/*` 中仅服务共享布局壳的 bridge 移到 `shared/model/layout/*` 或等价共享落点
- 更新 `AppLayout`、`TopNav`、`NotificationDrawer` 对 bridge 的引用
- 同步更新 features / pages / widgets / entities / tests / `components.d.ts`

## 预期改动文件

- `code/frontend/src/shared/**`
- `code/frontend/src/widgets/layout-shell/**`
- `code/frontend/src/components.d.ts`
- `code/frontend/vite.config.ts`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/src/__tests__/frontendArchitecturePolicy.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/pages/**`
- `code/frontend/src/features/**`
- `code/frontend/src/entities/**`
- `code/frontend/src/widgets/**`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/06-components.md`

## 验证

- `bash scripts/check-task-intake.sh --reuse-decision shared-ui-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/components/common/__tests__/ModalTemplates.test.ts src/components/common/__tests__/WorkspaceDataTable.test.ts src/components/common/__tests__/WorkspaceDirectoryToolbar.test.ts src/components/common/__tests__/WorkspaceDirectoryPagination.test.ts src/components/common/__tests__/AppEmptySurface.test.ts src/components/common/__tests__/AppToast.test.ts src/components/common/__tests__/InstancePanel.test.ts src/components/layout/__tests__/AppLayout.test.ts src/components/layout/__tests__/TopNav.test.ts src/components/layout/__tests__/Sidebar.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts src/components/layout/__tests__/BackofficeSubNav.test.ts src/components/charts/__tests__/EChartsMountGate.test.ts src/components/charts/__tests__/GridConfig.test.ts`
- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `git diff --check`

## Review 关注点

- `shared/ui` 是否只包含可渲染表面，没有把 `routeTarget`、`echartsMountGate`、`useOverlayBehavior` 这类非 UI 文件硬塞进去。
- `shared/model` / `shared/lib` 是否没有重新引入 `widgets`、`features`、`stores`、`router` 的低层反向依赖。
- `layout` 共享壳是否已经摆脱对 `widgets/layout-shell` 的反向依赖。
- 原有 raw-source 测试是否都已改到新 owner，而不是残留旧 `components/*` 路径。

## 回退

- `shared/*` 目录与架构守卫改动可独立回退。
- 如果 layout bridge 收口范围超出当前切片，可先回退该桥接迁移并重切任务，但不能在最终结果里保留 `shared -> widgets` 反向依赖。
