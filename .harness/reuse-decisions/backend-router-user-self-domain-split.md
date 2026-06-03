# Reuse Decision

## Change type
backend refactor / router registrar decomposition

## Existing code searched

- `code/backend/internal/app/router_user_self_routes.go`
- `code/backend/internal/app/router_user_teacher_routes.go`
- `code/backend/internal/app/router_authoring_challenge_routes.go`
- `code/backend/internal/app/router_authoring_asset_routes.go`
- `code/backend/internal/app/router_authoring_awd_routes.go`
- `code/backend/internal/app/router_*_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/composition/*.go`
- `code/backend/internal/module/*/runtime/module.go`
- `code/backend/internal/module/*/api/http/*.go`

## Similar implementations found

- 现有 router 治理已经形成稳定模式：总入口保留、局部 deps struct、独立 registrar 文件、结构测试保护迁移结果。
- `registerUserSelfRoutes` 内部天然按 contest participation、practice/runtime、自助画像/报表三段顺序组织，只是还没拆成独立 registrar。
- `router_test.go`、`full_router_integration_test.go`、`full_router_state_matrix_integration_test.go` 已覆盖 user self 路由存在性、runtime handler 接线与访问矩阵，本轮可以直接复用。

## Decision
refactor_existing

## Reason

当前最小正确改动是继续沿用已验证的 registrar 拆分模式，对 `registerUserSelfRoutes` 做整面收口，而不是为了追求更小文件去引入新的 route builder abstraction，或只把 `/users/me/*` 这类局部路径单独挪出去。

按 “contest + practice/runtime + self service” 三组 registrar 切分的原因是：

- 与用户侧行为 owner 一致
- 能保持 contest 内 instance / AWD 操作仍归属于 contest participation surface，而不被底层 handler module 打散
- 比按路径前缀、按 handler module、按是否需要 audit 等技术特征切分更贴近实际维护边界

## Files to modify

- `.harness/reuse-decisions/backend-router-user-self-domain-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-router-user-self-domain-split-plan.md`
- `code/backend/internal/app/router_user_self_routes.go`
- `code/backend/internal/app/router_user_contest_routes.go`
- `code/backend/internal/app/router_user_practice_routes.go`
- `code/backend/internal/app/router_user_self_service_routes.go`
- `code/backend/internal/app/router_user_self_domain_routes_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-user-self-domain-split-review.md`

## After implementation

- `registerUserSelfRoutes` 应退化为 contest / practice / self service 三个 registrar 调用。
- user self surface 的 oversized owner 应从单个文件迁出。
