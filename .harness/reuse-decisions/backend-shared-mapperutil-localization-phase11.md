# Reuse Decision

## Change type
shared helper localization

## Existing code searched

- `code/backend/internal/shared/mapperutil/strings.go`
- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/**/*.go`

## Similar implementations found

- `code/backend/internal/shared/mapperhelper/helper.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`

## Decision
refactor_existing

## Reason

`internal/shared/mapperutil` 当前只剩一个函数、一个调用点：

- `NormalizeOptionalString(*string) *string`
- 仅被 `assessment/application/commands/report_service.go` 使用

这不是稳定 shared kernel，而是 assessment 报告导出在 failed 状态下对 `ErrorMsg` 的出参整形。继续保留全局 `mapperutil` 只会增加一个没有复用价值的共享入口。

这次直接把该逻辑收回 assessment 命令层，保留行为不变：

- `nil` 保持 `nil`
- 空白字符串归一为 `nil`
- 非空字符串返回裁剪后的副本

## Files to modify

- `code/backend/internal/shared/mapperutil/strings.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`

## After implementation

- `internal/shared` 只保留真实复用的 shared-kernel 包。
- assessment 报告导出对失败消息的整形 owner 明确落在 assessment 模块内。
