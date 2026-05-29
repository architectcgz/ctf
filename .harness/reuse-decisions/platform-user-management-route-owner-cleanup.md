# Reuse Decision

## Change type
frontend refactor / platform user management route owner cleanup

## Existing code searched
- code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts
- code/frontend/src/views/platform/UserManage.vue
- code/frontend/src/views/platform/__tests__/UserManage.test.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/features/audit-log/model/useAuditLogPage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeQueryTransport.ts`
- `features/audit-log/model/useAuditLogPage.ts`
- `views/platform/__tests__/AuditLog.test.ts`

## Decision
refactor_existing

## Reason
`usePlatformUserManagePage.ts` 现在已经是 `UserManage` 的 page owner，但它直接 import `vue-router` 的职责面其实很窄：

- 读取 `route.query.panel`
- 在 overview / import 之间做 `replace`

这条如果继续把 `vue-router` 留在 page model，`featureRouterImportAllowlist` 不会下降；但如果为了消掉 allowlist 再新建一层 user-management-specific route wrapper，只是在 page owner 外再包一层中间态，也没有新增结构收益。更合理的收口方式是：

- panel 解析 / query 构建继续留在 `useUserGovernancePanelRoute.ts`
- page-level 的 `activePanel`、mounted refresh 和用户治理 workflow 继续留在 `usePlatformUserManagePage.ts`
- `useRoute / useRouter` transport 下沉到已有的 `routeQueryTransport.ts`

这样能在不改变 `UserManage` 页面 owner 的前提下，真实减少一条 feature 直接依赖 `vue-router` 的 allowlist。

## Files to modify
- .harness/reuse-decisions/platform-user-management-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-platform-user-management-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-platform-user-management-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts
- code/frontend/src/views/platform/__tests__/UserManage.test.ts

## After implementation
- `usePlatformUserManagePage.ts` 不再 import `vue-router`
- 用户治理页的 panel owner 继续留在 page model，但 router transport 改为复用共享 query transport
- `featureRouterImportAllowlist` 再收掉 `features/platform-user-management/model/usePlatformUserManagePage.ts -> vue-router`
