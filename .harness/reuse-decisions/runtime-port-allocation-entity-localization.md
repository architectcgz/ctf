# Reuse Decision

## Change type
entity localization / model compatibility shim / test migration

## Existing code searched
- `code/backend/internal/model/port_allocation.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## Similar implementations found
- `code/backend/internal/module/practice/entity/user_score.go`
  - 已经把持久化实体放回模块内部，外层只保留模块 owner 引用。
- `code/backend/internal/module/runtime/contracts/port_reservation.go`
  - 上一刀已经把 `PortAllocation` 的持久化 owner 收回 runtime，说明实体文件也应该跟着回 runtime。

## Decision
refactor_existing

## Reason
上一刀已经让 `practice` 退出 `PortAllocation` 的直接持久化操作，当前剩余问题主要是实体文件还挂在 `internal/model`。这刀先把实体定义迁到 `runtime/entity`，并把运行中的引用改成 runtime owner 路径；`internal/model/port_allocation.go` 仅保留过渡别名，避免这一刀同时触发删文件确认。

## Files to modify
- `code/backend/internal/model/port_allocation.go`
- `code/backend/internal/module/runtime/entity/port_allocation.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/repository_destroyed_at_test.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/runtime/application/instance_service_test.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `docs/plan/impl-plan/2026-05-18-runtime-port-allocation-entity-localization-implementation-plan.md`

## After implementation
- 运行中代码和测试默认走 `runtime/entity.PortAllocation`
- `internal/model/port_allocation.go` 只作为短期兼容别名，后续删除需单独确认
