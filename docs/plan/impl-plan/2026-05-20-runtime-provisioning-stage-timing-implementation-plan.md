# runtime provisioning stage timing 实施计划

## Objective

为 `ProvisioningService.CreateTopology()` 补齐分阶段耗时与失败日志，能够在压测或故障场景下区分 Docker 网络创建、镜像端口探测、容器创建、容器启动、额外网络连接与网络 IP 探测各阶段的耗时和失败点。

## Non-goals

- 不修改实例创建的超时预算、调度并发或 Docker 网络隔离策略
- 不重做 practice provisioning 与 runtime provisioning 的调用边界
- 不在这刀直接引入删除队列、重试队列或新的资源回收机制

## Inputs

- `docs/architecture/backend/05-key-flows.md`
- `docs/plan/impl-plan/2026-05-20-practice-provisioning-throughput-and-readiness-implementation-plan.md`
- `.harness/reuse-decisions/runtime-provisioning-stage-timing.md`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/service_test.go`
- `/tmp/ctf-backend.log`

## Ownership evaluation

- `practice.Service.provisionInstance()` 继续负责实例创建超时预算、状态推进和失败落库；不负责拆分 Docker 内部分阶段耗时。
- `runtime.ProvisioningService` 负责运行时拓扑创建的编排与观测日志，是这次阶段耗时日志的唯一 owner。
- `runtime engine` 继续只暴露 `CreateNetwork / ResolveServicePort / CreateContainer / StartContainer / ConnectContainerToNetwork / InspectContainerNetworkIPs` 等原子能力；不承担日志聚合。
- `runtime cleaner / delete flow` 这次只做现状核查，不在这份改动里改行为。

## Task slices

1. 在 `CreateTopology()` 内定义统一的阶段日志 helper，固定 `stage / duration / instance_id / node_key / image / network_key / network_name / subnet / host_port / container_id` 等字段口径。
2. 给 `network_create / service_port_resolve / container_create / container_start / connect_extra_networks / inspect_network_ips` 六类阶段补成功和失败日志。
3. 保持错误返回与 cleanup 顺序不变，只增加观测，不改变现有控制流。
4. 在 `runtime/service_test.go` 补日志回归测试，验证成功与失败时都能输出目标阶段日志。
5. 复核孤儿容器与端口冲突现状，确认它们是否属于创建链路超时后的次生问题，并在交付说明中单独写清。

## Data and compatibility impact

- 不改数据库结构、API 契约或实例状态机
- 后端运行日志会新增 runtime provisioning 阶段记录，日志量会随实例创建次数增加
- 现有清理与回滚逻辑保持不变

## Validation

- `go test ./internal/module/runtime/... -count=1`
- `bash scripts/check-consistency.sh`
- 如本地服务可复用，再做一次小规模实例创建 smoke，确认日志中可看到阶段耗时

## Review focus

- 阶段日志是否只在 `runtime provisioning` owner 内收口，没有把日志散落到调用方
- 失败日志是否保留足够上下文字段，能够直接定位网络、镜像、容器或端口阶段
- 日志新增是否不改变现有 cleanup、allocation release 和错误传播路径
- 测试是否覆盖成功与失败两条代表性路径

## Rollback

如果日志量或字段设计带来回归，可以直接回退 `provisioning_service.go` 的阶段日志 helper 与对应测试，不影响实例创建主流程。
