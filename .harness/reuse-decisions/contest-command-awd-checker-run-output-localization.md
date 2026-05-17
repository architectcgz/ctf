# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_support.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`

## Similar implementations found

- 既有 AWD round / team service / contest service 输出都已迁到 `commands` 本地类型。
- run 命令输出同样属于 command 输出契约，适合同一收口模式。

## Decision

refactor_existing

## Reason

`RunCurrentRoundChecks` 与 `RunRoundChecks` 仍返回全局 `dto.AWDCheckerRunResp`。本刀将该链路收口到 `commands.AWDCheckerRunResp`，仅覆盖 run 输出边界，preview 输出单独后续收口。

## Files to modify

- `.harness/reuse-decisions/contest-command-awd-checker-run-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
