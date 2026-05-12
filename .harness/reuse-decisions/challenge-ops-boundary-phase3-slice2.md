# Reuse Decision

## Change type
service / composition / event / notification

## Existing code searched
- `code/backend/internal/module/challenge/application/commands/`
- `code/backend/internal/module/challenge/runtime/`
- `code/backend/internal/module/challenge/contracts/`
- `code/backend/internal/module/ops/application/commands/`
- `code/backend/internal/module/ops/runtime/`
- `code/backend/internal/app/composition/`

## Similar implementations found
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`

## Decision
reuse_existing

## Reason
这次不新增新的通知基础设施，也不再造 challenge 专用 bridge。仓库里已经有三块可复用模式：`platformevents.Bus` 作为进程内事件总线、`ops.NotificationService` 作为统一通知 owner、`practice` / `contest` 在 application 层通过弱发布事件表达异步副作用。最小正确方案是让 `challenge` 复用同样的事件发布模式，并让 `ops` 侧扩展一个 challenge 事件消费者，而不是继续在 composition 里 new 一份临时通知 sender。

## Files to modify
- `code/backend/internal/app/composition/challenge_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/challenge/contracts/events.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/ops/runtime/module.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`
- `docs/plan/impl-plan/2026-05-12-challenge-ops-boundary-phase3-slice2-implementation-plan.md`

## After implementation
- 如果这次模式稳定，可以把“发布检查结果、通知、缓存失效统一走 owner 事件”的复用结论补进长期 reuse 历史或更上层 skill。
