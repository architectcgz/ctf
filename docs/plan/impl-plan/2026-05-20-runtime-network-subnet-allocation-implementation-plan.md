# runtime network subnet allocation 实施计划

## Objective

为 Jeopardy runtime network 增加显式 `/24` 子网分配，避免 Docker 默认地址池被大子网快速耗尽，同时保持“每实例独立网络”语义不变。

## Non-goals

- 不把 Jeopardy 改成共享网络 + ACL 隔离
- 不改变 AWD 共享网络语义
- 不重做 runtime / practice 模块分层或事务装配

## Inputs

- `docs/architecture/backend/03-container-architecture.md`
- `.harness/reuse-decisions/runtime-network-subnet-allocation.md`
- `internal/module/runtime/application/commands/provisioning_service.go`
- `internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `internal/module/runtime/infrastructure/repository.go`
- `internal/module/practice/application/commands/runtime_container_create.go`

## Ownership evaluation

- `config` 负责 Jeopardy 子网基址与掩码配置的加载、默认值和校验。
- `runtime provisioning` 负责为非共享 runtime network 申请/回滚显式 subnet，并把 subnet 写入 runtime details。
- `runtime repository` 负责 `network_allocations` 的并发安全持久化、owner 绑定和释放。
- `runtime cleanup` 与实例状态清理路径负责回收 subnet 占用，避免 runtime 销毁后泄漏。
- `practice` 只负责把实例 owner 透传到 runtime topology request；不直接拥有 subnet 分配逻辑。

## Task slices

1. 配置层补 `container.network.jeopardy_subnet_base` / `container.network.subnet_mask` 默认值与校验。
2. runtime contract / entity / migration 补 `network_allocations` 与 `runtime_details.networks[].subnet`。
3. runtime repository 增加 owner-aware subnet reserve / bind / release 能力，并复用现有端口占用模式保证并发安全。
4. provisioning 在 Jeopardy 独立网络路径分配显式 subnet，AWD 共享网络跳过；失败回滚释放 subnet。
5. practice 单容器路径改走单节点 `CreateTopology()`，把 `OwnerInstanceID` 统一带入 runtime provisioning。
6. cleanup 与实例状态流补 subnet 释放，防止 stop / expire / destroyed_at 清理遗漏。

## Data and compatibility impact

- 新增表：`network_allocations`
- `instance.runtime_details` JSON 中的 `networks[]` 新增 `subnet`
- `TopologyCreateRequest` / `TopologyCreateNetwork` 增加 owner 与 subnet 字段，调用方需要跟随更新

## Validation

- `go test ./internal/config -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/module/practice/... -count=1`

## Review focus

- Jeopardy 独立网络是否总是显式使用 `/24` 子网，而 AWD 共享网络没有被误伤
- `network_allocations` 是否在并发 reserve / bind / release 下不会把同一 subnet 分给两个实例
- `practice` 单容器链路是否已不再丢失 owner / subnet / runtime details
- 失败回滚、实例销毁、`ExpireInstanceRuntime`、`UpdateStatusAndReleasePort` 是否都能释放 subnet 占用

## Rollback

如果这刀有回归，可以先停用显式 subnet 分配并回退 `network_allocations` 读写逻辑，再单独重做 owner-aware subnet 预留。
