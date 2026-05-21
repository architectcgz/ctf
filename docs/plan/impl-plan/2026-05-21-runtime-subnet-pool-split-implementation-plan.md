# runtime subnet pool split 实施计划

## Objective

把 runtime 动态网络子网池拆成两套：单容器题目走独立小网段池，多容器 topology 走独立大网段池，从而提升单容器题目的地址容量，同时避免不同粒度子网混用产生 overlap。

## Non-goals

- 不修改 Docker 网络创建、DB 预留、Docker occupied subnet 预读与冲突重试的基本机制
- 不调整 shared network、显式 subnet network 的语义
- 不重做 practice / challenge / runtime 的服务边界

## Inputs

- `code/backend/internal/config/config.go`
- `code/backend/configs/config.yaml`
- `code/backend/configs/config.dev.yaml`
- `code/backend/configs/config.prod.yaml`
- `code/backend/internal/module/runtime/ports/topology.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/service_test.go`
- `.harness/reuse-decisions/runtime-subnet-pool-split.md`

## Ownership evaluation

- `ContainerNetworkConfig` 负责声明单容器池与 topology 池的配置边界与校验。
- `ProvisioningService` 继续作为动态子网分配路径的唯一 owner，负责按请求池类型选择 base/mask。
- `TopologyCreateRequest` 只承载内部池类型，不承担配置推导责任。
- `runtime adapters` 负责在单容器适配路径显式标记 `single_container` 池。

## Task slices

1. 扩展 `ContainerNetworkConfig`，移除旧 `jeopardy_subnet_base/subnet_mask`，引入 `single_container_*` 与 `topology_*` 字段。
2. 更新默认配置、dev/prod 配置与配置校验，新增“CIDR overlap”与掩码合法性约束。
3. 在 `TopologyCreateRequest` 中增加内部池类型，默认 topology，单容器路径显式写入 single_container。
4. 修改 `ProvisioningService` 子网分配逻辑，按池类型选择 base/mask，保留现有 occupied subnet 与冲突重试机制。
5. 补充配置校验、runtime adapter、runtime service 回归测试，覆盖单容器池、topology 池和冲突校验路径。

## Data and compatibility impact

- 配置结构发生破坏性变更，旧 `container.network.jeopardy_subnet_base` 与 `container.network.subnet_mask` 不再被读取。
- 新创建的单容器题目会落到独立地址池，例如 `10.11.0.0/16` 的 `/29`。
- 新创建的 topology 题目继续落到独立地址池，例如 `10.10.0.0/16` 的 `/24`。
- 不涉及数据库结构变更。

## Validation

- `go test ./internal/config -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/app/composition/... -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 单容器路径是否都被正确标记为 `single_container`，没有遗漏 challenge runtime probe 这类手工拼装请求的适配层
- topology 默认路径是否仍稳定使用 topology 池
- 配置校验是否能拦住重叠网段和非法掩码
- 现有 shared / explicit subnet / occupied subnet / owner reservation 语义是否未被回归

## Rollback

如果实现带来配置或分配回归，可以整体回退本次字段重构与池类型分流，恢复到单一动态子网池模型。
