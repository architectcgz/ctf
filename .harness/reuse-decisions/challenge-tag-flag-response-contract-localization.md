# Reuse Decision

## Change type

contract / command / query / handler / mapper dto localization

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/tag_service.go`
- `code/backend/internal/module/challenge/application/queries/{tag_service.go,flag_service.go,response_mapper_goverter*.go}`
- `code/backend/internal/module/challenge/api/http/{tag_handler.go,flag_handler.go,response_mapper*.go}`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter*.go`

## Similar implementations found

- `image` 链路已收口到 `challenge/contracts.ImageResp`
- `writeup` 分页已收口到 `challenge/contracts.PageResult`
- `api/http` 输出已统一通过 `to...` helper 做响应映射

## Decision

refactor_existing

## Reason

`tag` 与 `flag` 仍在 command/query/handler 边界暴露全局 `dto.TagResp`、`dto.FlagResp`。本次将两者收口到 `challenge/contracts`，并保留 HTTP 层 mapper 的统一转换入口。

## Files to modify

- `.harness/reuse-decisions/challenge-tag-flag-response-contract-localization.md`
- `code/backend/internal/module/challenge/contracts/tag_flag.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/commands/tag_service.go`
- `code/backend/internal/module/challenge/application/commands/tag_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/tag_service.go`
- `code/backend/internal/module/challenge/application/queries/tag_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/flag_service.go`
- `code/backend/internal/module/challenge/application/queries/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/api/http/tag_handler.go`
- `code/backend/internal/module/challenge/api/http/flag_handler.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
