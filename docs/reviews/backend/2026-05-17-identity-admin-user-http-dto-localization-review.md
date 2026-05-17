# Identity Admin User HTTP DTO Localization Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 identity admin user HTTP DTO 模块内化相关 diff，重点覆盖 `code/backend/internal/module/identity/api/http/*`、`code/backend/internal/app/full_router_state_matrix_integration_test.go`、`code/backend/internal/dto/admin_user.go` 以及对应文档
- Files reviewed:
  - `code/backend/internal/module/identity/api/http/admin_user_types.go`
  - `code/backend/internal/module/identity/api/http/request_mapper.go`
  - `code/backend/internal/module/identity/api/http/request_mapper_gen.go`
  - `code/backend/internal/module/identity/api/http/response_mapper.go`
  - `code/backend/internal/module/identity/api/http/response_mapper_gen.go`
  - `code/backend/internal/module/identity/api/http/handler.go`
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/internal/dto/admin_user.go`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/plan/impl-plan/2026-05-17-identity-admin-user-http-dto-localization-implementation-plan.md`
  - `.harness/reuse-decisions/identity-admin-user-http-dto-localization.md`
- Classification check: agree with pipeline，属于 non-trivial backend refactor + review gate
- Initial gate verdict: blocked

## Current Status

- 2026-05-17 补充修复状态：已完成。独立 review 只发现 1 个 blocker，为 `docs/architecture/backend/04-api-design.md` 把 admin user 更新接口误写成 `PATCH /api/v1/admin/users/:id`；当前已修正为与真实路由一致的 `PUT /api/v1/admin/users/:id`。

## Findings

1. `docs/architecture/backend/04-api-design.md:53`
   - material / blocker
   - 新增的 API 事实源条目把 admin user 更新接口写成了 `PATCH /api/v1/admin/users/:id`，但真实路由仍是 `PUT`。这会误导后续契约维护、测试和客户端接入。
   - 处理结果：已修正为 `PUT /api/v1/admin/users/:id`，与 `router_routes.go` 和全链路测试保持一致。

## Validation Evidence

- `cd code/backend && go generate ./internal/module/identity/api/http`
- `cd code/backend && go test ./internal/module/identity/... -count=1`
- `cd code/backend && go test ./internal/app -run TestFullRouter_AdminOpsAndNotificationStateMatrix -count=1`

## Final Review Verdict

- Gate verdict after fix: pass
- 结论：除文档里的 HTTP method 误写外，未发现新的 correctness、回归、DTO owner 漂移或外部契约变化问题。admin user HTTP request/response DTO 已收口到 `identity/api/http`，外部路径、字段和 envelope 保持不变。
