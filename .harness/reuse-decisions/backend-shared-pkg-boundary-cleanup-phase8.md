# Reuse Decision

## Change type
module public error contract cleanup

## Existing code searched

- `code/backend/internal/module/auth/contracts/public_errors.go`
- `code/backend/internal/module/auth/**/*.go`

## Similar implementations found

- `code/backend/internal/module/auth/contracts/public_errors.go`
- `code/backend/internal/module/contest/contracts/errors.go`
- `code/backend/internal/module/instance/contracts/errors.go`

## Decision
refactor_existing

## Reason

`auth/contracts` 里目前有两类错误：

- 已经接入实际链路的公开错误，例如 `ErrInvalidCredentials`、`ErrAccessTokenExpired`、`ErrTokenInvalid`
- 还没有任何运行时来源的预留错误，例如 `ErrRefreshTokenExpired`、`ErrTokenRevoked`

对于当前这轮 owner 收口，保留“未来也许会用到”的公开错误会让 contract 比真实能力更宽，反而模糊边界。当前仓库没有 refresh token 流程，也没有独立 revoke store，因此这两个错误不应该先暴露出来。

## Files to modify

- `code/backend/internal/module/auth/contracts/public_errors.go`

## After implementation

- 若未来补 refresh token / session blacklist，再由 auth owner 在对应实现落地时重新引入公开错误契约。
