# Reuse Decision

## Change type
frontend refactor / compatibility retirement

## Existing code searched
- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `code/frontend/src/utils/__tests__/redirectPath.test.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `docs/architecture/frontend/02-routing.md`
- frontend sliced architecture migration ledger

## Similar implementations found
- `sanitizeRedirectPath()` 已经是登录 redirect 参数的唯一安全 owner，适合继续把 legacy 页面路径的处理收口在这里，而不是把兼容判断再散回 router 或页面。
- `resolveLoginRedirectTarget()` 与 `getRoleDashboardPath()` 已经承担 `/` -> 角色默认页 的回退逻辑，因此废除 `/teacher/*` redirect 参数兼容后，不需要新增第二套 fallback owner。

## Decision
refactor_existing

## Reason
教师端 `/teacher/*` 页面 route 已经退出 router runtime。继续保留 `teacherLegacyRedirect.ts` 只是在把已经退场的旧页面路径继续偷偷映射回 `/academy/*`。

本轮最小正确改动是：

- 删除 `teacherLegacyRedirect.ts`
- 让 `sanitizeRedirectPath()` 直接拒绝 legacy `/teacher/*` 页面路径，回退成 `/`
- 依赖现有登录 fallback，把这类历史 redirect 导回角色默认首页
- 把源码 guardrail 收紧成“前端活跃源码里不再允许旧教师端页面路径 literal”

本轮不做：

- 不改教师端后端 `/api/v1/teacher/*` HTTP path
- 不改教师端 canonical `/academy/*` 页面路由
- 不扩展成全量旧路径白名单校验

## Files to modify
- `.harness/reuse-decisions/teacher-legacy-redirect-sanitize-retirement.md`
- `docs/plan/impl-plan/2026-05-31-teacher-legacy-redirect-sanitize-retirement-plan.md`
- `code/frontend/src/utils/redirectPath.ts`
- `code/frontend/src/utils/__tests__/redirectPath.test.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `docs/architecture/frontend/02-routing.md`
- frontend sliced architecture migration ledger (`frontend-sliced-architecture.md`)
- `code/frontend/src/utils/teacherLegacyRedirect.ts`

## After implementation
- 登录 redirect 参数里的 `/teacher/*` 前端页面路径不会再被 canonicalize 成 `/academy/*`。
- 这类历史 redirect 会回退到角色默认首页，而不是进入已退场页面。
- 前端活跃源码里不再保留旧教师端页面路径兼容 owner。
