# Reuse Decision

## Change type
module public error contract / auth session boundary cleanup

## Existing code searched

- `code/backend/internal/module/auth/contracts/`
- `code/backend/internal/module/auth/infrastructure/token_service*.go`
- `code/backend/internal/middleware/auth.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/apperror/`

## Similar implementations found

- `code/backend/internal/module/auth/contracts/public_errors.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/middleware/auth.go`
- `code/backend/internal/apperror/error.go`

## Decision
refactor_existing

## Reason

上一刀已经把 auth 公开错误收回到了模块 `contracts`，但 session 鉴权链路仍把多种失败统一折叠成 `ErrUnauthorized`，导致 `ErrTokenInvalid`、`ErrAccessTokenExpired` 只是定义存在，没有真正进入运行时链路。

这次直接在现有 owner 内收口：

- `middleware.Auth` 只负责区分“没有认证信息”和“token service 返回的 auth 公开错误”
- `token_service` 负责把 session 缺失、过期、损坏映射成 auth 模块自己的公开错误
- `apperror.AppError` 补齐 `errors.Is` 语义，保证模块公开错误在 `WithCause` 后仍保持稳定可判定
- 暂不强行使用 `ErrTokenRevoked`，因为当前实现没有单独的吊销状态来源

## Files to modify

- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/module/auth/infrastructure/token_service_test.go`
- `code/backend/internal/middleware/auth.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/apperror/error.go`
- `code/backend/internal/apperror/error_test.go`

## After implementation

- 若后续引入 refresh token 或显式 session blacklist，再决定 `ErrRefreshTokenExpired` / `ErrTokenRevoked` 的真实 owner 和状态来源。
