# Reuse Decision

## Change type
frontend architecture / layout component / decomposition

## Existing code searched
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts
- code/frontend/src/stores/notification.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `ContestAwdConfigWorkspaceShell.vue`、`AWDChallengeLibraryPage.vue`、`ClassStudentsPage.vue` 迁移说明，本轮可以继续沿用“先定 owner，再按稳定 UI 区块拆分”的最小切片方式。
- `TopNav.vue` 已经通过具名 `trigger` slot 自定义通知按钮，因此 `NotificationDrawer.vue` 本身可以只继续持有抽屉壳和交互，把内部静态视图片区块拆成子组件。
- `useNotificationDrawer.ts` 已经承担 store / router / 批量已读 owner，本轮不需要新增 composable，也不需要把通知流程迁到新的 feature owner。

## Decision
refactor_existing

## Reason
这次不是新增通知能力，而是收口布局层超大组件债。`NotificationDrawer.vue` 当前同时承载默认 trigger、drawer shell、header、summary、filters、empty/list body、footer 和大段样式，已经明显超出普通布局组件可维护范围。最小正确改动是保留 `components/layout/NotificationDrawer.vue` 作为 layout owner，只把稳定的视图片区块拆到同目录下的子组件，避免把 router/store/scroll-lock owner 打散。

## Files to modify
- .harness/reuse-decisions/notification-drawer-decomposition.md
- docs/plan/impl-plan/2026-05-27-notification-drawer-decomposition-implementation-plan.md
- docs/reviews/frontend/2026-05-27-notification-drawer-decomposition-review.md
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/notification-drawer/NotificationDrawerHeader.vue
- code/frontend/src/components/layout/notification-drawer/NotificationDrawerSummary.vue
- code/frontend/src/components/layout/notification-drawer/NotificationDrawerTabs.vue
- code/frontend/src/components/layout/notification-drawer/NotificationDrawerBody.vue
- code/frontend/src/components/layout/notification-drawer/NotificationDrawerFooter.vue
- code/frontend/src/components/layout/notification-drawer/types.ts
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `NotificationDrawer.vue` 继续负责 open/close、trigger slot、Escape/backdrop 关闭、body scroll lock 和与 `useNotificationDrawer()` 的对接。
- 子组件只承接 header、summary、filter tabs、empty/list body、footer 这些稳定视图，不新增新的请求 owner 或 route owner。
- 如果拆分顺利，这条 `P2` 布局层超大组件 backlog 会从“待开始”进入“已对 NotificationDrawer 落第一刀”的状态。
