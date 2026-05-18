# Reuse Decision

## Change type
entity / repository / port / service / mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/notification.go`
- `code/backend/internal/model/notification_batch.go`
- `code/backend/internal/module/ops/...`
- `code/backend/internal/app/full_router*_integration_test.go`
- `.harness/reuse-decisions/*notification*`

## Similar implementations found
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`
- `.harness/reuse-decisions/ops-notification-not-found-contract-phase5-slice13.md`

## Decision
refactor_existing

## Reason
`Notification` / `NotificationBatch` 只被 `ops` 模块和 app 集成测试消费，不满足继续留在全局 `internal/model` 共享桶的条件。最小正确方案不是新建全局复用层，也不是把实体拆成抽象 domain model，而是沿用现有 `ops` command/query/repository 结构，把这两个 GORM 持久化实体和通知类型常量直接收回 `internal/module/ops/entity`，同时保持表名、字段和仓储行为不变。

非目标：本刀不处理 `audit_log`，也不扩到 `user`、`contest`、`submission` 等多 owner 核心实体。

## Files to modify
- `code/backend/internal/model/notification.go`
- `code/backend/internal/model/notification_batch.go`
- `code/backend/internal/module/ops/entity/notification.go`
- `code/backend/internal/module/ops/entity/notification_batch.go`
- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/response_mapper.go`
- `code/backend/internal/module/ops/application/commands/response_mapper_gen.go`
- `code/backend/internal/module/ops/application/queries/notification_service.go`
- `code/backend/internal/module/ops/application/queries/response_mapper.go`
- `code/backend/internal/module/ops/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/ops/application/queries/notification_service_test.go`
- `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## After implementation
- 删除 `internal/model/notification*.go`
- 如该模式后续会复用到其他单一 owner 持久化实体，再补充 `harness/reuse/history.md` / `harness/reuse/index.yaml`
