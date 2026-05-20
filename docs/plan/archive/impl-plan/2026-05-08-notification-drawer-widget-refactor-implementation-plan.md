# Notification Drawer Widget Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把通知右侧抽屉重构为边界清晰的 widget：TopNav 拥有 trigger，NotificationDrawer 拥有 drawer panel 与通知工作流，移除脆弱的 `:deep()` 样式耦合，并顺手收口头部节奏与状态呈现。

**Architecture:** 保留 `useNotificationDrawer` 作为通知抽屉 workflow owner，但把触发器从组件内部默认实现提升为 slot contract。`TopNav` 通过 `trigger` scoped slot 提供导航风格按钮；`NotificationDrawer` 只暴露 `open / toggle / unreadCount / triggerRef` 等最小装配能力，并在自身内部负责 header、filters、status、list、footer 的布局和主题变量。

**Tech Stack:** Vue 3 `<script setup>`、TypeScript、Vitest、Vue Test Utils、项目现有 design tokens / `SlideOverDrawer`

---

## 依据

- 当前 `TopNav.vue` 通过 `:deep(.notification-trigger)` 反向覆盖 `NotificationDrawer` 内部按钮样式，已经形成 owner 混杂。
- `useNotificationDrawer.ts` 已经计算 `statusMeta` / `statusPillStyle`，但 `NotificationDrawer.vue` 没有渲染，状态链路对用户不可见。
- `NotificationDrawer.vue` 头部与筛选区域存在额外 `padding-inline`，与抽屉标题和列表内容不在同一视觉参考线。
- `.notification-list` 才持有顶部边线，切到空状态时结构线消失，存在视觉跳动。

## TDD 判定

- Mixed UI task：不做完整 TDD 流程，但会先补或调整受影响测试，再运行 focused tests + `typecheck`。

## Architecture-fit evaluation

这次直接触达一个已知结构债：通知 trigger 的 owner 当前同时落在 `TopNav` 和 `NotificationDrawer`，靠深度样式覆盖维持视觉一致。该债务必须在本轮收口，完成标准是：

- `TopNav` 不再通过 `:deep(.notification-trigger)` 覆盖通知按钮
- `NotificationDrawer` 不再假设自己一定拥有唯一 trigger UI
- trigger 与 drawer panel 的契约通过显式 slot / slot props 表达
- 头部 spacing 与分割线 ownership 在 drawer 内部收口，不再靠外层补丁样式

如果实现后仍保留 `TopNav -> :deep(.notification-trigger)` 这条链路，则视为本轮未完成。

### Task 1: 重构 NotificationDrawer 组件契约

**Files:**
- Modify: `code/frontend/src/components/layout/NotificationDrawer.vue`
- Modify: `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- Test: `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`

- [x] **Step 1: 定义新的 trigger slot contract**
  目标：为 `NotificationDrawer` 增加 `trigger` scoped slot，slot props 至少包含 `open`、`toggleOpen`、`unreadCount`、`triggerRef`，并保留内部 fallback trigger 以兼容其他调用方。

- [x] **Step 2: 渲染通知状态与统一头部结构**
  目标：在 header 区域接入 `statusMeta` / `statusPillStyle`，同时统一 summary、filters、body divider 的结构，让空状态与列表状态共享稳定的上边界。

- [x] **Step 3: 收口头部与列表 spacing ownership**
  目标：移除造成左侧参考线漂移的局部 `padding-inline`，改为由 drawer 自身的 header/body shell 统一控制对齐；分割线不再依赖 `.notification-list` 独占。

- [x] **Step 4: 更新 NotificationDrawer tests**
  目标：断言新 slot contract、状态 pill、稳定 divider 和新 header class 结构，而不是继续锁死旧 trigger 细节。

### Task 2: 重构 TopNav 装配关系

**Files:**
- Modify: `code/frontend/src/components/layout/TopNav.vue`
- Test: `code/frontend/src/components/layout/__tests__/TopNav.test.ts`

- [x] **Step 1: 在 TopNav 中通过 `trigger` slot 渲染通知按钮**
  目标：让 TopNav 直接输出符合导航栏样式的按钮，按钮视觉与其他 icon button 同级，不再深度覆盖子组件内部类名。

- [x] **Step 2: 删除通知 trigger 的 `:deep()` 样式耦合**
  目标：去掉 `:deep(.notification-trigger)` 与 hover/admin 变体覆盖，改为普通 TopNav 按钮样式和可读的 active 状态类。

- [x] **Step 3: 调整 TopNav tests**
  目标：mock 新的 slot contract 或验证新的装配方式，确保通知按钮仍接入 `notificationStatus` 并保持导航壳层结构稳定。

### Task 3: 验证与独立 review

**Files:**
- Review: `code/frontend/src/components/layout/NotificationDrawer.vue`
- Review: `code/frontend/src/components/layout/TopNav.vue`
- Review: `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- Test: `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`
- Test: `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
- Test: `code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts`
- Test: `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Test: `code/frontend/src/features/__tests__/featureBoundaries.test.ts`

- [x] **Step 1: 运行 focused tests**
  Run: `npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/components/layout/__tests__/TopNav.test.ts src/features/notifications/model/useNotificationDrawer.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/features/__tests__/featureBoundaries.test.ts`

- [x] **Step 2: 运行类型检查**
  Run: `npm run typecheck`

- [x] **Step 3: 做独立 review/self-review**
  检查点：
  - trigger owner 是否只在 TopNav
  - drawer panel owner 是否只在 NotificationDrawer
  - 是否还有 `:deep()` 跨组件样式耦合
  - 空状态与列表状态的结构线是否稳定
  - 状态 pill 是否真实接入 UI 而不是死代码

## Rollback / Recovery Notes

- 如果 slot contract 引入回归，可单独回退 `NotificationDrawer.vue` 与 `TopNav.vue` 的装配改动，不影响通知 store / API。
- 本轮不改通知数据模型与路由，因此回退不涉及后端或持久化数据恢复。
