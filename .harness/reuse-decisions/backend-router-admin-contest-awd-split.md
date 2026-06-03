# Reuse Decision

## Change type
backend refactor / router registrar decomposition

## Existing code searched

- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/router_session_routes_test.go`
- `code/backend/internal/app/composition/*.go`
- `code/backend/internal/module/*/runtime/module.go`

## Similar implementations found

- `registerTeacherAuthoringRoutes`、`registerAdminRoutes`、`registerUserRoutes` 已经体现“按访问域分入口”的现有模式，说明本轮应继续沿用 registrar 组织，而不是直接改成新框架或 module 自注册。
- `router_session_routes_test.go` 已经使用“局部路由组 + 最小 stub deps”的测试方式，适合作为本轮 registrar 拆分后的局部测试模式。
- 各 `internal/module/*/runtime/module.go` 已负责 handler 装配，但当前还没有稳定的“模块自带 route registrar”模式，因此本轮不应贸然下沉到 module runtime。

## Decision
refactor_existing

## Reason

当前最小正确改动不是更换 router 模式，而是沿用现有 `internal/app` registrar 结构，把 `registerAdminRoutes` 中最拥挤的 contest / AWD 子域拆到独立文件，并把 deps 收窄。

这样能同时满足三点：

- 保持当前 API、权限、中间件和 handler owner 不变
- 直接收口本轮 touched surface 上的 oversized router debt
- 为后续继续拆 admin identity / ops、user、teacher 路由保留一致模式

如果这一步直接把注册逻辑下沉到 `module runtime`，会同时放大改动面到 composition、module 边界和测试组织，不符合本轮最小切片目标。

## Files to modify

- `.harness/reuse-decisions/backend-router-admin-contest-awd-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-router-admin-contest-awd-split-plan.md`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`（如需要补充断言）
- `code/backend/internal/app/router_admin_contest_routes.go`
- `docs/reviews/backend/2026-06-03-backend-router-admin-contest-awd-split-review.md`

## After implementation

- `internal/app` 内形成可继续复用的 admin contest / AWD registrar 落点。
- 若后续同类拆分会重复查找，可再考虑把该模式补入本地 `.harness/reuse-index/`，本轮先不扩大范围。
