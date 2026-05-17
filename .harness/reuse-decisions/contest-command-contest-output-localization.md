# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/contest_create_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/api/http/handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`

## Similar implementations found

- 同模块 `submission` 已把命令输出从全局 `dto` 收口到 `submission_output.go`
- `contest_request_types.go` 已完成请求 DTO 模块内化，HTTP 层已具备模块 owner 结构

## Decision

refactor_existing

## Reason

`CreateContest/UpdateContest` 仍返回全局 `dto.ContestResp`，导致 command 输出 owner 漂到 `internal/dto`。本刀把输出收口到 `contest/application/commands/contest_output.go`，保持字段与行为不变，仅调整 owner 与依赖方向。

## Files to modify

- `.harness/reuse-decisions/contest-command-contest-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/contest_create_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
