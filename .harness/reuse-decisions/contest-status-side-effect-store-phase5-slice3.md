# Reuse Decision

## Change type
service / job / port / infrastructure / runtime

## Existing code searched
- `code/backend/internal/module/contest/application/`
- `code/backend/internal/module/contest/ports/`
- `code/backend/internal/module/contest/infrastructure/`
- `code/backend/internal/module/contest/runtime/`

## Similar implementations found
- `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/contest/ports/scoreboard.go`
- `code/backend/internal/module/contest/ports/realtime.go`
- `code/backend/internal/module/contest/ports/participation.go`
- `code/backend/internal/module/contest/infrastructure/awd_scoreboard_cache.go`
- `code/backend/internal/module/runtime/infrastructure/proxy_ticket_store.go`
- `code/backend/internal/module/contest/application/jobs/status_updater.go`

## Decision
refactor_existing

## Reason
这刀不需要新建通用缓存框架，而是把 `contest` 状态迁移副作用里现有的 Redis 读写，从 application/statusmachine 下沉到模块自己的 side-effect store port。port 设计复用 `contest/ports/*.go` 现有模式：像 `team.go`、`submission.go`、`scoreboard.go`、`realtime.go`、`participation.go` 一样，把能力收口成模块内小接口，保持 `ctx` 在首位，不把 Redis client 或 key 暴露到 application。infrastructure 侧则复用两个已存在模式：`runtime/infrastructure/proxy_ticket_store.go` 这种“Redis 细节停在 infrastructure”的 adapter，以及 `contest/infrastructure/awd_scoreboard_cache.go` 这种“contest 模块内 cache 能力由 infrastructure 实现”的组织方式。最小正确方案是沿这两类现有模式，新建一个窄 `ContestStatusSideEffectStore`，承接 frozen snapshot 和 runtime state cleanup，而不是继续让 `ContestService` 或 `SideEffectRunner` 直接拿 Redis client。

## Files to modify
- `code/backend/internal/module/contest/ports/contest.go`
- `code/backend/internal/module/contest/ports/status_side_effect_store_context_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
- `code/backend/internal/module/contest/application/commands/contest_service.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_service.go`
- `code/backend/internal/module/contest/application/jobs/status_updater.go`
- `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_service_test.go`
- `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- `code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-12-contest-status-side-effect-store-phase5-slice3-implementation-plan.md`

## After implementation
- 如果后续 phase 5 要继续清理 `contest_service.go`、`scoreboard_admin_service.go`、`status_updater.go` 这组 Redis 依赖，可以复用这次“先抽 side-effect store，再收构造注入”的模式。
