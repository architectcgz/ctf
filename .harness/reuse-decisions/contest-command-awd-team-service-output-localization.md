# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_service_upsert_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_service_upsert_response_support.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`

## Similar implementations found

- `contest` 模块近期 command 输出收口统一落在 `application/commands/contest_output.go`。
- `CreateRound` 与 `CreateAttackLog` 已采用 command 本地输出类型，HTTP 接口直接依赖 command 输出。

## Decision

refactor_existing

## Reason

`UpsertServiceCheck` 仍返回全局 `dto.AWDTeamServiceResp`。本刀将该链路收口到 `commands.AWDTeamServiceResp`，保持字段语义不变，仅限 upsert 输出边界，不扩散到 `RunRoundChecks` 等未收口链路。

## Files to modify

- `.harness/reuse-decisions/contest-command-awd-team-service-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/awd_service_upsert_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_service_upsert_response_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
