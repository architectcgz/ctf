# Reuse Decision

## Change type
backend refactor / router registrar decomposition

## Existing code searched

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_admin_identity_routes.go`
- `code/backend/internal/app/router_admin_ops_routes.go`
- `code/backend/internal/app/router_user_self_routes.go`
- `code/backend/internal/app/router_user_teacher_routes.go`
- `code/backend/internal/app/router_*_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/composition/*.go`
- `code/backend/internal/module/*/runtime/module.go`
- `code/backend/internal/module/*/api/http/*.go`

## Similar implementations found

- admin 侧 contest / identity / ops 和 user 侧 self / teacher 已经建立稳定模式：总入口保留、局部 deps struct、独立 registrar 文件、结构测试保护迁移结果。
- authoring 路由内部已经天然按普通题、共享资产、AWD 三类 handler owner 分组，只是还没拆成独立 registrar。
- `router_test.go`、`full_router_integration_test.go`、`full_router_state_matrix_integration_test.go` 已覆盖 `/authoring/*` 路径存在性、owner guard 与角色访问矩阵，本轮可以直接复用这批验证面。

## Decision
refactor_existing

## Reason

当前最小正确改动是继续沿用已验证的 registrar 拆分模式，对 `registerTeacherAuthoringRoutes` 做整面收口，而不是引入新的 route builder abstraction，或只把 AWD / image 某一小块随手搬出去。

按 “普通题 authoring + 资源资产 + AWD authoring” 三组 registrar 切分的原因是：

- 与现有 handler owner 一致
- 能把 `/authoring/*` 下真正不同生命周期的子域拆开，而不改变总入口和中间件约束
- 比按 method、是否带 ownerGuard、是否 import 路径等技术特征切分更贴近业务 owner

## Files to modify

- `.harness/reuse-decisions/backend-router-authoring-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-router-authoring-split-plan.md`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_authoring_challenge_routes.go`
- `code/backend/internal/app/router_authoring_asset_routes.go`
- `code/backend/internal/app/router_authoring_awd_routes.go`
- `code/backend/internal/app/router_authoring_routes_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-authoring-split-review.md`

## After implementation

- `registerTeacherAuthoringRoutes` 应退化为 challenge / asset / awd 三个 registrar 调用。
- authoring surface 的 oversized owner 应从总入口函数迁出。
