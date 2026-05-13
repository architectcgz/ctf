# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/queries/image_service.go`
- `code/backend/internal/module/challenge/application/queries/image_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`

## Decision

refactor_existing

## Reason

`challenge/application/queries/image_service.go` 只有一条单一 concrete：`GetImage` 直接把 repo 的 `gorm.ErrRecordNotFound` 映射成 `errcode.ErrImageNotFound`。这类 not-found 语义已经足够明确，不需要拉动 image command service、runtime builder 或 repository 全量重排。

最小正确方案是：

- 在 `challenge/ports` 增加一条模块内 sentinel 表达 image not-found
- 新增一个窄 image-query adapter，把 raw image repository 的 `gorm.ErrRecordNotFound` 映射成 sentinel
- `image_service.go` 只消费 sentinel，`challenge/runtime/module.go` 负责注入 adapter

## Files to modify

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/application/queries/image_service.go`
- `code/backend/internal/module/challenge/application/queries/image_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-challenge-image-query-not-found-contract-phase5-slice22-implementation-plan.md`

## After implementation

- `challenge/application/queries/image_service.go -> gorm.io/gorm` 这条例外应可删除
- `challenge` 模块会有第一条 image query 相关的模块内错误契约，后续若继续收 image command / topology / writeup 的 not-found，可复用同样模式
