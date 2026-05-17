# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_result_goverter.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`

## Similar implementations found

- AWD round、team service、checker run 输出都已迁到 command 本地类型。
- preview 输出同样属于 command 输出边界，适配相同收口规则。

## Decision

refactor_existing

## Reason

`PreviewChecker` 仍在 `run_commands` 中返回全局 `dto.AWDCheckerPreviewResp`，导致该文件继续依赖 `dto`。本刀把 preview 输出链路改为 `commands.AWDCheckerPreviewResp`，并同步 token 存储映射和 handler 接口。

## Files to modify

- `.harness/reuse-decisions/contest-command-awd-checker-preview-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_result_goverter.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_result_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
