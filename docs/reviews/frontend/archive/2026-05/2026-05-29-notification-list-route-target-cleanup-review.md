# Notification List Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-notification-list-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/notification-list-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-notification-list-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-notification-list-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
  - `code/frontend/src/views/notifications/NotificationList.vue`
  - `code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- Classification check：同意按单一 route target cleanup 处理；`useNotificationListPage.ts` 的 router 依赖只剩单次详情跳转，不是合理的 reviewed route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useNotificationListPage.ts` 现在只保留通知列表数据、分页、分类、批量已读、发布抽屉和刷新提示 owner，不再直接 import `vue-router`。
- 通知详情跳转已改成纯 `notificationDetailRoute()` helper，`NotificationList.vue` 直接用 `RouterLink` 渲染通知行。
- `NotificationList.test.ts` 已继续覆盖“点击通知进入详情页”与“page model 不再 import vue-router”的护栏。
- `featureRouterImportAllowlist` 已从这条通知列表 page model 上再减少 1 条。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/notifications/__tests__/NotificationList.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/notification-list-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-notification-list-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-notification-list-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/notifications/model/useNotificationListPage.ts code/frontend/src/views/notifications/NotificationList.vue code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `useNotificationDetailPage.ts` 还保留 route param owner，这轮不一并处理。
