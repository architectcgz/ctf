# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/image_service.go`
- `code/backend/internal/module/challenge/application/commands/image_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`

## Decision

refactor_existing

## Reason

`challenge/application/commands/image_service.go` 当前仍直接消费 `gorm.ErrRecordNotFound`，但 image command 这条线只需要一个稳定的模块语义：image lookup not-found。

最小正确改法不是改 raw `ImageRepository` 的全局行为，也不是让 image build/query 一起跟着改，而是沿用 challenge 模块已经成熟的 adapter 模式：

- 复用已有 `challenge/ports.ErrChallengeImageNotFound`
- 新增窄 `ImageCommandRepository` adapter，把 command lookup 上的 `gorm.ErrRecordNotFound` 收口成这个 sentinel
- `image_service.go` 只消费 sentinel，再继续映射成既有 `errcode.ErrImageNotFound` 或“允许创建”
- `runtime/module.go` 只给 image command service 注入 adapter，image query / image build 继续保留原有注入面

这次不扩到 shared allowlist、事实源文档、image build transaction surface 或 challenge import / package revision。

## Files to modify

- `code/backend/internal/module/challenge/application/commands/image_service.go`
- `code/backend/internal/module/challenge/application/commands/image_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_command_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_command_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `.harness/reuse-decisions/challenge-image-command-not-found-contract-phase5-slice33.md`
- `docs/plan/impl-plan/2026-05-13-challenge-image-command-not-found-contract-phase5-slice33-implementation-plan.md`

## After implementation

- `challenge/application/commands/image_service.go -> gorm.io/gorm`

这条 application concrete allowlist 依赖应由主线程在 shared cleanup 时删除。
