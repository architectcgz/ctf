# Reuse Decision

## Change type

contract / query / command / handler / mapper dto localization

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/image_service.go`
- `code/backend/internal/module/challenge/application/queries/image_service.go`
- `code/backend/internal/module/challenge/api/http/image_handler.go`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`
- `code/backend/internal/module/challenge/domain/{mappers.go,response_mapper_goverter*.go}`

## Similar implementations found

- `writeup` 分页结果已改为 `challenge/contracts.PageResult`
- `challenge` 主 handler 输出已通过 `api/http` mapper `to...` 收口
- 其他模块（assessment/practice）已有 contracts 持有模块响应类型的模式

## Decision

refactor_existing

## Reason

`image` 链路仍在 command/query/handler 边界暴露 `internal/dto.ImageResp` 与 `dto.PageResult`。本次将其收口到 `challenge/contracts`，让 challenge 模块内边界只依赖模块 contracts；HTTP 输出继续经 `to...` 映射到 `api/http` 响应类型。

## Files to modify

- `.harness/reuse-decisions/challenge-image-response-contract-localization.md`
- `code/backend/internal/module/challenge/contracts/image.go`
- `code/backend/internal/module/challenge/domain/mappers.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/commands/image_service.go`
- `code/backend/internal/module/challenge/application/queries/image_service.go`
- `code/backend/internal/module/challenge/api/http/image_handler.go`
- `code/backend/internal/module/challenge/api/http/image_handler_test.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
