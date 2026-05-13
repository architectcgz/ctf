# Reuse Decision

## Change type

query / command / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/contest/application/queries/scoreboard_*.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_*.go`
- `code/backend/internal/module/contest/ports/`
- `code/backend/internal/module/contest/infrastructure/`
- `docs/plan/impl-plan/2026-05-13-contest-awd-round-state-store-phase5-slice5-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_scoreboard_cache.go`
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_list_query.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_freeze_commands.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 scoreboard 能力，而是 live scoreboard、frozen snapshot、team rank、分数增量和 rebuild 的 Redis 细节同时散落在 `scoreboard_service` 与 `scoreboard_admin_service`。最小正确方案不是再新建并行 scoreboard 模块，也不是让 query / command 继续各自保留一套 Redis helper，而是沿用 phase5 既有做法：保留 scoreboard query/admin 的业务编排 owner，在 `contest/ports` 下新增窄 scoreboard state store，由 `contest/infrastructure` 实现统一 Redis adapter，并让 `status_side_effect_store` 复用同一组 frozen snapshot helper。

## Files to modify

- `code/backend/internal/module/contest/ports/contest.go`
- `code/backend/internal/module/contest/ports/scoreboard_state_context_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_service.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_support.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_rank_query.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_list_support.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_service_test.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_service.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_score_commands.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-contest-scoreboard-state-store-phase5-slice6-implementation-plan.md`

## After implementation

- 如果后续继续收口 `AWDService` 或 `submission_service` 的 scoreboard Redis 依赖，可以继续复用这次 scoreboard state store，而不是再让 application 自己碰 sorted-set 细节
- 如果 frozen scoreboard snapshot 以后还要被更多 status side effect 使用，优先继续复用这次 infrastructure helper，而不是再复制 key/pipeline 逻辑
