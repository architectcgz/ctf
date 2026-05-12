# Reuse Decision

## Change type
job / port / infrastructure / runtime

## Existing code searched
- `code/backend/internal/module/contest/application/jobs/`
- `code/backend/internal/module/contest/ports/`
- `code/backend/internal/module/contest/infrastructure/`
- `code/backend/internal/pkg/redislock/`

## Similar implementations found
- `code/backend/internal/module/contest/application/jobs/status_updater.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/contest/ports/scoreboard.go`
- `code/backend/internal/module/contest/ports/realtime.go`
- `code/backend/internal/module/contest/ports/participation.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/pkg/redislock/lock.go`

## Decision
refactor_existing

## Reason
这刀不需要新建通用调度框架，而是把 `StatusUpdater` 已有的 Redis 锁获取细节，从 application/jobs 下沉到模块自己的 lock store port。port 设计复用 `contest/ports/*.go` 现有模式：把能力表达成模块内窄接口，保持 `ctx` 在首位，不把 Redis client / key / `redislock` 类型暴露到 application。job 侧 keepalive 与“失锁即停”语义复用现有 `lock_keepalive.go` 和 `awd_round_scheduler_runtime.go` 的模式；infrastructure 侧复用刚落地的 `status_side_effect_store.go` 组织方式，在 contest 模块内部实现一个 Redis adapter，并继续复用 `internal/pkg/redislock/lock.go` 作为底层锁机制。最小正确方案是抽出 `StatusUpdater` 自己的 lock store / lease，而不是直接改造整个 AWD scheduler 锁体系。

## Files to modify
- `code/backend/internal/module/contest/ports/contest.go`
- `code/backend/internal/module/contest/ports/status_update_lock_context_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- `code/backend/internal/module/contest/application/jobs/status_updater.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- `code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go`
- `code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-12-contest-status-updater-lock-phase5-slice4-implementation-plan.md`

## After implementation
- 如果后续要继续收口 `AWDRoundUpdater` 的调度锁，可以复用这次“先抽 job-specific lock store，再保留 application 里的 keepalive 编排”的模式。
