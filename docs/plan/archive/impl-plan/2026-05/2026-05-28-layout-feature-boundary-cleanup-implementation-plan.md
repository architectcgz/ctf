> 状态：Current
> 事实源：`components/layout/*` 当前依赖关系、`features/notifications` / `features/auth` 现有 model 实现、`componentFeatureImportAllowlist`
> 替代：无

# Layout Feature Boundary Cleanup Implementation Plan

## 目标

- 清掉 `componentFeatureImportAllowlist` 中 layout 相关的 3 条残留：
  - `components/layout/AppLayout.vue -> @/features/notifications`
  - `components/layout/NotificationDrawer.vue -> @/features/notifications`
  - `components/layout/TopNav.vue -> @/features/auth`
- 把通知实时同步、通知抽屉 workflow 和退出登录动作收口到中立层，避免共享 layout 继续反向依赖 feature
- 保持现有通知、登出和 topnav / app layout 的用户行为不变

## 非目标

- 本轮不继续处理 `Sidebar.vue` 的更深层拆分
- 本轮不处理 `challenge-topology-studio` 的 6 条 allowlist
- 本轮不重写通知 store、通知列表页 / 详情页、登录注册页或 auth API 契约
- 本轮不大范围迁移 feature 目录结构；只新增 layout shell 需要的最小 `widgets/*` bridge

## 输入依据

- `code/frontend/src/components/layout/AppLayout.vue`
- `code/frontend/src/components/layout/NotificationDrawer.vue`
- `code/frontend/src/components/layout/TopNav.vue`
- `code/frontend/src/components/layout/__tests__/AppLayout.test.ts`
- `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`
- `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
- `code/frontend/src/features/auth/model/useAuth.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AppLayout.vue` 当前通过 `useNotificationRealtime()` 启动通知初始化与 websocket，同步 owner 直接挂在 layout 上，但实现落在 feature model。
- `NotificationDrawer.vue` 当前通过 `useNotificationDrawer()` 承接抽屉开合后的导航、全部已读和 store 读写；这是基础壳里的共享 workflow，不是通知列表页专属流程。
- `TopNav.vue` 当前通过 `useAuth()` 获取 `logout()`，但它实际只消费“退出登录”这一条共享 session 能力，不依赖登录 / 注册流程。
- 因为 layout 是共享壳层，不能靠“把组件迁进 feature”来解决这 3 条 allowlist；正确 owner 是引入显式 `widgets/layout-shell` bridge，让 layout 只组合 cross-cutting workflow，而不直接碰 feature。

## 设计边界

### `widgets/layout-shell/*` 本轮负责

- `useLayoutNotificationRealtimeBridge()`：layout 消费通知实时同步的显式 bridge
- `useLayoutNotificationDrawerBridge()`：layout 消费通知抽屉 workflow 的显式 bridge
- `useLayoutSessionActionsBridge()`：layout 消费 session logout 的显式 bridge

### `features/notifications/model/*` 本轮继续负责

- 对外保留现有 public API
- 改为复用中立通知 composable，而不是继续持有 layout 专属 owner

### `features/auth/model/useAuth.ts` 本轮继续负责

- `login()` / `register()` owner 保持在 auth feature
- `logout()` 改为委托给中立 `useSessionLogout()`

### `components/layout/*` 本轮继续负责

- `AppLayout.vue`：shell 装配与通知状态下传
- `NotificationDrawer.vue`：抽屉模板、filter state、键盘 / body overflow cleanup
- `TopNav.vue`：route/theme/brand/notification/logout 装配

## 任务切片

### Slice 1：建立 layout shell widget bridge

- 目标：
  - 新增 `widgets/layout-shell/*`
  - 让 bridge 显式承接 layout 对 notifications/auth 的跨层消费
- 影响文件：
  - `code/frontend/src/widgets/layout-shell/*`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/notifications/model/useNotificationDrawer.test.ts src/features/notifications/model/useNotificationRealtime.test.ts`
- Review focus：
  - bridge 是否显式且足够薄，没有重新定义 feature workflow owner
  - layout shell 是否获得更清楚的跨切 workflow 组合边界

### Slice 2：layout 改用中立 owner 并清 allowlist

- 目标：
  - `AppLayout.vue`、`NotificationDrawer.vue`、`TopNav.vue` 改用 `widgets/layout-shell` bridge
  - 更新 layout 相关测试 mock
  - 删除 3 条 layout allowlist
- 影响文件：
  - `code/frontend/src/components/layout/*`
  - `code/frontend/src/components/layout/__tests__/*`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/AppLayout.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts src/components/layout/__tests__/TopNav.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - layout 是否已经不再反向依赖 feature
  - 通知抽屉、实时同步和退出登录行为是否仍由明确 owner 承担

### Slice 3：backlog / review 收尾

- 目标：
  - 更新 backlog 中 layout 这条的进展记录
  - 归档独立 review 结论
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 的结构债是否真的收口
  - 是否还留有新的反向依赖或 undocumented 例外

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/notifications/model/useNotificationDrawer.test.ts src/features/notifications/model/useNotificationRealtime.test.ts src/components/layout/__tests__/AppLayout.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts src/components/layout/__tests__/TopNav.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `Sidebar.vue` 仍然是 layout 大组件债的一部分，但本轮不改它的 owner，只更新 backlog 进展
- `useNotificationDrawer()` 仍然依赖 router，这条依赖目前通过 feature allowlist 继续存在；本轮只解决 layout 反向依赖，不处理 feature model 对 router 的设计问题
- 如果 `logout()` 在其他位置还有隐式复制逻辑，本轮不会主动全局统一，只保证 layout / auth 的 owner 一致
