# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_round_admin_commands.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`

## Similar implementations found

- 近期 `contest` command 输出收口都采用模块内 output 类型，HTTP 接口直接依赖 command output。
- `CreateRound` 输出链路与上述模式一致。

## Decision

refactor_existing

## Reason

`CreateRound` 仍返回全局 `dto.AWDRoundResp`。本刀收口成 `commands.AWDRoundResp`，保持字段与行为不变；`AWDCheckerRunResp` 仍保留原状，避免本刀扩散。

## Files to modify

- `.harness/reuse-decisions/contest-command-awd-round-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/awd_round_admin_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
