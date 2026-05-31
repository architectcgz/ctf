# Reuse Decision

## Change type
frontend refactor / route compatibility retirement

## Existing code searched
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/utils/__tests__/redirectPath.test.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `docs/architecture/frontend/02-routing.md`
- frontend sliced architecture migration ledger

## Similar implementations found
- `redirectPath.ts` 已经把 legacy 教师端路径归一逻辑收口到共享 sanitize owner，说明登录后的 redirect 参数不依赖 router runtime redirect 才能回到 `/academy/*`。
- `sharedRoutes.test.ts` 已经集中检查 canonical route 与 legacy route allowlist，适合作为 runtime redirect 退场后的单点 route guardrail。

## Decision
refactor_existing

## Reason
上一轮已经用源码级 guardrail 限制新的 `/teacher/*` 前端页面路径 producer，只保留 `teacherLegacyRedirect.ts` 作为兼容 owner。当前活跃前端代码已不再直接产出 `/teacher/*` 页面跳转，继续保留 router runtime redirect 只是在维护外部兼容入口。

本轮最小正确改动是：

- 从 router runtime 移除 `/teacher/*` legacy 页面 route
- 保留 `sanitizeRedirectPath()` 的 canonicalize 逻辑，继续收口登录 redirect 参数里的旧路径
- 把当前事实同步回路由文档和迁移台账

本轮不做：

- 不删除 `teacherLegacyRedirect.ts` 的 canonicalize helper
- 不修改教师端后端 API `/teacher/*`
- 不改 `/academy/*` canonical 页面路由

## Files to modify
- `.harness/reuse-decisions/teacher-legacy-runtime-redirect-retirement.md`
- `docs/plan/impl-plan/2026-05-31-teacher-legacy-runtime-redirect-retirement-plan.md`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `docs/architecture/frontend/02-routing.md`
- frontend sliced architecture migration ledger (`frontend-sliced-architecture.md`)

## After implementation
- router runtime 不再注册 `/teacher/*` 前端页面 route。
- 登录页 redirect 参数仍可把 legacy 教师端路径归一到 `/academy/*`。
- 教师端页面正式事实只剩 `/academy/*`，旧 `/teacher/*` 页面入口彻底退出 router。
