# Reuse Decision

## Change type

command / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/contest/application/commands/submission_*.go`
- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/contest/infrastructure/`
- `docs/plan/impl-plan/2026-05-13-contest-scoreboard-state-store-phase5-slice6-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/application/commands/submission_support.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_score_commands.go`

## Decision

refactor_existing

## Reason

当前 debt 不是 submission 缺少限流能力，而是 `SubmissionService` 直接持有 Redis client，并把错误提交限流的 `Exists/Set` 细节留在 application 层。最小正确方案不是引入通用 cache helper，也不是把限流 key 继续当作 application 的 Redis 实现细节保留在 service 内，而是沿用 phase5 既有做法：保留 submission application 的提交流程 owner，在 `contest/ports` 下新增窄 `ContestSubmissionRateLimitStore`，由 `contest/infrastructure` 提供 Redis adapter，并由 `runtime/module.go` 负责配置 prefix 与注入。

## Files to modify

- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/contest/ports/submission_rate_limit_context_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_rate_limit_store.go`
- `code/backend/internal/module/contest/application/commands/submission_service.go`
- `code/backend/internal/module/contest/application/commands/submission_support.go`
- `code/backend/internal/module/contest/application/commands/submission_submit.go`
- `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- `code/backend/internal/module/contest/application/commands/submission_incorrect_submit.go`
- `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-contest-submission-rate-limit-store-phase5-slice7-implementation-plan.md`

## After implementation

- 如果后续还要收口其他 contest rate-limit 或短期 Redis 状态，优先继续按具体 use case 建窄 store，而不是回退到 application 直接拿 Redis client
- AWD current round / preview token / flag 这些更宽状态面仍需单独切片，不和本次 submission 限流混刀
