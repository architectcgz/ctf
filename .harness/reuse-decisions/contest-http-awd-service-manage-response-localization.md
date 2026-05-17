# Reuse Decision

## Change type

backend / handler response localization

## Existing code searched

- `code/backend/internal/module/contest/api/http/awd_service_manage_handler.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/architecture_test.go`

## Similar implementations found

- 既有 handler 输出转换统一走 `contestRequestMapper`。
- command 输出收口已完成，handler 侧应只保留 bind/call/return，不内嵌字段拷贝。

## Decision

refactor_existing

## Reason

`ListContestAWDServices` 在 handler 文件内手写字段映射，不符合既有“handler 与业务之间由 mapper 负责转换”的约定。改为在 mapper 支撑层完成转换，handler 只做流程编排。

## Files to modify

- `.harness/reuse-decisions/contest-http-awd-service-manage-response-localization.md`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/request_mapper_awd_service_support.go`
- `code/backend/internal/module/contest/api/http/awd_service_manage_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`
