# Reuse Decision

## Change type

contract / command / query / handler / mapper dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/awd_challenge_handler*.go`
- `code/backend/internal/module/challenge/application/{commands,queries}/awd_challenge*.go`
- `code/backend/internal/module/challenge/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/challenge/domain/{mappers.go,response_mapper_goverter*.go}`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`

## Similar implementations found

- `topology`、`image`、`tag/flag`、`writeup` 已改为模块 contracts 输出
- `api/http` 层保持 `to...` helper 统一映射

## Decision

refactor_existing

## Reason

`awd` handler 与 command/query 边界仍暴露全局 `dto.AWDChallenge*` 与 `dto.AWDChallengeImportPreviewResp`。本次将 awd 响应类型收口到 `challenge/contracts`，统一 challenge 模块内部输出 owner。

## Files to modify

- `.harness/reuse-decisions/challenge-awd-response-contract-localization.md`
- `code/backend/internal/module/challenge/contracts/awd_challenge.go`
- `code/backend/internal/module/challenge/domain/mappers.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service.go`
- `code/backend/internal/module/challenge/api/http/awd_challenge_handler.go`
- `code/backend/internal/module/challenge/api/http/awd_challenge_handler_test.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
