# Reuse Decision

## Change type
test / mapper / model localization follow-up

## Existing code searched
- `code/backend/internal/module/challenge/**/*.go`
- `.harness/reuse-decisions/challenge-image-build-job-model-localization.md`
- `.harness/reuse-decisions/challenge-image-response-contract-localization.md`

## Similar implementations found
- `code/backend/internal/module/challenge/application/commands/image_service_context_test.go`
- `code/backend/internal/module/challenge/domain/mappers.go`

## Decision
refactor_existing

## Reason
`Image` 类型已经收口到 `challenge/entity` 后，少量上下文契约测试和 mapper 单测仍保留旧引用。这里不新增新的共享抽象，直接跟随现有 image 收口链路做最小改动，保持行为不变，仅修正类型归属。

## Files to modify
- `code/backend/internal/module/challenge/application/commands/image_service_context_test.go`
- `code/backend/internal/module/challenge/domain/image_mapper_test.go`

## After implementation
- 复用现有 image 收口模式，不新增全局共享层。
