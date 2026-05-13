# Reuse Decision

## Change type

command / port / infrastructure / contract

## Existing code searched

- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `docs/plan/impl-plan/2026-05-13-assessment-profile-lock-and-recommendation-cache-phase5-slice12-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 report service 能力，而是 `assessment/application/commands/report_service.go` 为了识别“报告不存在 / 竞赛不存在”直接 import `gorm.io/gorm` 并判断 `gorm.ErrRecordNotFound`。最小正确方案不是在 application 层继续泄漏 ORM sentinel，也不是把多个 lookup contract 改成 nullable 宽 helper，而是把 not-found 语义收口成 `assessment/ports` 自己的错误契约，由 `assessment/infrastructure/report_repository.go` 负责把 GORM not-found 映射成模块内 sentinel，report service 只处理 assessment 自己的 repo 语义。

## Files to modify

- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-assessment-report-not-found-contract-phase5-slice14-implementation-plan.md`

## After implementation

- 如果 assessment 其他 application surface 以后还要区分 persistence not-found，优先继续在 `assessment/ports` 暴露模块内错误契约，而不是把 GORM sentinel 再抬回 application
- 如果后续处理 challenge / contest / practice 的 GORM concrete allowlist，优先先看是不是类似 not-found contract 泄漏，再决定是否要做更大的 repository surface 重排
