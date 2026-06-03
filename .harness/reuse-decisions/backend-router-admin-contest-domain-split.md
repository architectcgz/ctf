# Reuse Decision

## Change type
backend refactor / router registrar decomposition

## Existing code searched

- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_admin_identity_routes.go`
- `code/backend/internal/app/router_admin_ops_routes.go`
- `code/backend/internal/app/router_authoring_*_routes.go`
- `code/backend/internal/app/router_user_*_routes.go`
- `code/backend/internal/app/router_admin_contest_routes_test.go`
- `code/backend/internal/app/router_*_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/module/*/api/http/*.go`
- `code/backend/internal/module/*/runtime/module.go`
- `code/backend/internal/app/composition/*.go`

## Similar implementations found

- 当前 router 治理已经形成稳定模式：总入口保留、局部 deps struct、独立 registrar 文件、结构测试保护迁移结果。
- `router_admin_contest_routes.go` 自身已经有 `contestByID` 和 `registerAdminContestAWDRoutes` 的层级轮廓，说明继续按 owner 拆成子 registrar 比改路由框架更符合现有结构。
- `router_test.go`、`full_router_integration_test.go`、`full_router_state_matrix_integration_test.go` 已覆盖 admin contest / AWD 路径存在性、practice wiring 与访问矩阵，本轮可以直接复用。

## Decision
refactor_existing

## Reason

当前最小正确改动是继续沿用已验证的 registrar 拆分模式，对 `router_admin_contest_routes.go` 做整面收口，而不是继续把几类 owner 留在一个文件里，或引入新的 route DSL / module 自注册方案。

按 “contest core + challenge roster + participation + AWD” 四组 registrar 切分的原因是：

- 与 admin contest 维护时的实际行为 owner 一致
- 能保留 `contestByID` 组装和局部 middleware 复用，不扩大改动面到 composition root
- 比按 HTTP method、是否调用 practice handler、是否带 awdReadinessAudit 这类技术特征切分更贴近真实维护边界

## Files to modify

- `.harness/reuse-decisions/backend-router-admin-contest-domain-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-router-admin-contest-domain-split-plan.md`
- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_admin_contest_core_routes.go`
- `code/backend/internal/app/router_admin_contest_challenge_routes.go`
- `code/backend/internal/app/router_admin_contest_participation_routes.go`
- `code/backend/internal/app/router_admin_contest_awd_routes.go`
- `code/backend/internal/app/router_admin_contest_routes_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-admin-contest-domain-split-review.md`

## After implementation

- `router_admin_contest_routes.go` 应退化为 core / challenge / participation / awd 的 delegator。
- admin contest surface 的 oversized owner 应从单个文件迁出。
