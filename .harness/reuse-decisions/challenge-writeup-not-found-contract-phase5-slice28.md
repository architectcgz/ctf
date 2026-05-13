# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_topology_service_test.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/application/queries/image_service.go`
- `code/backend/internal/module/challenge/application/commands/flag_service.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`

## Decision

refactor_existing

## Reason

`challenge` 的 writeup command/query 共享同一组 raw repository lookup，但它们需要分支的 not-found 语义并不相同：

- challenge lookup not-found -> `errcode.ErrChallengeNotFound`
- official/released/teacher submission writeup not-found -> `errcode.ErrNotFound`
- user-challenge submission writeup not-found -> 某些用例返回 `nil`
- requester user not-found -> `errcode.ErrUnauthorized`

因此不适合把所有 not-found 混成一条统一 sentinel，也不需要重写 raw `Repository` 的 GORM 行为。最小正确方案是：

- 在 `challenge/ports` 增加按语义拆分的 writeup sentinel
- 新增一个窄 writeup repository adapter，只负责把 raw writeup repository 的 `gorm.ErrRecordNotFound` 映射成这些 sentinel
- command/query service 只消费 `challenge/ports` sentinel，不再直接 import GORM
- `challenge/runtime/module.go` 负责把 adapter 注入 writeup command/query wiring

## Files to modify

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_topology_service_test.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `docs/plan/impl-plan/2026-05-13-challenge-writeup-not-found-contract-phase5-slice28-implementation-plan.md`

## After implementation

- `challenge/application/commands/writeup_service.go -> gorm.io/gorm`
- `challenge/application/queries/writeup_service.go -> gorm.io/gorm`

这两条例外应可由主线程在共享 allowlist 收口时删除。
