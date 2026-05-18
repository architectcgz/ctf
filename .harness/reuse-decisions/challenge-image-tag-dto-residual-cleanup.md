# Reuse Decision

## Change type

contract / api / app-test / dto cleanup

## Existing code searched

- `code/backend/internal/dto/{image.go,tag.go}`
- `code/backend/internal/module/challenge/api/http/{challenge_request_types.go,challenge_response_types.go}`
- `code/backend/internal/module/challenge/contracts/{image.go,tag_flag.go,page_result.go}`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Similar implementations found

- `challenge/api/http` 已持有 `CreateImageReq`、`UpdateImageReq`、`ImageQuery`、`CreateTagReq`、`AttachTagsReq`、`TagQuery`
- `challenge/api/http` 已持有 `ImageResp`、`TagResp`
- `challenge/contracts` 已持有模块内 `ImageResp`、`TagResp` 与 `PageResult`
- 历史收口记录：`challenge-image-response-contract-localization.md`、`challenge-tag-flag-response-contract-localization.md`

## Decision

refactor_existing

## Reason

`internal/dto/tag.go` 已经没有任何消费方，`internal/dto/image.go` 只剩 app 集成测试还在借 `dto.ImageResp`。owner 已经非常明确，最小正确方案不是再造一层中转，而是直接复用 `challenge` 模块现有的 request / response DTO owner，并删除残留全局文件。

## Files to modify

- `.harness/reuse-decisions/challenge-image-tag-dto-residual-cleanup.md`
- `docs/plan/impl-plan/2026-05-18-challenge-image-tag-dto-residual-cleanup-implementation-plan.md`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/image.go`
- `code/backend/internal/dto/tag.go`

## After implementation

- app 集成测试不再依赖 `dto.ImageResp`
- `internal/dto/image.go`、`internal/dto/tag.go` 删除
- challenge image/tag request / response DTO 只保留模块 owner
