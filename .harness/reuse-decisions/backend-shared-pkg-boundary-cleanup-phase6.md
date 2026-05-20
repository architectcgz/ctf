# Reuse Decision

## Change type
shared kernel / module public error contract / boundary cleanup

## Existing code searched

- `code/backend/internal/apperror/`
- `code/backend/internal/module/auth/{contracts,application,api,infrastructure}/`
- `code/backend/internal/module/identity/{contracts,application}/`

## Similar implementations found

- `code/backend/internal/module/challenge/contracts/errors.go`
- `code/backend/internal/module/contest/contracts/errors.go`
- `code/backend/internal/module/instance/contracts/errors.go`
- `code/backend/internal/module/auth/contracts/public_errors.go`
- `code/backend/internal/module/identity/contracts/errors.go`

## Decision
refactor_existing

## Reason

`internal/apperror` 里剩余的 auth / identity 错误并不是跨模块共享内核，而是模块自己的对外错误语义：

- auth owner：登录失败、账户锁定、CAS 未配置 / 票据无效等
- identity owner：管理员创建 / 更新用户时的唯一性冲突、修改密码时的旧密码错误 / 新旧密码相同

继续把它们留在共享层，会让 `apperror` 同时承担公共错误内核和具体模块公开契约，owner 还是不清。更合理的边界是：

- `internal/apperror` 只保留真正平台级公共错误
- auth 对外错误收回 `internal/module/auth/contracts`
- identity 对外错误收回 `internal/module/identity/contracts`
- identity 仓储分支错误如 `ErrUserNotFound` 继续保留在 `identity/contracts`，但不再与公开 `AppError` 混在共享层

## Files to modify

- `code/backend/internal/apperror/error.go`
- `code/backend/internal/module/auth/contracts/**/*.go`
- `code/backend/internal/module/auth/{application,api}/**/*.go`
- `code/backend/internal/module/identity/contracts/**/*.go`
- `code/backend/internal/module/identity/application/**/*.go`

## After implementation

- 若模式稳定，后续可把“共享错误内核 + 模块公开错误契约 + 模块内部 sentinel 分支错误”沉淀到长期 reuse 规则。
