# Reuse Decision

## Change type
entity / repository / port / service / app / middleware / test / model localization

## Existing code searched
- `code/backend/internal/model/audit_log.go`
- `code/backend/internal/module/ops/...`
- `code/backend/internal/app/...`
- `code/backend/internal/module/auth/...`
- `.harness/reuse-decisions/*audit*log*`

## Similar implementations found
- `code/backend/internal/module/ops/entity/notification.go`
- `code/backend/internal/module/challenge/entity/image_build_job.go`
- `code/backend/internal/module/contest/entity/announcement.go`

## Decision
refactor_existing

## Reason
`AuditLog` 持久化实体的 owner 在 `ops` 模块，但审计动作常量属于跨模块共享语义，不能通过 `ops` 私有层向外暴露。最小正确方案是把 GORM 实体收回 `internal/module/ops/entity`，同时把动作常量放到共享的 `internal/auditlog` 包，保持表名、字段、动作值和外部行为不变。

非目标：本刀不处理 `User`、`Contest`、`Submission`、`AWD*` 等仍被多个模块共同拥有的核心实体。

## Files to modify
- `code/backend/internal/model/audit_log.go`
- `code/backend/internal/module/ops/entity/audit_log.go`
- `code/backend/internal/module/ops/ports/audit.go`
- `code/backend/internal/module/ops/infrastructure/audit_repository.go`
- `code/backend/internal/module/ops/infrastructure/risk_repository.go`
- `code/backend/internal/module/ops/application/commands/audit_service.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/middleware/audit_test.go`
- `code/backend/internal/middleware/awd_readiness_audit.go`
- `code/backend/internal/middleware/awd_readiness_audit_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/timeline_query_repository.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/ops/application/commands/audit_service_test.go`
- `code/backend/internal/module/ops/application/queries/audit_service_test.go`

## After implementation
- 删除 `internal/model/audit_log.go`
- 后续再单独处理其他跨模块共享但 owner 不清晰的模型
