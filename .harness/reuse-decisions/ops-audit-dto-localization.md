# Reuse Decision

## Change type

contract / query / ops / audit

## Existing code searched

- `code/backend/internal/dto/audit.go`
- `code/backend/internal/module/ops/api/http/*.go`
- `code/backend/internal/module/ops/application/queries/*.go`

## Similar implementations found

- `ops/application/queries/notification_output.go` 已承接 ops query request / response 类型
- `ops/api/http/notification_handler.go` 已直接依赖 `ops/application/queries` 本地 query 类型

## Decision

refactor_existing

## Reason

`AuditLogQuery` 和 `AuditLogItem` 当前只被 `ops` 的 audit handler、query service 和对应测试使用，owner 已经很单一。最小正确方案是把这组类型收回 `ops/application/queries`，沿用 notification 的现有模式，让 handler 和 service 共享同一组模块内类型，不再依赖全局 `dto`。

## Files to modify

- `.harness/reuse-decisions/ops-audit-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-18-ops-audit-dto-localization-implementation-plan.md`
- `code/backend/internal/dto/audit.go`
- `code/backend/internal/module/ops/api/http/audit_handler.go`
- `code/backend/internal/module/ops/application/queries/audit_output.go`
- `code/backend/internal/module/ops/application/queries/audit_service.go`
- `code/backend/internal/module/ops/application/queries/audit_service_test.go`

## After implementation

- `ops` audit handler / query service / tests 不再引用 `dto.AuditLogQuery` 或 `dto.AuditLogItem`
- `internal/dto/audit.go` 删除
