# Reuse Decision

## Change type

job / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/contest/application/jobs/`
- `code/backend/internal/module/contest/ports/`
- `code/backend/internal/module/contest/infrastructure/`
- `code/backend/internal/pkg/redislock/`
- `docs/plan/impl-plan/2026-05-12-contest-status-updater-lock-phase5-slice4-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/contest/application/jobs/status_updater.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`

## Decision

refactor_existing

## Reason

`AWDRoundUpdater` 已经拥有完整的业务编排语义，当前 debt 在于 Redis 锁、round flag、current round 和 live status cache 的 concrete 细节散落在 application/jobs。最小正确方案不是新建第二个 AWD 子模块，也不是把所有 Redis 用法机械改成全局 helper，而是沿用 `contest` 已经在 phase 5 slice 3/4 落地的模式：保留 job owner，新增模块内窄 port，由 infrastructure 在 contest 模块内实现 Redis adapter。这样可以复用现有 `redislock`、runtime wiring 和 keepalive 行为，同时真正收缩 AWD round updater 的 allowlist。

## Files to modify

- `code/backend/internal/module/contest/ports/contest.go`
- `code/backend/internal/module/contest/ports/awd_round_state_context_contract_test.go`
- `code/backend/internal/module/contest/ports/status_update_lock_context_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_lock.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_sync.go`
- `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_check_sync.go`
- `code/backend/internal/module/contest/application/jobs/awd_check_writeback.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime_internal_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_testsupport_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-contest-awd-round-state-store-phase5-slice5-implementation-plan.md`

## After implementation

- 如果后续继续收口 AWD checker preview / scoreboard query 相关 Redis 依赖，可以继续复用“job-specific state store / query-specific cache store”的模式
- 如果别的 scheduler 也需要可续租分布式锁，优先复用这次抽出来的 generic scheduler lock lease，而不是重新把 `redislock.Lock` 暴露回 application
