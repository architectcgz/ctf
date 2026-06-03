# Reuse Decision

## Change type
backend refactor / router registrar decomposition

## Existing code searched

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_admin_identity_routes.go`
- `code/backend/internal/app/router_admin_ops_routes.go`
- `code/backend/internal/app/router_*_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/composition/*.go`
- `code/backend/internal/module/*/runtime/module.go`
- `code/backend/internal/module/*/api/http/*.go`

## Similar implementations found

- admin 侧三刀已经建立了稳定模式：总入口保留、局部 deps struct、独立 registrar 文件、结构测试保护迁移结果。
- `router_test.go`、`full_router_integration_test.go`、`full_router_state_matrix_integration_test.go` 已覆盖 user / teacher 路由存在性与访问矩阵，说明本轮可直接复用现有验证面，不需要新造行为测试框架。

## Decision
refactor_existing

## Reason

当前最小正确改动是继续沿用已验证的 registrar 拆分模式，对 `registerUserRoutes` 做整面收口，而不是引入新的 routing abstraction 或只拆一半。

按 “user self + teacher” 两组 registrar 切分的原因是：

- 与现有权限边界一致
- 能自然容纳 `users/:id/skill-profile` 这类 teacher-protected 但不在 `/teacher/*` 命名空间下的路由
- 比单纯按路径前缀切分更贴近行为 owner

## Files to modify

- `.harness/reuse-decisions/backend-router-user-teacher-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-router-user-teacher-split-plan.md`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_user_self_routes.go`
- `code/backend/internal/app/router_user_teacher_routes.go`
- `code/backend/internal/app/router_user_teacher_routes_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-user-teacher-split-review.md`

## After implementation

- `registerUserRoutes` 应退化为两个 registrar 调用。
- user / teacher surface 的 oversized owner 应从总入口函数迁出。
