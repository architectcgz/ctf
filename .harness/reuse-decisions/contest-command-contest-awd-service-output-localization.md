# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`

## Similar implementations found

- 既有 contest command 输出收口都放在 `application/commands/contest_output.go`。
- 最近 AWD round / team service 输出都采用 command 本地输出类型，handler 接口直接依赖 command 输出。

## Decision

refactor_existing

## Reason

`CreateContestAWDService` 仍返回全局 `dto.ContestAWDServiceResp`。本刀将输出链路收口到 `commands.ContestAWDServiceResp`，保持字段语义不变；仅改 create 输出边界，不扩散 list/query 侧 dto。

## Files to modify

- `.harness/reuse-decisions/contest-command-contest-awd-service-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/awd_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit_test.go`
- `code/backend/internal/module/contest/architecture_test.go`
