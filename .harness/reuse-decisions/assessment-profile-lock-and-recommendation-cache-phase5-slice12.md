# Reuse Decision

## Change type

command / query / port / infrastructure / runtime / store

## Existing code searched

- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/ops/infrastructure/dashboard_state_store.go`
- `docs/plan/impl-plan/2026-05-13-ops-dashboard-state-store-phase5-slice11-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/ops/infrastructure/dashboard_state_store.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 assessment 的画像或推荐能力，而是 `profile_service.go` 里的画像锁和 `recommendation_service.go` 里的推荐缓存都直接挂在 application 层的 `*redis.Client` 上。最小正确方案不是造一个宽泛的 assessment cache service，也不是分两次留下半截 Redis concrete surface，而是把 assessment 剩余两条 allowlist 作为同一个 slice12 一次收口：在 `assessment/ports` 下新增两个窄 port，分别表达画像锁语义和推荐缓存语义，再由 `assessment/infrastructure` 用一个模块内 Redis adapter 文件统一承接 lock key、recommendation key、TTL、JSON 编解码和事件失效细节。

## Files to modify

- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/ports/state_store_context_contract_test.go`
- `code/backend/internal/module/assessment/infrastructure/state_store.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-assessment-profile-lock-and-recommendation-cache-phase5-slice12-implementation-plan.md`

## After implementation

- 如果 assessment 后续再新增 Redis 画像重建状态或推荐缓存失效场景，优先继续在这两个窄 port 上补方法，而不是把 lock key、TTL 和 recommendation cache 细节重新抬回 application
- 如果 phase5 之后要继续收口 GORM concrete surface，优先保持“按 use case surface 建窄 port”的模式，不要为了统一名字把 runtime 重新做成宽依赖桶
