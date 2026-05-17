# Reuse Decision

## Change type

backend / request mapper cleanup

## Existing code searched

- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/request_mapper_awd_service_support.go`

## Similar implementations found

- AWD service manage 列表响应已改用 `contestcmd` 响应类型与 command-preview 映射方法。
- 原 `dto` 映射方法在当前代码中已无调用。

## Decision

refactor_existing

## Reason

移除 `request_mapper` 中无调用的 AWD `dto` 方法声明，避免无意义的 `dto` 转换继续留在 mapper 入口，保持转换面与当前链路一致。

## Files to modify

- `.harness/reuse-decisions/contest-http-request-mapper-awd-dto-dead-method-cleanup.md`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
