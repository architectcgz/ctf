# Reuse Decision

## Change type
frontend refactor / auth login route owner cleanup

## Existing code searched
- code/frontend/src/features/auth/model/useLoginPage.ts
- code/frontend/src/features/auth/model/useLoginPage.test.ts
- code/frontend/src/features/auth/model/useLoginViewPage.ts
- code/frontend/src/router/guards.ts
- code/frontend/src/router/__tests__/guards.test.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/utils/roleRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeNavigationTransport.ts`
- `composables/routeQueryTransport.ts`
- `utils/roleRoutes.ts`

## Decision
refactor_existing

## Reason
`useLoginPage.ts` 当前有两类边界越层：

- 直接 import `vue-router` 读取 redirect query 并做 push
- 直接 import `router/guards.ts` 只为了复用 `sanitizeRedirectPath()`

更合理的落点是：

- redirect query 读取改由共享 `routeQueryTransport`
- 登录成功后的跳转改由共享 `routeNavigationTransport`
- `sanitizeRedirectPath()` 提升到更中性的 `utils/*`，由 router guards 和 login page 共用

这样能一次清掉 `auth` 剩余的两条 allowlist，同时保持登录提交、重复提交 guard、probe 提示和默认 dashboard fallback 仍留在 login page owner。

## Files to modify
- .harness/reuse-decisions/auth-login-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-auth-login-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-auth-login-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/auth/model/useLoginPage.ts
- code/frontend/src/features/auth/model/useLoginPage.test.ts
- code/frontend/src/router/guards.ts
- code/frontend/src/utils/redirectPath.ts

## After implementation
- `useLoginPage.ts` 不再 import `vue-router`
- `useLoginPage.ts` 不再 import `@/router/guards`
- redirect query 改由 `routeQueryTransport` 读取
- 登录成功跳转改由 `routeNavigationTransport` 执行
- `sanitizeRedirectPath()` 改为由中性 util owner 提供
- `featureRouterImportAllowlist` 清空
