# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/flag_service.go`
- `code/backend/internal/module/challenge/application/queries/flag_service.go`
- `code/backend/internal/module/challenge/application/commands/flag_service_test.go`
- `code/backend/internal/module/challenge/application/queries/flag_service_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`

## Decision

refactor_existing

## Reason

`challenge/application/commands/flag_service.go` 和 `challenge/application/queries/flag_service.go` 共享同一个 `ChallengeFlagRepository`，当前只有一条明确的 concrete 泄漏：challenge lookup 直接把 `gorm.ErrRecordNotFound` 映射成 `errcode.ErrNotFound`。

这类 not-found 语义已经足够窄，不需要重排 challenge repository 本体，也不需要把 `ChallengeWriteRepository` 或 `ChallengeReadRepository` 的全局 not-found 语义一起改掉。

最小正确方案是：

- 在 `challenge/ports` 增加一条只给 flag use case 使用的模块内 sentinel
- 新增一个窄 flag repository adapter，把 raw challenge repository 的 `gorm.ErrRecordNotFound` 映射成 sentinel
- commands / queries 两侧 flag service 都只消费 sentinel
- `challenge/runtime/module.go` 负责给 flag command/query service 注入 adapter

## Files to modify

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/application/commands/flag_service.go`
- `code/backend/internal/module/challenge/application/commands/flag_service_test.go`
- `code/backend/internal/module/challenge/application/queries/flag_service.go`
- `code/backend/internal/module/challenge/application/queries/flag_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/flag_repository.go`
- `code/backend/internal/module/challenge/infrastructure/flag_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-challenge-flag-not-found-contract-phase5-slice23-implementation-plan.md`

## After implementation

- `challenge/application/commands/flag_service.go -> gorm.io/gorm`
- `challenge/application/queries/flag_service.go -> gorm.io/gorm`

这两条例外应可一起删除。
