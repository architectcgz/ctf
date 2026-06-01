> 状态：Current
> 事实源：通知抽屉 feature workflow owner、layout-shell bridge、前端架构 allowlist、NotificationDrawer 集成测试
> 替代：无

# Notification Drawer Router Owner Cleanup Plan

## 目标

- 把 `useNotificationDrawer.ts` 从 route-aware owner 收口回纯 feature workflow owner。
- 让 `useLayoutNotificationDrawerBridge.ts` 保留唯一 router owner。
- 删除对应 `featureRouterImportAllowlist` 条目。

## 非目标

- 不重构 `NotificationDrawer.vue` 当前的 layout P2 view-state / CSS 清理。
- 不改通知 store、`markAllRead()`、trigger slot contract 或通知详情路由语义。
- 不处理 `featureRouterImportAllowlist` 其它 feature 条目。

## 输入依据

- `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- `code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts`
- `code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts`
- `code/frontend/src/components/layout/NotificationDrawer.vue`
- `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useNotificationDrawer.ts` 当前主要负责 unread count、status pill、mark-all-read 和通知项动作编排；直接持有 `useRouter()` 会把 route owner 漂进 feature model。
- `useLayoutNotificationDrawerBridge.ts` 已经是 layout 对 notifications feature 的唯一桥接点，继续作为 route-aware owner 更合理。
- `NotificationDrawer.vue` 当前有未提交的 layout P2 改动，本轮应避开组件本体，缩到 model + bridge 边界。

## 设计边界

### `useLayoutNotificationDrawerBridge.ts` 本轮负责

- `useRouter()` 获取
- 导航到通知列表
- 导航到通知详情
- 把导航 callback 注入 feature workflow

### `useNotificationDrawer.ts` 本轮负责

- unread count / items / statusMeta / statusPillStyle
- notification type meta
- 打开 / 关闭抽屉
- `markAllRead()` workflow
- 调用外部注入的导航 callback

## 任务切片

### Slice 1：feature model 去掉 router 依赖

- 目标：
  - 从 `useNotificationDrawer.ts` 移除 `useRouter()` 和 `vue-router` import
  - 改为 callback 注入导航动作
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/notifications/model/useNotificationDrawer.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts`
- Review focus：
  - feature model 是否已经不再 import `vue-router`
  - 通知详情 / 查看全部通知行为是否保持不变

### Slice 2：bridge / allowlist / 护栏收尾

- 目标：
  - 由 `useLayoutNotificationDrawerBridge.ts` 接住 router owner
  - 删除 allowlist 条目
  - 更新 raw-source 护栏与 backlog 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/notifications/model/useNotificationDrawer.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - router owner 是否只回到 bridge，而没有漂到别的 layout / feature helper
  - allowlist 是否真实下降

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/notifications/model/useNotificationDrawer.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/notification-drawer-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-notification-drawer-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-notification-drawer-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/notifications/model/useNotificationDrawer.ts code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口一条 `featureRouterImportAllowlist`，不代表剩余条目都不合理；仍需逐条判定。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
