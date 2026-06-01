> 状态：Current
> 事实源：`NotificationDrawer.vue` 当前 owner、`useLayoutNotificationDrawerBridge()` layout bridge、通知抽屉 raw-source / 主题 token 护栏测试
> 替代：无

# Notification Drawer Owner Cleanup Plan

## 目标

- 把 `NotificationDrawer.vue` 从“layout bridge + 本地筛选/摘要/overlay lifecycle + 样式壳体”收口成明确的 layout shell owner
- 为通知抽屉补齐本地视图 state composable 与独立 shell CSS 文件
- 保持通知抽屉的 trigger slot、关闭行为、筛选行为、跳转行为和主题 token 护栏对外 contract 不变

## 非目标

- 本轮不调整 `useNotificationDrawer()` 的 feature workflow owner
- 本轮不改通知 store、`markAllRead()` 语义或通知详情路由
- 本轮不继续拆 `NotificationDrawerHeader.vue`、`NotificationDrawerSummary.vue`、`NotificationDrawerTabs.vue`、`NotificationDrawerBody.vue`、`NotificationDrawerFooter.vue`

## 输入依据

- `code/frontend/src/components/layout/NotificationDrawer.vue`
- `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts`
- `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- layout 与 notifications feature 之间的 bridge 已经明确，当前主问题不在 feature import，而在 `NotificationDrawer.vue` 仍把本地 view state、overlay cleanup 与样式壳体混在父层。
- 这条线的本地状态都属于 layout 内局部行为，不应再继续塞入 feature bridge；它们更适合由一个局部 composable 承接。
- 抽屉壳体样式已经足够稳定，继续留在 SFC 内只会让父文件继续承担非 owner 的体量。

## 设计边界

### `NotificationDrawer.vue` 本轮继续负责

- `realtimeStatus` props contract
- `useLayoutNotificationDrawerBridge()` 的 workflow 装配
- trigger slot contract
- panel shell 组合与子组件 props / emits 编排

### `useNotificationDrawerViewState.ts` 本轮负责

- `activeFilter`
- `filterOptions`
- `filteredItems`
- `emptyState`
- `hasUnread`
- `unreadBadgeLabel`
- `drawerSummary`
- `Escape` 关闭与 `body` scroll lock cleanup

### `notificationDrawer.css` 本轮负责

- `.notification-shell`
- `.notification-panel`
- `.panel-inner`
- `.content-divider`
- `.notification-drawer-trigger`
- `.notification-drawer-trigger__badge`
- 本地 light / dark token 变量与响应式壳体样式

## 任务切片

### Slice 1：抽出通知抽屉本地视图 state owner

- 目标：
  - 新增 `useNotificationDrawerViewState.ts`
  - 把 filter / summary / empty state / overlay lifecycle 从父组件迁出
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts`
- Review focus：
  - layout bridge 是否仍是唯一远端通知 workflow owner
  - `Escape` 关闭与 scroll lock cleanup 是否仍只有一个 owner

### Slice 2：抽出通知抽屉 shell CSS

- 目标：
  - 新增 `notificationDrawer.css`
  - `NotificationDrawer.vue` 改为只保留模板组合与 imports
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - raw-source / theme-token 护栏是否仍覆盖 shell CSS
  - light / dark panel token 与移动端宽度 contract 是否没有漂移

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上的“layout bridge + 本地 view state + overlay lifecycle + 壳体样式”混写是否真的收口
  - 测试是否已经转成聚合源码视角，而不是继续假设样式和本地逻辑都留在父 SFC

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 通知抽屉的 header / summary / body 子组件本身仍各自带有样式和展示逻辑；本轮只收口父层 owner，不再继续扩成“整个通知中心再分层重排”。
- `Sidebar.vue` 和 `TopNav.vue` 仍是 layout P2 的主要剩余大组件；本轮完成后，layout 线的下一刀会更适合转到这两个 surface，而不是继续在通知抽屉内部扩大改动面。
