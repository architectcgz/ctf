# Reuse Decision

## Change type

contract / query / handler / dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/writeup_handler.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/contracts/writeup.go`
- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`

## Similar implementations found

- `challenge/api/http/challenge_response_types.go` 已有模块内分页壳用于 handler 输出
- `assessment`、`practice` 已采用“模块 contracts 持有对外响应类型，避免 handler 直接依赖全局 dto”模式

## Decision

refactor_existing

## Reason

`writeup` 链路在 query service 与 handler 接口仍暴露 `internal/dto.PageResult`。本次将分页壳下沉到 `challenge/contracts`，保证 challenge HTTP/application 边界只依赖模块 contracts，而不是全局 DTO 桶。

## Files to modify

- `.harness/reuse-decisions/challenge-writeup-page-result-localization.md`
- `code/backend/internal/module/challenge/contracts/page_result.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/api/http/writeup_handler.go`
