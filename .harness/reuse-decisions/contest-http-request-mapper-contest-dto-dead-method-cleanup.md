# Reuse Decision

## Change type

backend / request mapper cleanup

## Existing code searched

- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/contest_query_handler.go`
- `code/backend/internal/module/contest/api/http/request_mapper_contest_support.go`

## Similar implementations found

- `contest_query_handler` 已切到 `ToContestCommandResp*` 链路。
- `request_mapper` 里旧的 `ToContestResp*` 已无调用。

## Decision

refactor_existing

## Reason

删除 `request_mapper` 中无调用的 `ToContestResp / Ptr / Resps`，减少 `dto` 映射冗余接口面，保持 mapper 与当前 handler 链路一致。

## Files to modify

- `.harness/reuse-decisions/contest-http-request-mapper-contest-dto-dead-method-cleanup.md`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
