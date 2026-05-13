# Reuse Decision

## Change type

command / port / infrastructure / runtime / store

## Existing code searched

- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/contest/infrastructure/submission_rate_limit_store.go`
- `docs/plan/impl-plan/2026-05-13-contest-submission-rate-limit-store-phase5-slice7-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/submission_rate_limit_store.go`
- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少 practice 提交限流能力，而是 flag submit 的窗口计数 Redis 细节直接挂在 `practice/application/commands/service.go` 的 `redis` field 上，并被 `submission_service.go` 透传使用。最小正确方案不是再造一个宽缓存服务，也不是把 score state store 扩成通用 Redis bucket，而是沿用 contest slice7 已验证的模式：保留 `Service.SubmitFlag` 的业务 owner，在 `practice/ports` 下新增窄 `PracticeFlagSubmitRateLimitStore`，由 `practice/infrastructure` 统一承接 key、`Incr` 和首次窗口 `Expire` 细节。

## Files to modify

- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/ports/submission_rate_limit_context_contract_test.go`
- `code/backend/internal/module/practice/infrastructure/submission_rate_limit_store.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-submission-rate-limit-store-phase5-slice10-implementation-plan.md`

## After implementation

- 如果后续还要扩展 practice 的其他限流场景，优先在同一个 submission rate-limit store 上按具体 use case 增窄方法，而不是重新把 `RateLimit.RedisKeyPrefix` 和 key 拼接抬回 application
- 如果 assessment 或 ops 还要继续收口 Redis rate-limit / cache state，优先复用这次 contest/practice 的窄 store 模式，而不是创建全局通用 Redis service
