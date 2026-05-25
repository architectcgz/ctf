# Reuse Decision

## Change type

api / response mapper / dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/{tag_handler.go,image_handler.go,flag_handler.go}`
- `code/backend/internal/module/challenge/api/http/request_mapper*.go`
- `code/backend/internal/module/ops/api/http/notification_response_mapper*.go`
- `code/backend/internal/dto/{tag.go,image.go,challenge.go}`

## Similar implementations found

- `ops/api/http` 已采用“模块内 response type + response mapper”模式
- `auth/api/http` 已采用 request/response mapper 分离模式

## Decision

refactor_existing

## Reason

`challenge` 的 image/tag/flag handler 仍直接把全局 `dto` 透传到 HTTP 层。先在 `challenge/api/http` 落一层 response mapper，把输出转换收口到模块边界，再继续扩展到 challenge/topology/awd 其他响应。

## Files to modify

- `.harness/reuse-decisions/challenge-response-dto-localization-image-tag-flag.md`
- `docs/plan/archive/impl-plan/2026-05-17-challenge-response-dto-localization-image-tag-flag-implementation-plan.md`
- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_assign.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/{tag_handler.go,image_handler.go,flag_handler.go}`
