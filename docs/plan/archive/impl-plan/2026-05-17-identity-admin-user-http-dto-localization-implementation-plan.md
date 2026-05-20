# identity admin user HTTP DTO 模块内化实现方案

## Objective

把 `internal/dto/admin_user.go` 里的 admin user HTTP request/response DTO 收回 `identity/api/http`，保持 `/api/v1/admin/users` 相关外部契约不变。

## Non-goals

- 不改 `identity/contracts` 与 `identity` application 的现有 contract 边界
- 不改 admin user 路由、分页、状态码、JSON 字段
- 不处理 `auth`、`profile`、`report` 等其他全局 DTO

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/plan/impl-plan/2026-05-17-identity-contract-dto-convergence-implementation-plan.md`
- `code/backend/internal/module/identity/api/http/*.go`
- `code/backend/internal/dto/admin_user.go`

## Task Slices

1. 在 `identity/api/http` 定义本地 admin user request/response DTO
   - 新增本地 query / create / update / import response 类型
   - 保持 form / json tag 与原外部契约一致

2. 收口 request / response mapper 与 handler
   - `identity/api/http` 改为只依赖本地 HTTP DTO 与 `identity/contracts`
   - 重新生成 mapper

3. 调整 app 级测试解码类型
   - `full_router_state_matrix_integration_test.go` 改为使用 `identity/api/http` 本地类型解码

4. 删除废弃全局 DTO
   - 若 `internal/dto/admin_user.go` 无剩余引用则删除

## Expected Changes

- `code/backend/internal/module/identity/api/http/*.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/admin_user.go`

## Validation

- `go generate ./internal/module/identity/api/http`
- `go test ./internal/module/identity/... -count=1`
- `go test ./internal/app -run TestFullRouter_AdminOpsAndNotificationStateMatrix -count=1`

## Review Focus

- admin user HTTP DTO owner 是否已经回到 `identity/api/http`
- 是否只改了 owner，没有改外部 API 契约
- 是否没有把 HTTP DTO 重新漏回 `identity/contracts` 或 application

