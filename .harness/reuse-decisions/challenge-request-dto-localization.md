# Reuse Decision

## Change type

api / request mapper / dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/*.go`
- `code/backend/internal/module/challenge/api/http/request_mapper*.go`
- `code/backend/internal/dto/{challenge.go,awd_challenge.go,tag.go,image.go,topology.go}`
- `docs/plan/impl-plan/2026-05-17-challenge-contest-instance-awd-dto-localization-next-batch-plan.md`

## Similar implementations found

- `contest/api/http` request DTO 已完成模块内化并通过 mapper 收口
- `teaching_query/api/http` request DTO 已按“本地类型 + mapper 转换”模式收口

## Decision

refactor_existing

## Reason

`challenge` HTTP 入参与查询参数长期挂在全局 `internal/dto`，导致 handler 与模块边界 owner 不清晰。先把 request/query 类型收回 `challenge/api/http`，保持外部 API 契约不变，再继续后续 response/output 切片。

## Files to modify

- `.harness/reuse-decisions/challenge-request-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-challenge-request-dto-localization-implementation-plan.md`
- `code/backend/internal/module/challenge/api/http/challenge_request_types.go`
- `code/backend/internal/module/challenge/api/http/request_mapper.go`
- `code/backend/internal/module/challenge/api/http/request_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/{handler.go,tag_handler.go,image_handler.go,topology_handler.go,flag_handler.go,awd_challenge_handler.go}`
- `docs/architecture/backend/04-api-design.md`
