# Reuse Decision

## Change type

command / query / port / infrastructure / runtime / store

## Existing code searched

- `code/backend/internal/module/practice/application/commands/score_service.go`
- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/`
- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- `docs/plan/impl-plan/2026-05-13-contest-scoreboard-state-store-phase5-slice6-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- `code/backend/internal/module/contest/ports/scoreboard_state_context_contract_test.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 practice score 能力，而是用户计分锁、用户得分缓存和排行榜 sorted-set 细节直接散落在 `practice/application/commands/score_service.go` 与 `practice/application/queries/score_service.go`。最小正确方案不是新建并行 score 模块，也不是继续让 command / query 各自保留 Redis helper，而是沿用 phase5 已验证的 port + infrastructure store 模式：保留 practice score command/query 的业务编排 owner，在 `practice/ports` 下新增窄 `PracticeScoreStateStore`，由 `practice/infrastructure` 统一承接锁、缓存和排行榜状态细节。

## Files to modify

- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/ports/score_state_context_contract_test.go`
- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/practice/application/commands/score_service.go`
- `code/backend/internal/module/practice/application/commands/score_service_test.go`
- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/application/queries/score_service_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-score-state-store-phase5-slice9-implementation-plan.md`

## After implementation

- 如果后续继续收口 `practice/application/commands/service.go` 上的 flag submit 限流 Redis 依赖，应单独抽成 submission rate-limit store，而不是把 score store 扩成宽泛的 practice cache bucket
- 如果 practice 后续还需要更多用户得分相关缓存读写，优先复用这次 score state store，而不是重新把 `cache.UserScoreKey`、`cache.RankingKey` 或 score lock 细节抬回 application
