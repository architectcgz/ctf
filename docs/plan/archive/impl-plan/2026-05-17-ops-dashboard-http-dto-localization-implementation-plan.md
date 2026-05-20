# ops dashboard HTTP DTO 模块内化实现方案

## Objective

把 ops dashboard 的 HTTP response DTO 从 `internal/dto/dashboard.go` 收回 `internal/module/ops/api/http`，让 `/api/v1/admin/dashboard` 的查询/HTTP 边界只依赖 ops 模块自身类型。

## Non-goals

- 不调整 `notification`、`audit`、`risk` 的任何 DTO owner
- 不修改 `/api/v1/admin/dashboard` 的外部 JSON 字段、含义和状态码
- 不改 README 或共享架构文档
- 不重命名 `ops/ports.DashboardStatsSnapshot`

## Inputs

- `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- `code/backend/internal/module/ops/api/http/dashboard_handler.go`
- `code/backend/internal/module/ops/ports/dashboard.go`
- `code/backend/internal/dto/dashboard.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/ops-dashboard-http-dto-localization.md`

## Ownership And Boundary Decision

- query owner：`ops/application/queries`，返回 `ops/ports.DashboardStatsSnapshot`
- HTTP DTO owner：`ops/api/http`
- cache owner：继续由 `ops/ports.DashboardStatsSnapshot` + `ops/infrastructure/DashboardStateStore` 持有
- 不再让 `ops` dashboard query/handler 依赖全局 `internal/dto/dashboard.go`

## Task Slices

1. 在 `ops/api/http` 定义 dashboard response DTO 与 response mapper
   - 新增模块内 `DashboardStats`、`ContainerStat`、`ResourceAlert`
   - 新增从 `ops/ports.DashboardStatsSnapshot` 到 HTTP DTO 的 mapper
   - 验证：JSON tag 与现有字段完全一致

2. 收口 query 与 handler
   - `DashboardService.GetDashboardStats` 改为返回 `*opsports.DashboardStatsSnapshot`
   - handler 接口改为消费 snapshot，并在 HTTP 边界完成映射
   - 验证：`go test ./internal/module/ops/...`

3. 更新测试并判断删除条件
   - `dashboard_service_test.go` 改用 snapshot 断言
   - `full_router_state_matrix_integration_test.go` 改用 `ops/api/http.DashboardStats` 解码
   - 全仓搜索 `internal/dto/dashboard.go` 的剩余引用；若无引用则删除文件
   - 验证：`go test ./internal/module/ops/...`，必要时补 `go test ./internal/app/... -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix`

## Expected Changes

- `.harness/reuse-decisions/ops-dashboard-http-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-ops-dashboard-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/ops/api/http/`
- `code/backend/internal/module/ops/application/queries/`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/dashboard.go`

## Compatibility

- `/api/v1/admin/dashboard` 对外 JSON 契约保持不变
- 现有缓存内容的 JSON 结构不变，因为 cache 继续存 `DashboardStatsSnapshot`

## Validation

- `go test ./internal/module/ops/...`
- `go test ./internal/app/... -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix`
- `rg -n "dto\\.DashboardStats|dto\\.ContainerStat|dto\\.ResourceAlert|internal/dto/dashboard.go" code/backend`

## Review Focus

- HTTP DTO 是否已经收回 `ops/api/http`
- application 层是否不再依赖全局 dashboard DTO
- handler 是否只在 HTTP 边界做映射，且 JSON 契约未变
- `internal/dto/dashboard.go` 是否已经无引用，可安全删除

## Rollback

- 如果本地 response mapper 或 handler 映射导致集成测试异常，可先保留 `internal/dto/dashboard.go`，回退到 query 返回 snapshot 但 handler 临时用手写映射，再重新收紧到模块内 DTO
