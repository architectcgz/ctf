# Reuse Decision

## Change type

service / handler / api

## Existing code searched

- `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- `code/backend/internal/module/ops/api/http/dashboard_handler.go`
- `code/backend/internal/module/ops/api/http/response_mapper_assign.go`
- `code/backend/internal/module/identity/api/http/response_mapper.go`
- `code/backend/internal/module/ops/ports/dashboard.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/dto/dashboard.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Similar implementations found

- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/ops/ports/dashboard.go`

## Decision

refactor_existing

## Reason

这次不是新增一套 dashboard 读模型，也不是把查询结果继续留在全局 `internal/dto`。现有 `ops/ports.DashboardStatsSnapshot` 已经覆盖 dashboard query 的内部数据形状，并且同时承担缓存读写的 owner；最小正确方案是复用这份模块内 snapshot 作为 query 返回值，再在 `ops/api/http` 定义本地 response DTO 与 mapper。

这样可以同时满足三件事：

- `ops` dashboard 的 query/HTTP 边界不再依赖全局 `internal/dto/dashboard.go`
- application 层不需要反向依赖 `ops/api/http`，避免为了 DTO 本地化把层次关系改坏
- `/api/v1/admin/dashboard` 外部 JSON 契约保持不变，只调整 owner

另外，admin dashboard 的全路由回归会经过 `/api/v1/admin/users`；现有 `identity/api/http` 已经依赖 generated response mapper，但缺少非 `goverter` 构建下的 assign 文件，因此这次顺手补齐初始化属于同一条 admin HTTP 回归链上的必要修补，而不是扩大 DTO 重构范围。

## Files to modify

- `.harness/reuse-decisions/ops-dashboard-http-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-ops-dashboard-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/ops/api/http/dashboard_handler.go`
- `code/backend/internal/module/ops/api/http/response_types.go`
- `code/backend/internal/module/ops/api/http/response_mapper.go`
- `code/backend/internal/module/ops/api/http/response_mapper_assign.go`
- `code/backend/internal/module/ops/api/http/response_mapper_gen.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/dashboard.go`
- `code/backend/internal/module/identity/api/http/response_mapper_assign.go`

## After implementation

- 如果 `internal/dto/dashboard.go` 在仓库内无剩余引用，直接删除
- 不新增新的全局 shared DTO 或跨模块 contracts
- 这次决定仅覆盖 ops dashboard 这条链，不扩到 notification / audit / risk
