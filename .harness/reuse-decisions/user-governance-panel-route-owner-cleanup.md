# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts
- code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts
- code/frontend/src/features/platform-user-management/ui/UserGovernancePage.vue
- code/frontend/src/views/platform/UserManage.vue
- code/frontend/src/views/platform/__tests__/UserManage.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useStudentAnalysisNavigation.ts`、`useNotificationDrawer.ts` 这类下游 helper 已经改成纯 contract / callback，不再直接拿 `vue-router`。
- `usePlatformUserManagePage.ts` 当前已经是 `UserManage` 的 page owner，适合承接 `panel` query 读写。

## Decision
refactor_existing

## Reason
`useUserGovernancePanelRoute.ts` 不是完整 page owner，只是在做 `panel=overview/import` 的 query 解析和切换。继续把 `useRoute()` / `useRouter()` 挂在它上面，会让 `UserGovernancePage.vue` 这种 UI shell 间接持有 route owner。

最小正确改动是：

- 把 `activePanel` / `switchPanel` 收回 `usePlatformUserManagePage.ts`
- 让 `UserGovernancePage.vue` 改为纯 props / emits contract
- 把 `useUserGovernancePanelRoute.ts` 降成无 `vue-router` 依赖的纯 helper
- allowlist 从非 page helper 挪到 page owner

本轮不做：

- 不改用户治理列表、导入、创建/编辑/删除的业务流程
- 不继续拆 `usePlatformUserManagePage.ts`
- 不处理其它 router allowlist 条目

## Files to modify
- .harness/reuse-decisions/user-governance-panel-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-user-governance-panel-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-user-governance-panel-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts
- code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts
- code/frontend/src/features/platform-user-management/model/index.ts
- code/frontend/src/features/platform-user-management/ui/UserGovernancePage.vue
- code/frontend/src/views/platform/UserManage.vue
- code/frontend/src/views/platform/__tests__/UserManage.test.ts

## After implementation
- `useUserGovernancePanelRoute.ts` 不再 import `vue-router`
- `usePlatformUserManagePage.ts` 成为 `UserManage` 的唯一 panel query owner
- `UserGovernancePage.vue` 退回纯展示壳
- allowlist 从 helper owner 切到 page owner
