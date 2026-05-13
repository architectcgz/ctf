# Reuse Decision

## Change type

query / port / infrastructure / runtime / store

## Existing code searched

- `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service_test.go`
- `code/backend/internal/module/ops/ports/dashboard.go`
- `code/backend/internal/module/ops/runtime/module.go`
- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `docs/plan/impl-plan/2026-05-13-practice-score-state-store-phase5-slice9-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 dashboard 查询能力，而是 dashboard cache 和在线会话计数的 Redis 细节同时散落在 `ops/application/queries/dashboard_service.go`。最小正确方案不是新建一层通用缓存服务，也不是把 auth session 结构直接抬成跨模块 contract，而是沿用 phase5 既有模式：保留 `DashboardService` 的查询编排 owner，在 `ops/ports` 下新增窄 `DashboardStateStore`，由 `ops/infrastructure` 统一承接缓存 key、会话扫描、JSON 载荷和 Redis `Get/Set/Scan/MGet` 细节。

## Files to modify

- `code/backend/internal/module/ops/ports/dashboard.go`
- `code/backend/internal/module/ops/ports/dashboard_state_context_contract_test.go`
- `code/backend/internal/module/ops/infrastructure/dashboard_state_store.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service_test.go`
- `code/backend/internal/module/ops/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-ops-dashboard-state-store-phase5-slice11-implementation-plan.md`

## After implementation

- 如果后续 `assessment` 还要收口 recommendation / profile 的 Redis cache，优先复用这次“按 use case 建窄 state store”的模式，而不是回到全局通用 Redis helper
- 如果 `ops` dashboard 以后还要新增更多 Redis runtime state，优先扩展同一个 `DashboardStateStore` 的窄方法，而不是把 key / JSON 细节重新抬回 query 层
