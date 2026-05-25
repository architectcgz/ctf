# Reuse Decision

## Change type

api / mapper / command / query / contract

## Existing code searched

- `code/backend/internal/module/auth/**/*.go`
- `code/backend/internal/module/ops/**/*notification*.go`
- `code/backend/internal/module/ops/api/http/request_mapper*.go`
- `code/backend/internal/dto/auth.go`
- `code/backend/internal/dto/notification.go`
- `docs/plan/archive/impl-plan/2026-05-17-identity-admin-user-http-dto-localization-implementation-plan.md`
- `docs/plan/archive/impl-plan/2026-05-17-ops-dashboard-http-dto-localization-implementation-plan.md`

## Similar implementations found

- `identity` 已完成 admin user HTTP DTO 收口：request/response 类型回收 `api/http`，application 不再依赖全局 DTO。
- `ops` dashboard 已完成 query snapshot 与 HTTP DTO 分层：`application/queries` 输出模块内快照，`api/http` 负责映射。
- `practice` / `contest` 已完成 submission output 收口：`application/commands` 使用模块内 output 类型。

## Decision

refactor_existing

## Reason

`auth.go` 和 `notification.go` 中的类型并不属于跨模块共享领域，而是由 `auth` 与 `ops(notification)` 独占消费。继续留在 `internal/dto` 只会扩大可见面并放大 owner 漂移。最小正确改动是沿用现有收口模式，把 request/response 与 application output 收回模块边界，并保持外部 API 契约不变。

## Files to modify

- `.harness/reuse-decisions/auth-notification-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-auth-notification-dto-localization-implementation-plan.md`
- `docs/plan/archive/impl-plan/2026-05-17-challenge-contest-instance-awd-dto-localization-next-batch-plan.md`
- `code/backend/internal/module/auth/**`
- `code/backend/internal/module/ops/**/*notification*.go`
- `code/backend/internal/module/ops/api/http/request_mapper*.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/auth.go`
- `code/backend/internal/dto/notification.go`
