# Reuse Decision

## Change type
service / query / port / infrastructure / composition

## Existing code searched
- `code/backend/internal/module/challenge/application/queries/`
- `code/backend/internal/module/challenge/ports/`
- `code/backend/internal/module/challenge/infrastructure/`
- `code/backend/internal/module/challenge/runtime/`
- `code/backend/internal/module/{practice,contest,assessment,ops}/`

## Similar implementations found
- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/runtime/infrastructure/proxy_ticket_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_scoreboard_cache.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`

## Decision
refactor_existing

## Reason
这刀不是新增一套缓存框架，而是把 `challenge` 里已经存在的 solved-count 缓存逻辑，从 application query service 里下沉到模块自己的 port + infrastructure adapter。仓库里已有两类可复用模式：一类是 `runtime/infrastructure/proxy_ticket_store.go` 这种“Redis 细节留在 infrastructure”的 adapter 写法，另一类是 `contest/infrastructure/awd_scoreboard_cache.go` 这种“用例只拿窄 cache 能力”的模块内缓存适配。最小正确方案是复用这两类模式，在 `challenge` 模块内补一个 solved-count cache port，而不是继续让 query service 持有 `*redis.Client`，也不是上升成跨模块通用缓存抽象。

## Files to modify
- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/ports/solved_count_cache_context_contract_test.go`
- `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`
- `docs/plan/impl-plan/2026-05-12-challenge-solved-count-cache-phase5-slice1-implementation-plan.md`

## After implementation
- 如果 phase 5 后续继续按“模块内 cache port + Redis adapter”模式收口其他 allowlist，可以把这个结论追加到 `harness/reuse/history.md` 或长期 reuse 索引。
