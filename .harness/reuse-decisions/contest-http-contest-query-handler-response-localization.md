# Reuse Decision

## Change type

backend / handler response localization

## Existing code searched

- `code/backend/internal/module/contest/api/http/contest_query_handler.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/request_mapper_contest_support.go`
- `code/backend/internal/module/contest/architecture_test.go`

## Similar implementations found

- 既有 handler 收口约定：handler 只负责 bind/call/return，转换逻辑进入 mapper 或 mapper support。
- AWD service manage 已按该模式从 handler 内字段拷贝迁移到 mapper support。

## Decision

refactor_existing

## Reason

`contest_query_handler` 仍直接依赖全局 `dto` 并在 handler 内组装分页响应；本刀改为模块内响应类型 + mapper support 转换，避免 handler 直接依赖 `dto` 和直接调用 mapper 细节。

## Files to modify

- `.harness/reuse-decisions/contest-http-contest-query-handler-response-localization.md`
- `code/backend/internal/module/contest/api/http/contest_query_handler.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/request_mapper_contest_support.go`
- `code/backend/internal/module/contest/architecture_test.go`
