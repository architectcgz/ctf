# Reuse Decision

## Change type
backend refactor / router registrar decomposition

## Existing code searched

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_admin_contest_routes_test.go`
- `code/backend/internal/app/router_session_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/composition/*.go`
- `code/backend/internal/module/*/runtime/module.go`
- `code/backend/internal/module/*/api/http/*.go`

## Similar implementations found

- 上一刀已经建立了 `router_admin_contest_routes.go` + 局部 deps struct + registrar 结构测试的模式，适合作为当前 admin ops / identity/session 拆分的直接复用模板。
- `router_session_routes_test.go` 已覆盖 session list / revoke 行为，说明这次不需要新造复杂 session fixture，只需保证 registrar 迁移后继续复用现有测试。

## Decision
refactor_existing

## Reason

当前最小正确改动是继续沿用刚落地的 registrar 拆分模式，把 admin ops 和 admin identity/session 从 `registerAdminRoutes` 中迁出。

这比直接改 module runtime 或一次性重写整个 router 更稳，因为：

- 当前模式已经在上一刀验证过
- 本轮 write surface 仍然局限在 `internal/app`
- 现有 session 行为测试可直接复用，不需要扩大行为改动面

## Files to modify

- `.harness/reuse-decisions/backend-router-admin-identity-ops-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-router-admin-identity-ops-split-plan.md`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_admin_ops_routes.go`
- `code/backend/internal/app/router_admin_identity_routes.go`
- `code/backend/internal/app/router_session_routes_test.go`（如需要）
- `code/backend/internal/app/router_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-admin-identity-ops-split-review.md`

## After implementation

- `registerAdminRoutes` 应只保留 admin 入口级分发。
- admin ops 与 admin identity/session 将形成可继续复用的独立 registrar 落点。
