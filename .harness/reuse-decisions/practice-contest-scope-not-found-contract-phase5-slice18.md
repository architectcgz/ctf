# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/challenge/contracts/contracts.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 `practice` 缺少 contest / runtime subject 查询能力，而是 `contest_instance_scope.go` 和 `contest_awd_operations.go` 直接知道 `gorm.ErrRecordNotFound`。最小正确方案不是全局改写 `practice` 原始仓储或 `challenge` 原始仓储的错误语义，那会波及还没进入本 slice 的 `manual_review`、`submission` 等 surface。这里沿用前面几个 not-found contract slice 的模式，但把 owner 收得更窄：

- `practice` 自己新增局部 sentinel 和两个窄 adapter port
- `practice/infrastructure` 用局部 adapter 把 raw repository / challenge contract 的 GORM not-found 映射成模块内 sentinel
- application 只看模块内 sentinel，不再直接 import `gorm`

## Files to modify

- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/composition/practice_module.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-contest-scope-not-found-contract-phase5-slice18-implementation-plan.md`

## After implementation

- 后续如果继续收口 `practice` 里其他 challenge / user / review 查询的 GORM concrete，优先复用这次“局部 adapter + 模块内 sentinel”的模式，而不是直接改 raw repository 的全局错误语义
- 如果未来 `practice` 对 `challenge` 的 cross-module contract 继续收窄，可把这次 runtime subject adapter 继续沉到更明确的 practice-owned capability，而不是重新回到宽 `challengeRepo`
