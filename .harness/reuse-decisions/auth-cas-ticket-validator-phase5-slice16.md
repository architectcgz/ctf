# Reuse Decision

## Change type

service / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/auth/application/commands/cas_service.go`
- `code/backend/internal/module/auth/application/commands/cas_service_test.go`
- `code/backend/internal/module/auth/application/queries/cas_service.go`
- `code/backend/internal/module/auth/runtime/module.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- `code/backend/internal/module/auth/application/commands/cas_service.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 CAS 登录能力，而是 `auth/application/commands/cas_service.go` 同时承担了用户同步编排和 CAS ticket 的 HTTP / XML 校验。最小正确方案不是把整个 CAS service 拆散，也不是把 auth 继续做成宽 helper，而是沿用刚刚在 challenge 里验证过的模式：application 只保留登录编排与用户同步，新增 auth 模块内窄 port 表达 CAS ticket validator，由 infrastructure 统一承接 request、XML 解析、用户名校验和 ticket invalid sentinel。

## Files to modify

- `code/backend/internal/module/auth/application/commands/cas_service.go`
- `code/backend/internal/module/auth/application/commands/cas_service_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/module/auth/runtime/module.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-auth-cas-ticket-validator-phase5-slice16-implementation-plan.md`

## After implementation

- 如果后续 auth 还要接 LDAP、OIDC 或其他校验入口，优先新增同类 validator port / adapter，而不是把网络请求重新塞回 application service
- 如果 CAS login/query 未来要统一 URL 构造逻辑，优先在 auth 内收敛共享 helper，但不反向让 application 重新持有 HTTP concrete
