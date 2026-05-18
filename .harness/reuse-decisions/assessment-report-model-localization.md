# Reuse Decision

## Change type
entity / repository / port / service / domain / mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/report.go`
- `code/backend/internal/module/assessment/...`
- `code/backend/internal/app/full_router*_integration_test.go`
- `.harness/reuse-decisions/*report*`

## Similar implementations found
- `code/backend/internal/module/ops/entity/notification.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`

## Decision
refactor_existing

## Reason
`Report` 及其 report type / format / status 常量只服务于 `assessment` 报告导出链路和对应 app 集成测试，不满足继续留在全局 `internal/model` 共享桶的条件。最小正确方案不是把报告实体塞进 `assessment/infrastructure`，也不是在全局层继续保留一份“通用报告模型”，而是沿用当前模块分层，把 GORM 持久化实体和与之强绑定的常量收回 `internal/module/assessment/entity`，保持表名、字段和行为不变。

非目标：本刀不处理 `skill_profile`，也不扩到 `Dimension*` 常量。

## Files to modify
- `code/backend/internal/model/report.go`
- `code/backend/internal/module/assessment/entity/report.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/assessment/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/assessment/domain/report.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## After implementation
- 删除 `internal/model/report.go`
- 如后续继续把 assessment 其他单一 owner 持久化实体收回模块内，再补充长期 reuse 索引
