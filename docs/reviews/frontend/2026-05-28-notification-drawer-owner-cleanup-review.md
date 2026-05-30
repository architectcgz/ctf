# Notification Drawer Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-notification-drawer-owner-cleanup-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/notification-drawer-owner-cleanup.md`
    - `docs/plan/impl-plan/2026-05-28-notification-drawer-owner-cleanup-plan.md`
    - `docs/reviews/frontend/2026-05-28-notification-drawer-owner-cleanup-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/components/layout/NotificationDrawer.vue`
    - `code/frontend/src/components/layout/notification-drawer/useNotificationDrawerViewState.ts`
    - `code/frontend/src/components/layout/notification-drawer/notificationDrawer.css`
    - `code/frontend/src/components/layout/__tests__/NotificationDrawer.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Classification check：同意按 layout 共享基础设施上的非 trivial frontend refactor 处理，目标是继续收口 `NotificationDrawer.vue` 的本地 owner 混写，而不是改通知 feature workflow。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `NotificationDrawer.vue` 现在只保留 trigger slot、layout bridge workflow 装配和 panel 组合，不再继续混放 filter / summary / dismiss lifecycle / 壳体样式。
- `useNotificationDrawerViewState.ts` 已明确承接 `activeFilter`、`filteredItems`、`emptyState`、`hasUnread`、`unreadBadgeLabel`、`drawerSummary` 以及 `Escape` 关闭与 `body` scroll lock cleanup；这批本地视图状态不再滞留在父 SFC。
- `notificationDrawer.css` 已承接通知抽屉 shell、trigger 与 light/dark token 变量，raw-source 与 shared-theme 护栏也已切到聚合源码视角，不再假设样式必须留在父 SFC。
- `NotificationDrawer.vue` 文件体量从约 `496` 行降到 `118` 行；本轮 touched surface 上“layout bridge + 本地视图 state + overlay lifecycle + 壳体样式”混写债已完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/NotificationDrawer.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据。当前流程里没有额外派生 reviewer，所以这条缺口需要在交付说明里明确。
- `Sidebar.vue` 与 `TopNav.vue` 仍然是 layout P2 的主要剩余大组件；本轮只处理通知抽屉本地 owner，不扩大到 route 判定、breadcrumbs 或 sidebar nav contract。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中这条 layout P2 对 `NotificationDrawer.vue` 的“更深层本地视图 owner 清理”已在 touched surface 内完成收口；当前 residual 重点已经进一步收敛到 `Sidebar.vue`、`TopNav.vue`，而不再是通知抽屉父组件继续混写本地筛选、overlay lifecycle 与壳体样式。
