# Reuse Decision

## Change type
service / handler / api / form

## Existing code searched
- `code/backend/internal/module/auth/contracts/token_service.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/module/auth/infrastructure/token_service_test.go`
- `code/backend/internal/module/auth/application/commands/service_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/frontend/src/features/profile/model/useSecuritySettingsPage.ts`
- `code/frontend/src/pages/profile/SecuritySettingsRoutePage.vue`
- `code/frontend/src/pages/profile/__tests__/SecuritySettings.test.ts`

## Similar implementations found
- `code/backend/internal/module/auth/infrastructure/token_service.go`
  - 已有 `CreateSession / GetSession / DeleteSession / RevokeAllUserSessions`，说明修复应继续收口在现有 token service，而不是新增第二套 auth 失效流程。
- `code/backend/internal/middleware/auth.go`
  - 现有鉴权统一走 `tokenService.GetSession(...)`，说明“改密后会话失效”的硬语义应落在 `GetSession` 可验证的服务端判定，而不是只靠 handler 里删 cookie。
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
  - 已有改密 happy path 集成测试和内存 token service，可直接扩展失败分支，不需要新建独立测试框架。
- `code/frontend/src/pages/profile/SecuritySettingsRoutePage.vue`
  - 路由页已作为密码提交流程与登出跳转 owner，前端只需要保持“仅在后端确认成功后登出”，不需要新增 store 或 API 层。

## Decision
extend_existing

## Reason
这次问题是现有 auth session 失效语义不够强，不是缺少新模块。最小正确修复是继续扩展现有 `tokenService` 与 `ChangePassword` 链路：

1. 在 `tokenService` 内把“用户级会话撤销”收口为鉴权时可判定的失效语义，覆盖没有反向索引的历史 session。
2. 在 `ChangePassword` handler 内把撤销失败变成可见失败，而不是继续返回成功。
3. 保持前端只在后端确认成功后执行本地登出和跳转。

这样能复用现有 auth 边界、测试基座和前端 owner，不引入第二套状态源。

## Files to modify
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/module/auth/infrastructure/token_service_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/frontend/src/pages/profile/__tests__/SecuritySettings.test.ts`
- `docs/reviews/security/2026-06-01-local-review-password-change-session-revocation.md`
- `feedback/2026-06-02-auth-session-revocation-must-not-depend-on-cleanup-index.md`

## After implementation
- 如果这次修复沉淀出可复用的 review / auth guardrail，会记录到 `feedback/`，而不是继续只留在一次性 review 结论里。
