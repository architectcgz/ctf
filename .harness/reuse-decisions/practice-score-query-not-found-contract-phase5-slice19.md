# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/application/queries/score_service_test.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 `practice` 缺少排行榜或用户目录查询能力，而是 `practice/application/queries/score_service.go` 直接知道 `gorm.ErrRecordNotFound`。最小正确方案仍然沿用 phase5 已经验证过的 not-found contract 模式：

- `practice/ports` 新增局部 sentinel：`ErrPracticeUserScoreNotFound`
- `practice/infrastructure` 新增一个很窄的 score query adapter，只负责把 raw score repository 的 `gorm.ErrRecordNotFound` 映射成模块内 sentinel
- `practice` query application 只看模块内 sentinel，不改 raw repository 的全局错误语义

这样可以把 slice 收在 score query surface，不把范围扩到 `manual_review_service.go`、`submission_service.go` 或 `score_repository.go` 的其他调用点。

## Files to modify

- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/application/queries/score_service_test.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-score-query-not-found-contract-phase5-slice19-implementation-plan.md`

## After implementation

- 后续继续收 `practice/application/commands/manual_review_service.go` 和 `submission_service.go` 时，优先复用这次“局部 sentinel + 窄 adapter”的模式，而不是直接改 `practice` 或 `challenge` raw repository 的全局 not-found 语义
- 如果后续 `practice` score query 还需要更多 read-model contract，可继续在这个 adapter 上收口，而不是把 GORM concrete 带回 query application
