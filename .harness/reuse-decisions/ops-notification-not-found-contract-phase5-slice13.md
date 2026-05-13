# Reuse Decision

## Change type

command / port / infrastructure / runtime / contract

## Existing code searched

- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/ops/runtime/module.go`
- `docs/plan/impl-plan/2026-05-13-ops-dashboard-state-store-phase5-slice11-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/assessment/infrastructure/state_store.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 notification command 能力，而是 `ops/application/commands/notification_service.go` 为了识别“通知不存在”直接 import `gorm.io/gorm` 并判断 `gorm.ErrRecordNotFound`。最小正确方案不是在 application 层继续泄漏 ORM sentinel，也不是把 `FindByID` 改成宽泛 nullable helper，而是把 not-found 语义收口成 `ops/ports` 自己的错误契约，由 `ops/infrastructure/notification_repository.go` 把 `gorm.ErrRecordNotFound` 映射成模块内 sentinel，application 只处理 `ops` 自己的 repo 语义。

## Files to modify

- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-ops-notification-not-found-contract-phase5-slice13-implementation-plan.md`

## After implementation

- 如果 `ops` 其他 command / query 以后还要区分 repository not-found，优先继续在 `ops/ports` 暴露模块内错误契约，而不是把 GORM sentinel 再抬回 application
- 如果后续处理其他模块的 GORM concrete allowlist，优先先看是不是 “只为了识别 ErrRecordNotFound” 这类 contract 泄漏，再决定是否需要更大的 repository surface 重构
