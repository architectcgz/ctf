# Reuse Decision

## Change type

practice / event consumer / query cache / runtime wiring

## Existing code searched

- `code/backend/internal/module/practice/application/commands/{submission_service.go,manual_review_service.go}`
- `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/contracts/events.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少事件机制，而是 `practice` progress cache 仍留在写路径里同步删除。最小合理改法不是新建 bridge，也不是把删除逻辑继续塞进 command service，而是复用现有 `practice.flag_accepted` 事件，在 `practice` 查询 owner 上补一个本模块消费者，并通过已有 progress cache adapter 承接 Redis 失效实现。

## Files to modify

- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/ports/progress_timeline_context_contract_test.go`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
- `code/backend/internal/module/practice/application/queries/progress_timeline_context_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-12-practice-progress-cache-eventization-phase3-slice4-implementation-plan.md`

## After implementation

- 如果 practice 未来还有其他 query-side cache，需要继续沿“query owner + 事件消费者”的方式收口
- 如果后续要把 Redis 细节继续下沉到更窄 adapter，本轮事件 owner 划分仍可保持不变
