# Reuse Decision

## Change type
service / repository / port / config / migration

## Existing code searched
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
- `code/backend/internal/module/runtime/ports/{container_runtime.go,topology.go,port_reservation.go}`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/domain/topology_runtime.go`
- `code/backend/internal/config/{config.go,config_test.go}`
- `code/backend/configs/config.yaml`
- `code/backend/migrations`
- `docs/architecture/backend/03-container-architecture.md`

## Similar implementations found
- `runtime/infrastructure/repository.go`
  - 已有宿主机端口占用的保留、绑定、释放模式，可以作为子网占用的持久化 owner 参考
- `runtime/application/commands/provisioning_service.go`
  - 已经是网络创建、容器创建和失败回滚的唯一编排 owner，适合把显式子网分配收口在这里
- `runtime/application/commands/runtime_cleanup_service.go`
  - 已有运行时网络清理和端口释放链路，适合补齐子网占用释放，而不是新增独立清理器
- `config.Load()` / `Config.Validate()`
  - 已经是容器运行配置的唯一加载和校验 owner，适合新增 Jeopardy 子网基址和掩码配置

## Decision
extend_existing

## Reason
这次修复的目标是把现有“每实例独立网络”的实现补成显式小子网分配，而不是切换到共享大网段后再用 ACL 人工隔离。仓库里现有的 runtime provisioning、runtime cleanup、runtime repository 和容器配置 owner 已经覆盖了“分配资源、持久化占用、失败回滚、运行时清理”这条链路，正确做法是在这些 owner 内补齐子网分配能力，并把 Docker `CreateNetwork` 改成显式 IPAM。

这样能保持现有 Jeopardy / AWD 语义不变，也能最小化调用面变更。共享网络 + 跨实例 ACL 会引入新的 owner 和新的规则生命周期，不适合作为这次并发失败的首选修复。

## Files to modify
- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/configs/config.yaml`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
- `code/backend/internal/module/runtime/contracts/runtime_details.go`
- `code/backend/internal/module/runtime/ports/{container_runtime.go,topology.go,port_reservation.go}`
- `code/backend/internal/module/runtime/entity/*`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/runtime/infrastructure/repository_destroyed_at_test.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/app/composition/instance_practice_runtime_adapter.go`
- `code/backend/internal/testutil/runtimeadapters/practice_runtime_service.go`
- `code/backend/migrations/000009_create_network_allocations.up.sql`
- `code/backend/migrations/000009_create_network_allocations.down.sql`

## After implementation
- 如果后续还需要把 AWD 共享网络也改为显式 IPAM，再把这次 Jeopardy 子网分配的模式补进 `harness/reuse/history.md`
