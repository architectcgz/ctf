# Reuse Decision

## Change type

service / handler / mapper / api

## Existing code searched

- `code/backend/internal/module/identity/contracts/*.go`
- `code/backend/internal/module/identity/application/commands/*.go`
- `code/backend/internal/module/identity/application/queries/*.go`
- `code/backend/internal/module/identity/api/http/*.go`
- `code/backend/internal/module/auth/api/http/*.go`
- `code/backend/internal/dto/admin_user.go`
- `code/backend/internal/dto/auth.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`

## Similar implementations found

- `identity/contracts` 已经承担输入 contract 和仓储 contract，不需要再新增并行 contract 目录
- `identity/api/http/request_mapper.go` 已经是请求 DTO 到 contract 的既有映射落点，适合继续承接 response 映射
- `auth/api/http/handler.go` 已经消费 `identity` profile contract，这次只需要把 profile response 从 contract 映射回 HTTP DTO

## Decision

refactor_existing

## Reason

这次不是新增新的用户模块、并行 DTO 包或第二套 admin/profile handler，而是在既有 `identity` 模块里把跨模块 contract 与 HTTP DTO 的职责重新收口：

- `identity/contracts` 继续作为模块对外稳定 contract owner
- `identity/application/*` 返回 `identity/contracts` 类型，不再直接暴露 `internal/dto`
- `identity/api/http` 成为 `dto <-> contracts` 的唯一映射边界
- `auth/api/http` 只补一层薄映射，把 `identitycontracts.ProfileUser` 转回 `dto.AuthUser`

这样复用了现有模块、现有 handler 和现有 mapper 结构，没有引入新的并行抽象，同时把 `internal/dto` 从 contract/application 层退出。

## Files to modify

- `.harness/reuse-decisions/identity-contract-dto-convergence.md`
- `docs/plan/impl-plan/2026-05-17-identity-contract-dto-convergence-implementation-plan.md`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/response_mapper.go`
- `code/backend/internal/module/identity/api/http/handler.go`
- `code/backend/internal/module/identity/api/http/request_mapper.go`
- `code/backend/internal/module/identity/api/http/request_mapper_gen.go`
- `code/backend/internal/module/identity/api/http/response_mapper.go`
- `code/backend/internal/module/identity/api/http/response_mapper_gen.go`
- `code/backend/internal/module/identity/application/commands/admin_service.go`
- `code/backend/internal/module/identity/application/commands/response_mapper.go`
- `code/backend/internal/module/identity/application/commands/response_mapper_gen.go`
- `code/backend/internal/module/identity/application/commands/support.go`
- `code/backend/internal/module/identity/application/queries/admin_service.go`
- `code/backend/internal/module/identity/application/queries/admin_service_test.go`
- `code/backend/internal/module/identity/application/queries/profile_service.go`
- `code/backend/internal/module/identity/application/queries/response_mapper.go`
- `code/backend/internal/module/identity/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/identity/application/queries/support.go`
- `code/backend/internal/module/identity/architecture_test.go`
- `code/backend/internal/module/identity/contracts/admin.go`
- `code/backend/internal/module/identity/contracts/profile.go`

## After implementation

- `identity/contracts` 不再直接 import `internal/dto`
- `identity` 的 admin / profile 对外服务面改为先出 contract，再由 API 层映射到 HTTP DTO
- `dto.AdminUser*`、`dto.ImportUsersResp`、`dto.AuthUser` 仍保留在 HTTP 边界和集成测试使用，后续如果继续清理，需要下一刀把 HTTP DTO 落点也模块内化
