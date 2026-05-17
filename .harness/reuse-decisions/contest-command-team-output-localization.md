# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/team_handler.go`

## Similar implementations found

- 同模块已完成 `submission`、`contest create/update` 命令输出收口
- `team` 命令输出与上述两条链路结构一致，适合同样的本地 output owner 模式

## Decision

refactor_existing

## Reason

`team create/join` 仍返回全局 `dto.TeamResp`，让 command 输出 owner 依赖 `internal/dto`。本刀把输出收回 `contest/application/commands`，保持字段与行为不变，只调整 owner 和依赖方向。

## Files to modify

- `.harness/reuse-decisions/contest-command-team-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/team_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`
