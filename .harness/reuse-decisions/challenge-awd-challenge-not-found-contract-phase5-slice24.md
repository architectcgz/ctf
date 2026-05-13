# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service_test.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/flag_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`

## Decision

refactor_existing

## Reason

`challenge/application/commands/awd_challenge_service.go` 和 `challenge/application/queries/awd_challenge_service.go` 共享同一个 AWD challenge repository not-found 语义：AWD challenge lookup 缺失时转成 `errcode.ErrNotFound`。

这类语义和上一刀 flag 很接近，都是“application 只关心业务 not-found，不关心底层是不是 GORM”。因此最小正确方案仍然是：

- 在 `challenge/ports` 增加一条 AWD challenge not-found sentinel
- 新增一个窄 AWD challenge repository adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 映射成 sentinel
- command/query service 只消费 sentinel
- `challenge/runtime/module.go` 负责把 adapter 注入 AWD command/query wiring

## Files to modify

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service_test.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-challenge-awd-challenge-not-found-contract-phase5-slice24-implementation-plan.md`

## After implementation

- `challenge/application/commands/awd_challenge_service.go -> gorm.io/gorm`
- `challenge/application/queries/awd_challenge_service.go -> gorm.io/gorm`

这两条例外应可一起删除。
