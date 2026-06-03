# Reuse Decision

## Change type
backend test refactor / oversized router test split

## Existing code searched
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/router_session_routes_test.go`
- `code/backend/internal/app/router_admin_contest_routes_test.go`
- `code/backend/internal/app/router_admin_identity_ops_routes_test.go`
- `code/backend/internal/app/router_authoring_routes_test.go`
- `code/backend/internal/app/router_user_self_domain_routes_test.go`
- `code/backend/internal/app/router_user_teacher_routes_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`

## Similar implementations found
- `internal/app` 已经有一批按访问域拆开的 router 测试文件，说明这个目录接受“shared helper 留在原文件，新文件承载更窄测试 owner”的模式。
- `practice_flow` 和 `full_router` 的最近拆分也已经验证了同样的做法：不改测试基建，只把 oversized 文件里的顶层测试按职责拆开。

## Decision
refactor_existing

## Reason
本轮不引入新的测试 helper 包，也不改 `newAppTestDependencies`、TLS helper 或反射断言 helper，而是在 `router_test.go` 上继续做最小结构收敛：

- 原文件保留 shared helper owner
- 路由注册、guard、query param、dial failure 这些行为测试移到独立文件
- composition struct / builder / typed deps / source marker 守卫移到按主题归类的新文件

相比继续把新增 router/composition 守卫堆回 `router_test.go`，这条路径能直接降低 1000+ 行文件的 review 成本，同时保持断言语义不变。

## Files to modify
- `.harness/reuse-decisions/backend-internal-app-router-test-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-router-test-split-plan.md`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/router_route_wiring_test.go`
- `code/backend/internal/app/router_composition_structure_test.go`
- `code/backend/internal/app/router_composition_typed_deps_test.go`

## After implementation
- `router_test.go` 只保留 shared helper、fixture 和 TLS helper。
- 路由行为测试单独放在 `router_route_wiring_test.go`。
- composition 结构与 builder/source marker 守卫放在 `router_composition_structure_test.go`。
- module typed deps 与 cross-module guard 放在 `router_composition_typed_deps_test.go`。
