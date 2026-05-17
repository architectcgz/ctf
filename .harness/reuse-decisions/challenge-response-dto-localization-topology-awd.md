# Reuse Decision

## Change type

api / response mapper / dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/{topology_handler.go,awd_challenge_handler.go}`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`
- `code/backend/internal/dto/{topology.go,awd_challenge.go,awd_challenge_import.go,challenge_import.go}`
- `docs/plan/impl-plan/2026-05-17-challenge-response-dto-localization-image-tag-flag-implementation-plan.md`

## Similar implementations found

- `challenge` 的 image/tag/flag 已采用“本地 response type + response mapper”模式
- `ops/api/http/notification_response_mapper.go` 已采用 mapper 收口 handler 输出

## Decision

refactor_existing

## Reason

`topology` 与 `awd_challenge` handler 仍直接返回全局 `dto`。本切片沿用已验证的 response mapper 模式，把这两组响应收回 `challenge/api/http`，不改变外部 JSON 契约。

## Files to modify

- `.harness/reuse-decisions/challenge-response-dto-localization-topology-awd.md`
- `docs/plan/impl-plan/2026-05-17-challenge-response-dto-localization-topology-awd-implementation-plan.md`
- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/{topology_handler.go,awd_challenge_handler.go}`
- `docs/architecture/backend/04-api-design.md`
