# runtime node schedulable is not execution health

## 问题描述

HA T5 的 runtime node health review 发现一个容易混淆的契约：`schedulable=false` 原本只表示不再接收新调度，但初版实现把它也用于显式 `node_id` 旧容器操作和健康扫描过滤。结果是一个健康但被 cordon 的节点会被旧容器访问、清理、checker、文件写入、SSH 等 node-bound 操作判定为不可用，同时又因为不再被健康扫描而不会触发 offline failover。

对应 review：

- `docs/reviews/backend/2026-06-12-backend-review-runtime-node-health-and-failover-rebuild.md`

## 原因分析

- `schedulable` 是 placement eligibility，回答“还能不能接新实例”。
- `health_status + last_seen_at` 是 execution availability，回答“已有容器所在节点还能不能执行原地操作”。
- 健康扫描范围是 observability / failure detection owner，不能直接复用“新调度候选节点”列表；否则 cordoned 节点会从探测面消失。
- inventory / maintenance 视角需要覆盖仍可能承载旧容器的节点，不能只看新调度候选节点。
- review 如果只看“默认 selector 是否跳过故障节点”，会漏掉旧绑定、维护扫描和 failover 触发这三条路径。

## 解决方案

T5 修复后明确拆成三组 contract：

- 新调度：`ListSchedulableHealthyNodes` / `FindSchedulableHealthyNodeByName`，要求 `schedulable=true + ready/degraded + fresh last_seen_at`。
- 显式旧绑定执行：`FindHealthyByID`，只要求 `ready/degraded + fresh last_seen_at`，不要求 `schedulable=true`。
- 健康扫描和 inventory：`ListHealthCheckNodes` 覆盖所有 runtime nodes，让 cordoned 节点继续 heartbeat、offline detection 和 managed container inventory。

新增回归测试覆盖：

- cordoned 健康节点被默认新调度跳过。
- 显式 `node_id` 旧容器操作仍可路由到 cordoned 健康节点。
- offline / stale 的显式绑定节点仍返回不可用，不自动漂移到其他节点。
- 健康服务会探测 unschedulable 节点。
- managed container inventory 会扫描 cordoned 节点。

## 收获

- 状态字段不能因为名字接近就复用成多个 owner；placement、execution、observability、maintenance 是四种不同语义。
- Cordon / drain 类能力必须先定义“停止接新流量”还是“停止所有操作”。如果只是停止新调度，就不能破坏旧绑定操作，也不能退出健康探测。
- 引入健康过滤时，review 要逐条检查默认选择、显式绑定、后台扫描、维护清单和 failover 触发路径，而不是只看一个 selector。
- 对 runtime node / worker / queue / gateway 这类 HA 语义，测试要包含“不可新调度但仍健康”的中间态；只测 healthy / offline 两端不够。

## 沉淀状态

- 状态：implemented
- Owner：先沉淀在项目 `feedback/`，作为 runtime node HA review 复盘；若后续 runtime / worker / gateway review 再出现同类字段语义混用，应上收到共享 `go-backend` 或 `code-reviewer` skill。
- 链接：
  - `docs/reviews/backend/2026-06-12-backend-review-runtime-node-health-and-failover-rebuild.md`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`

## 证据

- commit:
  - `b5c5a668e feat(backend): 支持 runtime node 健康与故障重建`
- review finding:
  - Blocker: explicit node-bound operations were tied to `schedulable=true`.
- fixed behavior:
  - `FindHealthyByID` 不再要求 `schedulable=true`。
  - `ListSchedulableHealthyNodes` 仍保留新调度过滤。
  - `ListHealthCheckNodes` 覆盖所有 runtime nodes。
  - `runtimeNodeExecutionRouter.ListManagedContainers` 改为 all-node inventory。
- reviewer re-validation:
  - `cd code/backend && timeout 90s go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1`
  - `cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1`
  - `cd code/backend && timeout 90s go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeNodeFailover|RuntimeModule' -count=1`

## Decision Log

- 2026-06-12: HA T5 independent review blocked `schedulable` 与 execution health 语义混用。
- 2026-06-13: 实现修复为 scheduling / explicit execution / health scan 三组 repository contract，并补充回归测试。
- 2026-06-13: re-review 通过，剩余仅为文档精度建议。
- 2026-06-13: 将经验记录到项目 `feedback/`，供后续 HA / runtime / worker review 复用。
