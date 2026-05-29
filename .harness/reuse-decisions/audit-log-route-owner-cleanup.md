# Reuse Decision

## Change type
frontend refactor / audit log route owner cleanup

## Existing code searched
- code/frontend/src/features/audit-log/model/useAuditLogPage.ts
- code/frontend/src/views/platform/AuditLog.vue
- code/frontend/src/views/platform/__tests__/AuditLog.test.ts
- code/frontend/src/views/platform/__tests__/auditLogPageStateExtraction.test.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeQueryTransport.ts`
- `features/challenge-list/model/useChallengeListPage.ts`
- `views/platform/__tests__/AuditLog.test.ts`

## Decision
refactor_existing

## Reason
`useAuditLogPage.ts` 当前直接 import `vue-router`，但它实际碰 router 的面很窄：

- 读取 `action / resource_type / actor_user_id / page` 这些 query
- 在筛选和分页后做 `replace`

这条如果新建 route wrapper，只是把 allowlist 从一个 page model 平移到另一个 page model；没有必要。更合理的收口是：

- query hydrate、筛选默认值、auto-apply 节奏、分页和加载 owner 继续留在 `useAuditLogPage.ts`
- `useRoute / useRouter` transport 下沉到已有的 `routeQueryTransport.ts`

这样能在不改 audit log 页面交互和数据 owner 的前提下，真实减少 `featureRouterImportAllowlist`。

## Files to modify
- .harness/reuse-decisions/audit-log-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-audit-log-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-audit-log-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/audit-log/model/useAuditLogPage.ts
- code/frontend/src/views/platform/__tests__/AuditLog.test.ts
- code/frontend/src/views/platform/__tests__/auditLogPageStateExtraction.test.ts

## After implementation
- `useAuditLogPage.ts` 不再 import `vue-router`
- audit log 的 query hydrate、auto-apply 和分页 owner 保持不变
- `featureRouterImportAllowlist` 再收掉 `features/audit-log/model/useAuditLogPage.ts -> vue-router`
