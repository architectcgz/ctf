# Reuse Decision

## Change type
frontend refactor / redirect owner convergence

## Existing code searched
- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/features/auth/model/useLoginViewPage.ts`
- `code/frontend/src/router/guards.ts`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/routes/route-helpers.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/pages/auth/__tests__/LoginRoutePage.test.ts`
- frontend sliced architecture migration ledger

## Similar implementations found
- `teacherRoutes.ts` 已经把 `/teacher/* -> /academy/*` redirect 收口成 allowlist，这是当前 legacy 教师端页面入口的唯一事实源。
- `sanitizeRedirectPath()` 已经是登录 redirect 参数的单点安全清洗 owner，适合继续承接 canonicalization，而不是把旧命名空间兼容散到 page / guard / auth model。

## Decision
refactor_existing

## Reason
当前登录 redirect 仍会把 `/teacher/*` 旧前端路径带入登录后导航。虽然运行时 redirect 还能兜底，但兼容 owner 已经分裂成两层：

- router route redirect
- 登录 redirect query 的后置跳转

最小正确改动是：

- 抽出教师端 legacy redirect 的单点 canonical path owner
- 让 `teacherRoutes.ts` 与 `sanitizeRedirectPath()` 复用同一套映射事实
- 保持登录页和 router guard 继续只消费通用 redirect sanitize owner

本轮不做：

- 不删除现有 `/teacher/*` redirect allowlist
- 不调整教师端 `/academy/*` 正式路由
- 不改教师端后端 API 命名

## Files to modify
- `.harness/reuse-decisions/login-redirect-canonicalization.md`
- `docs/plan/impl-plan/2026-05-31-login-redirect-canonicalization-plan.md`
- `docs/architecture/frontend/02-routing.md`
- frontend sliced architecture migration ledger (`frontend-sliced-architecture.md`)
- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `code/frontend/src/utils/__tests__/redirectPath.test.ts`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/pages/auth/__tests__/LoginRoutePage.test.ts`

## After implementation
- 登录 redirect 参数不会再把 `/teacher/*` 旧页面路径继续传到登录后导航。
- `/teacher/*` 兼容映射由单点 owner 维护，router redirect 与 redirect sanitize 不再各自抄一份事实。
- 教师端正式前端入口继续只认 `/academy/*`。
