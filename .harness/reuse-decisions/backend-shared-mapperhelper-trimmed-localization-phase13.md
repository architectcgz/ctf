# Reuse Decision

## Change type
shared helper localization

## Existing code searched

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/module/identity/application/**/*.go`
- `code/backend/internal/middleware/*.go`
- `code/backend/internal/module/contest/domain/*.go`
- `code/backend/internal/**/*.go`

## Similar implementations found

- `code/backend/internal/module/identity/application/commands/support.go`
- `code/backend/internal/module/identity/application/queries/support.go`
- `code/backend/internal/middleware/audit.go`
- `code/backend/internal/module/contest/domain/awd_readiness.go`

## Decision
refactor_existing

## Reason

`NormalizeOptionalTrimmedString` 现在的调用面已经收敛成 3 个明确 owner：

- `identity`：用户资料 / 管理员用户出参整形
- `middleware`：审计日志里的 `UserAgent`
- `contest/domain`：AWD 试跑预览里的访问地址清洗

这类 trimmed string 归一化已经不是稳定 shared kernel。继续挂在 `internal/shared/mapperhelper` 下，会让 shared 包继续承载已经有明确归属的局部整形规则。

## Files to modify

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/shared/mapperhelper/helper_test.go`
- `code/backend/internal/module/identity/application/commands/support.go`
- `code/backend/internal/module/identity/application/queries/support.go`
- `code/backend/internal/module/identity/application/queries/profile_service.go`
- `code/backend/internal/module/identity/application/commands/string_support.go`
- `code/backend/internal/module/identity/application/queries/string_support.go`
- `code/backend/internal/middleware/audit.go`
- `code/backend/internal/middleware/awd_readiness_audit.go`
- `code/backend/internal/middleware/string_support.go`
- `code/backend/internal/module/contest/domain/awd_readiness.go`

## After implementation

- `mapperhelper` 只剩仍在多模块复用的 `NormalizeOptionalString`
- `identity`、`middleware`、`contest/domain` 各自负责自己的 trimmed string 归一化语义
