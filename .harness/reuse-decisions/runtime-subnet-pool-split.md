# Reuse Decision

## Change type

config / runtime provisioning / test

## Existing code searched

- `code/backend/internal/config/config.go`
- `code/backend/configs/config.yaml`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/service_test.go`

## Similar implementations found

- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`

## Decision

extend_existing

## Reason

这次不是新增一套独立 allocator 服务，也不是把子网分配逻辑从 runtime provisioning 抽走。现有唯一 owner 仍应保持在 `ProvisioningService`，只是把“动态子网池”从一套拆成两套，并显式区分单容器与 topology 路径。

最小正确方案是：

- 继续复用现有 `ProvisioningService` 作为运行时网络分配编排的唯一 owner
- 扩展现有 `ContainerNetworkConfig`，明确单容器池与 topology 池的 base/mask
- 在 `TopologyCreateRequest` 内增加内部池类型字段，默认 topology，单容器路径显式标为 single_container
- 保持仓储层的 DB 预留、Docker occupied subnet 预读、冲突重试与 cleanup 语义不变，只替换 pool 选择逻辑

这样能在不重做 allocator 架构的前提下，直接把单容器题目的地址池容量从 `/24` 模式提升到 `/29` 模式，同时不影响已有多容器 topology 的隔离假设。

## Files to modify

- `.harness/reuse-decisions/runtime-subnet-pool-split.md`
- `docs/plan/archive/impl-plan/2026-05-21-runtime-subnet-pool-split-implementation-plan.md`
- `code/backend/internal/config/config.go`
- `code/backend/configs/config.yaml`
- `code/backend/configs/config.dev.yaml`
- `code/backend/configs/config.prod.yaml`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/module/runtime/ports/topology.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/runtime/adapters_test.go`
- `code/backend/internal/module/runtime/service_test.go`

## After implementation

- 单容器题目使用独立的小子网池，不再消耗 topology 的 `/24` 网段
- 多容器 topology 继续使用独立的大子网池
- 两套地址池不重叠，避免 `/29` 与 `/24` 混发导致的 Docker overlap
- 配置校验会在启动时阻止错误 CIDR、错误掩码和地址池重叠
