# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`

## Decision

refactor_existing

## Reason

`challenge/application/queries/challenge_service.go` 当前直接消费 `gorm.ErrRecordNotFound`，但这条 query 线已经有两个稳定且必须保留的 application 语义：

- `GetChallenge` 的 challenge lookup not-found -> `errcode.ErrChallengeNotFound`
- `GetPublishedChallenge` 的同一类 lookup not-found -> `errcode.ErrNotFound`

最小正确改法不是改 raw `Repository`，也不是把 `GetChallenge` / `GetPublishedChallenge` 压成同一个错误码，而是沿用 challenge 模块已有模式：

- 在 `challenge/ports` 增加 query lookup 的 sentinel
- 新增窄 `ChallengeQueryRepository` adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 收口成这个 sentinel
- `challenge_service.go` 只做 sentinel -> 既有 `errcode` 映射
- `runtime/module.go` 负责给 query service 注入 adapter

这次不扩到 image / topology / command / shared allowlist / docs。

## Files to modify

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `.harness/reuse-decisions/challenge-query-not-found-contract-phase5-slice31.md`
- `docs/plan/impl-plan/2026-05-13-challenge-query-not-found-contract-phase5-slice31-implementation-plan.md`

## After implementation

- `challenge/application/queries/challenge_service.go -> gorm.io/gorm`

这条 application concrete allowlist 依赖应可由主线程在共享 allowlist 收口时删除。
