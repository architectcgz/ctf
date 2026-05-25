# Reuse Decision

## Change type

handler / api / mapper

## Existing code searched

- `code/backend/internal/module/identity/api/http/*.go`
- `code/backend/internal/dto/admin_user.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/ops/api/http/response_types.go`

## Similar implementations found

- `assessment/api/http` 与 `ops/api/http` 已经把模块自有 HTTP DTO 收回 `api/http`
- `identity/api/http` 已经有 request / response mapper，可直接沿用

## Decision

refactor_existing

## Reason

admin user 这批类型只被 `identity/api/http` 和 app 测试消费，不满足继续留在全局 `internal/dto` 的共享条件。最小合理方案是复用现有 `identity/api/http` mapper / handler 结构，把 request/response DTO 直接收回模块边界。

## Files to modify

- `.harness/reuse-decisions/identity-admin-user-http-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-identity-admin-user-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/identity/api/http/admin_user_types.go`
- `code/backend/internal/module/identity/api/http/handler.go`
- `code/backend/internal/module/identity/api/http/request_mapper.go`
- `code/backend/internal/module/identity/api/http/request_mapper_gen.go`
- `code/backend/internal/module/identity/api/http/response_mapper.go`
- `code/backend/internal/module/identity/api/http/response_mapper_gen.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/admin_user.go`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-identity-admin-user-http-dto-localization-review.md`
