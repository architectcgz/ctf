# Reuse Decision

## Change type

backend / request mapper interface slimming

## Existing code searched

- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/*_handler.go`
- `code/backend/internal/module/contest/api/http/request_mapper_*_support.go`

## Similar implementations found

- handler 与 support 目前只调用 `contestRequestMapper` 的部分方法。
- 大量 `dto` 方向单项方法仅作为历史遗留，当前无调用。

## Decision

refactor_existing

## Reason

按当前调用面收缩 `ContestRequestMapper` 接口，仅保留被 handler/support 实际使用的方法，减少 `dto` 相关历史映射暴露面。

## Files to modify

- `.harness/reuse-decisions/contest-http-request-mapper-interface-slimming.md`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
