# Reuse Decision

## Change type

readmodel / handler / service / repository / port / composition / api

## Existing code searched

- `code/backend/internal/module/practice_readmodel/**`
- `code/backend/internal/module/practice/**`
- `code/backend/internal/app/composition/**`
- `code/backend/internal/app/{router.go,router_routes.go,router_test.go,full_router_integration_test.go,practice_flow_integration_test.go}`
- `code/backend/internal/module/challenge/{ports,infrastructure}/**`
- `code/backend/internal/module/contest/{ports,infrastructure}/**`

## Similar implementations found

- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`

## Decision

refactor_existing

## Reason

`practice_readmodel` 的 progress / timeline 只读取 `practice` 自有事实，没有跨 owner contract 依赖，不满足独立 readmodel 的保留条件。最小合理改法不是新建第三套 query 壳，而是复用现有 `practice` 模块的 handler / application / infrastructure / runtime 结构，把查询并回 `practice`，同时沿用 challenge / contest 最近的模块内 port + infrastructure adapter 模式收口 progress cache。

## Files to modify

- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/api/http/handler_progress_test.go`
- `code/backend/internal/module/practice/application/queries/*`
- `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
- `code/backend/internal/module/practice/infrastructure/*`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/composition/practice_module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/architecture/backend/{01-system-architecture.md,07-modular-monolith-refactor.md}`
- `docs/design/backend-module-boundary-target.md`

## After implementation

- 如这次并回 `practice` 的模式后续仍会反复出现，再把结论补进 `harness/reuse/history.md` 或 `harness/reuse/index.yaml`
