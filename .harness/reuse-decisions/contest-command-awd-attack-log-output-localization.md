# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_submit_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_response_support.go`
- `code/backend/internal/module/contest/application/commands/awd_response_mappers.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`

## Similar implementations found

- 同模块 `submission`、`contest`、`team`、`challenge`、`participation` 命令输出已逐步收口到 `contest/application/commands`。
- AWD attack log 输出映射沿用同一 mapper + wrapper 模式。

## Decision

refactor_existing

## Reason

`CreateAttackLog/SubmitAttack` 仍返回全局 `dto.AWDAttackLogResp`。本刀把 attack log 输出收口到模块内 output，并同步 handler 接口与测试桩签名，不改外部 JSON 字段与业务语义。

## Files to modify

- `.harness/reuse-decisions/contest-command-awd-attack-log-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_submit_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_response_support.go`
- `code/backend/internal/module/contest/application/commands/awd_response_mappers.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
