# Reuse Decision

## Change type
repository / runtime owner / practice contract convergence

## Existing code searched
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `.harness/reuse-decisions/runtime-port-release-semantics.md`

## Similar implementations found
- `runtime/infrastructure/repository.go`
  - 已经是 `PortAllocation` 的主要持久化 owner，承接 reserve / bind / release 语义。
- `runtime/application/commands/provisioning_service.go`
  - 已经把未绑定端口预留失败回滚收口在 runtime。
- `runtime/application/commands/runtime_cleanup_service.go`
  - 已经把实例运行时清理后的 owner-aware 端口释放收口在 runtime。

## Decision
refactor_existing

## Reason
`PortAllocation` 是 runtime 运行时资源占用表，不应该继续由 `practice/infrastructure/repository.go` 直接 CRUD。`practice` 现在需要的只是“在同一事务里预留/绑定/释放/重启同步宿主机端口”的能力，因此这刀先把 owner 收回 runtime repository，让 `practice` 保留现有事务入口，但内部委托给 tx 绑定的 runtime port source。

这样可以先消除 `practice` 对 `PortAllocation` 持久化形状的直接依赖，后续再把实体文件从 `internal/model` 迁到 runtime 模块时，改面会更小。

## Files to modify
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/runtime/contracts/port_reservation.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/ports/port_reservation.go`
- `code/backend/internal/module/runtime/infrastructure/repository_destroyed_at_test.go`
- `docs/plan/impl-plan/2026-05-18-practice-runtime-port-owner-delegation-implementation-plan.md`

## After implementation
- `practice` 继续通过本地 port interface 工作，但不再自己直接增删改 `PortAllocation`。
- `ResetInstanceRuntimeForRestart` 的端口同步语义由 runtime owner 承接。
- `PortAllocation` 实体文件迁移到 runtime 模块留待下一刀单独处理。
