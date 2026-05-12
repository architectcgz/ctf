# Reuse Decision

## Change type
service / runtime / test / composition

## Existing code searched
- `code/backend/internal/module/contest/application/commands/`
- `code/backend/internal/module/contest/runtime/`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Similar implementations found
- `code/backend/internal/module/contest/application/commands/challenge_service.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/challenge_service_test.go`

## Decision
refactor_existing

## Reason
这次不是新增 port 或适配器，而是把 `contest` challenge command service 里已经不再使用的 Redis 注入链删掉。代码搜索确认这条依赖只停留在 struct 字段、构造参数和测试 helper，没有实际业务读写。最小正确方案就是直接复用现有 service 和 runtime 结构，删除死参数并同步 allowlist，而不是为了一个未使用字段额外引入新的包装层。

## Files to modify
- `code/backend/internal/module/contest/application/commands/challenge_service.go`
- `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/plan/impl-plan/2026-05-12-contest-challenge-redis-dead-dep-phase5-slice2-implementation-plan.md`

## After implementation
- 如果后续 phase 5 继续清死参数或无效 wiring，可以把“先确认依赖是否已失效，再决定删还是抽 port”的模式追加到长期 reuse 历史。
