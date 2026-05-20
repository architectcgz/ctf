# Reuse Decision

## Change type
shared helper final localization

## Existing code searched

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/auth/application/commands/support.go`
- `code/backend/internal/module/ops/application/**/*.go`
- `code/backend/internal/**/*.go`

## Similar implementations found

- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/auth/application/commands/support.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/queries/notification_service.go`

## Decision
refactor_existing

## Reason

`NormalizeOptionalString` 已经不再是稳定 shared kernel，当前只剩两类明确 owner：

- `auth/api/http` 与 `runtime/api/http`：`UserAgent` 可空化
- `auth/application/commands` 与 `ops/application`：响应字段可空化

这些整形规则已经收敛到局部模块和局部层级。继续保留 `internal/shared/mapperhelper` 只会留下一个没有真实共享语义的空壳包。

## Files to modify

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/shared/mapperhelper/helper_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/string_support.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/runtime/api/http/string_support.go`
- `code/backend/internal/module/auth/application/commands/support.go`
- `code/backend/internal/module/auth/application/commands/string_support.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/string_support.go`
- `code/backend/internal/module/ops/application/queries/notification_service.go`
- `code/backend/internal/module/ops/application/queries/string_support.go`

## After implementation

- `internal/shared/mapperhelper` 不再承载任何运行时代码
- 最后一个 mapper helper 也回到明确 owner，shared 包清空
