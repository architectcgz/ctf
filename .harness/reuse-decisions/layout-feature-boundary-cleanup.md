# Reuse Decision

## Change type
frontend architecture / layout-feature boundary cleanup / widget bridge extraction

## Existing code searched
- code/frontend/src/components/layout/AppLayout.vue
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/__tests__/AppLayout.test.ts
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.ts
- code/frontend/src/features/notifications/model/useNotificationRealtime.ts
- code/frontend/src/features/auth/model/useAuth.ts
- code/frontend/src/stores/notification.ts
- code/frontend/src/stores/auth.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 当前 layout 壳层已经把路由导航、主题切换、sidebar 开合这类基础设施 owner 放在 `components/layout` 与中立消费层，例如 `useTheme()`、`useWorkspaceShellNavigation()`。
- `notifications` 与 `auth` feature 里的这三段能力，实际被 layout 使用的只有通知实时同步、通知抽屉 workflow 和退出登录动作；它们都不是通知列表页或登录页私有流程。
- 现有测试已经把 `AppLayout.vue`、`NotificationDrawer.vue`、`TopNav.vue` 当成独立 owner 验证，说明这轮更适合抽中立 workflow，而不是把 layout 反向挂进 feature。

## Decision
refactor_existing

## Reason
`components/layout/*` 属于共享壳层，继续让它依赖 `@/features/notifications` 和 `@/features/auth` 会长期保留反向依赖。最小正确改动不是把 layout 组件迁进 feature，也不是在 shared composable 里重建 workflow owner，而是新增显式 `widgets/layout-shell` bridge，让 layout 只通过 widget 层消费 cross-cutting workflow。

## Files to modify
- .harness/reuse-decisions/layout-feature-boundary-cleanup.md
- docs/plan/impl-plan/2026-05-28-layout-feature-boundary-cleanup-implementation-plan.md
- docs/reviews/frontend/2026-05-28-layout-feature-boundary-cleanup-review.md
- code/frontend/src/widgets/layout-shell/index.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutNotificationRealtimeBridge.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts
- code/frontend/src/components/layout/AppLayout.vue
- code/frontend/src/components/layout/NotificationDrawer.vue
- code/frontend/src/components/layout/TopNav.vue
- code/frontend/src/components/layout/__tests__/AppLayout.test.ts
- code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts
- code/frontend/src/components/layout/__tests__/TopNav.test.ts
- code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts
- code/frontend/src/features/notifications/model/useNotificationRealtime.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `AppLayout.vue` 继续只负责 shell 装配和 sidebar / main / route transition owner，不再直接从 feature 拉通知实时同步。
- `NotificationDrawer.vue` 继续负责抽屉视图壳与 filter state，本地 workflow 改为依赖 `widgets/layout-shell` bridge，而不是直接依赖 feature model。
- `TopNav.vue` 继续负责 topnav route/theme/brand/notification/logout 装配，不再直连 `@/features/auth`。
- `features/notifications` 与 `features/auth` 继续保留现有 public API；layout 通过 widget bridge 组合它们，不再需要 shared composable 过渡层。
