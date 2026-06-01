> 状态：Current
> 事实源：`NotificationDrawer.vue` 的 layout owner 与拆分边界
> 替代：无

# Notification Drawer Decomposition Implementation Plan

## 目标

- 把 `code/frontend/src/components/layout/NotificationDrawer.vue` 从单文件大壳拆成“layout owner + 稳定子视图”的结构。
- 保留 `NotificationDrawer.vue` 作为通知抽屉的交互 owner，不改变 `useNotificationDrawer.ts` 的 store / router / 批量已读职责。

## 非目标

- 本轮不把 `NotificationDrawer` 迁到 `features/*/ui`。
- 本轮不改 `TopNav.vue` 的通知按钮交互和 slot contract。
- 本轮不新增通知实时状态文案、不改通知 store 结构、不改通知 API。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/layout/NotificationDrawer.vue`
- `code/frontend/src/components/layout/TopNav.vue`
- `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`
- `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `NotificationDrawer.vue` 仍应留在 `components/layout/`，因为它是全局 layout 能力的一部分，并通过 slot 允许 `TopNav` 自己持有按钮壳。
- `useNotificationDrawer.ts` 已经是通知抽屉的行为 owner；本轮只收口视图分块，不改变数据流。
- 最稳妥的切片不是继续迁 owner，而是把 header、summary、tabs、content、footer 这些稳定块拆成局部子组件。

## 设计边界

### `components/layout/NotificationDrawer.vue` 继续负责

- 具名 `trigger` slot 与默认 trigger
- `Teleport`、overlay、drawer panel 壳
- open/close、Escape 关闭、backdrop 关闭、body scroll lock
- `useNotificationDrawer()` 返回值的装配和局部 filter state

### `components/layout/notification-drawer/*` 本轮负责

- 抽屉头部视觉结构
- 摘要行与“全部设为已读”动作壳
- 全部/未读/已读 tabs 壳
- empty state 与通知列表展示
- 底部“查看全部通知”动作壳

### 本轮不负责

- 调整通知 store、通知 API、realtime owner
- 变更 `TopNav` 的外部 slot contract
- 重做抽屉视觉或主题 token 体系

## 任务切片

### Slice 1：补 layout 子组件并保留父壳 owner

- 目标：
  - 新增 `components/layout/notification-drawer/` 下的稳定子组件
  - `NotificationDrawer.vue` 只保留 owner、slot、panel shell 与局部 filter 逻辑
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts`
- Review focus：
  - owner 是否仍清楚留在父组件
  - 子组件是否只承接视图，不重新拿走 router/store owner

### Slice 2：同步源码断言与 backlog

- 目标：
  - 更新原始源码断言，改为聚合抽屉父壳与子组件源码
  - 更新 backlog 中 `NotificationDrawer.vue` 的进展
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - guardrail 是否仍能覆盖抽屉壳、主题 token 和关键布局约束

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- 这轮优先收口 `NotificationDrawer.vue`，`Sidebar.vue` 与 `TopNav.vue` 仍留在后续 `P2` 切片。
- `useNotificationDrawer.ts` 里若还存在与当前 UI 不再对应的旧派生，可能需要在实现时一并清掉。
