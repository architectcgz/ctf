# Notification Drawer Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-notification-drawer-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/notification-drawer-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-notification-drawer-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-notification-drawer-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
  - `code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts`
  - `code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts`
- Classification check：预期按单条 feature router owner cleanup 处理；`useNotificationDrawer.ts` 不属于合理的 route-aware feature owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useLayoutNotificationDrawerBridge.ts` 继续持有 `useRouter()` 合理，因为它本来就是 layout 对 notifications feature 的 bridge owner。
- `useNotificationDrawer.ts` 应该只保留 notification workflow / presentation owner，不应直接认识 `vue-router`。
- 这条 allowlist 如果继续保留，会给 feature model 混入 router 开后门；本轮应当在 touched surface 内收掉。
- 当前实现已经把通知列表 / 通知详情的导航动作回收到 layout-shell bridge，`NotificationDrawer.vue` 本体无需继续改动也能保持原有交互。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/notifications/model/useNotificationDrawer.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/notification-drawer-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-notification-drawer-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-notification-drawer-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/notifications/model/useNotificationDrawer.ts code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有大量条目，需要后续继续按“page owner / non-page owner”逐条判定。
