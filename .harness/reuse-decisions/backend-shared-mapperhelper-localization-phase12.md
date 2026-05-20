# Reuse Decision

## Change type
shared helper localization

## Existing code searched

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/shared/mapperhelper/helper_test.go`
- `code/backend/internal/module/identity/application/**/*.go`
- `code/backend/internal/module/teaching_query/application/queries/*.go`
- `code/backend/internal/**/*.go`

## Similar implementations found

- `code/backend/internal/module/identity/application/commands/support.go`
- `code/backend/internal/module/identity/application/queries/support.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`

## Decision
refactor_existing

## Reason

`internal/shared/mapperhelper` 里目前混着两类东西：

- 仍在跨模块复用的字符串归一化：`NormalizeOptionalString`、`NormalizeOptionalTrimmedString`
- 已经有明确 owner 的局部整形：`SingleString`、`CopyTimeToPtr`、`NonNilSlice`

其中：

- `NonNilSlice` 只剩 `teaching_query` 使用
- `SingleString`、`CopyTimeToPtr` 只剩 `identity` 使用

继续把这三类放在全局 shared，会让 shared 包继续承担已经内聚回模块内的实现细节。这次先把有明确 owner 的 helper 收回模块内，保留真正跨模块的字符串归一化函数。

## Files to modify

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/shared/mapperhelper/helper_test.go`
- `code/backend/internal/module/identity/application/commands/support.go`
- `code/backend/internal/module/identity/application/queries/support.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/overview_service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/slice_support.go`

## After implementation

- `mapperhelper` 只保留仍有跨模块复用价值的字符串归一化能力。
- `identity` 自己负责 admin user 响应整形细节。
- `teaching_query` 自己负责查询响应里的空切片归一化语义。
