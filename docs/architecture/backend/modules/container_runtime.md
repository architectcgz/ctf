# container_runtime 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/container_runtime/`、`code/backend/internal/app/composition/container_runtime_module.go`
> 替代：无

## 定位

`container_runtime` 是容器执行能力模块，负责 Docker / runtime-agent / sandbox executor / runtime node / allocation 等底层运行时事实和能力，不拥有练习、竞赛、题目或实例业务状态。

## 事实来源

- 用例编排：`code/backend/internal/module/container_runtime/application/`
- 运行时契约：`code/backend/internal/module/container_runtime/contracts/`
- runtime-agent 协议：`code/backend/internal/module/container_runtime/agentcontracts/`
- 端口定义：`code/backend/internal/module/container_runtime/ports/`
- 基础设施：`code/backend/internal/module/container_runtime/infrastructure/`
- 装配视图：`code/backend/internal/app/composition/container_runtime_module.go`

## 当前设计

- `container_runtime/application/commands`
  - 负责：镜像运行能力、容器文件访问、provisioning、runtime cleanup 的命令编排。
  - 不负责：实例是否允许启动、竞赛是否开赛、题目是否发布。
- `container_runtime/application/queries`
  - 负责：容器统计与运行时查询。
  - 不负责：业务面板聚合或教师复盘查询。
- `container_runtime/application/jobs`
  - 负责：runtime node 健康探测、心跳过期、capacity snapshot 和 schedulable 事实维护。
  - 不负责：直接修改 instance / practice / contest 的业务生命周期。
- `container_runtime/infrastructure`
  - 负责：Docker Engine、本地 sandbox executor、runtime-agent client/server、ACL、端口/网络 allocation、runtime metrics、node repository。
  - 不负责：写业务模块表，或决定业务错误码与权限。
- `ContainerRuntimeModule`
  - 负责：向 challenge / contest / ops / instance 暴露显式 runtime capability fields。
  - 不负责：对外暴露一个覆盖所有业务需求的宽 runtime service。

## API 入口设计

`container_runtime` 当前不直接拥有面向用户的 HTTP 路由；它通过 app composition 给其他模块提供 runtime capability。

| 消费入口 | 用户路由来源 | Runtime 能力 |
| --- | --- | --- |
| 题目发布自检 / 镜像探针 | `challenge` authoring 路由 | image runtime、probe、container file runtime |
| 普通训练开题 / 代理访问 | `practice` + `instance` 用户路由 | provisioning runtime、container stats、access URL |
| AWD checker / preview / workspace | `contest` 和 `practice` 路由 | sandbox executor、HTTP/TCP/script checker runtime、node routing |
| 运维 dashboard | `ops` 管理端路由 | runtime stats、runtime node health、container inventory |

runtime-agent server/client 是基础设施通道，不作为公开业务 API 文档入口。

## Application / Service 设计

| Service / Job | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Image runtime service | `container_runtime/application/commands/image_runtime_service.go` | 镜像可用性、运行时镜像探针能力 | 题目发布语义 |
| Provisioning service | `container_runtime/application/commands/provisioning_service.go` | 容器/拓扑创建、端口和网络资源申请 | 训练或竞赛是否允许开题 |
| Container file service | `container_runtime/application/commands/container_file_service.go` | 容器目录和文件读取能力 | 题目附件存储 |
| Runtime cleanup service | `container_runtime/application/commands/runtime_cleanup_service.go` | 清理容器和 runtime 资源 | 实例生命周期状态迁移 |
| Container stats query | `container_runtime/application/queries/container_stats_service.go` | 容器和 runtime 统计查询 | 运维 dashboard 聚合展示 |
| Node health service | `container_runtime/application/jobs/node_health_service.go` | runtime node 心跳、健康状态、capacity snapshot | 自动修改业务实例状态 |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `runtime_nodes` | `container_runtime/entity.RuntimeNode` | runtime 节点身份、心跳、健康、schedulable 和 capacity snapshot | node health service / runtime agent |
| `network_allocations` | `NetworkAllocation` | 容器网络资源分配 | provisioning service |
| `port_allocations` | `PortAllocation` / `allocation_repository.go` | host port 资源分配和释放，表名来自 GORM pluralization 与 migration | provisioning / cleanup |
| Docker Engine | `engine*.go`、`docker_sandbox_executor.go` | 容器、网络、exec、文件、镜像检查 | infrastructure adapter |
| runtime-agent 通道 | `agentclient`、`agentserver`、`agentcontracts` | 跨宿主机执行协议与 bridge | infrastructure adapter |

Redis、PostgreSQL 和 Docker 的 concrete client 都停留在 infrastructure / composition，application 只依赖 `ports`。

## 边界

- `container_runtime` 拥有 runtime node、端口 allocation、网络 allocation、容器 inventory 和 runtime capability。
- `instance` 拥有实例命令、查询、proxy ticket、startup recovery 和访问入口。
- `practice` 拥有训练开题、provisioning scheduler、AWD desired runtime reconciliation。
- `contest` 拥有 AWD 轮次、服务运行态、checker、runtime placement。
- 业务模块只能通过 contracts / ports / app composition 消费容器能力，不能把业务规则写回 runtime module。

## 主要用例

- 创建、停止、清理容器或拓扑容器。
- 读取容器文件、目录、运行时详情和统计。
- 执行镜像探针和运行时 readiness probe。
- runtime-agent bridge 和多 runtime node 路由。
- 维护 runtime node 心跳、健康状态、capacity snapshot 与 schedulable 标记。
- 为 checker / preview / sandbox 执行提供受限容器执行能力。

## 数据与副作用

- PostgreSQL：`runtime_nodes`、`port_allocations`、`network_allocations` 等运行时资源事实。
- Docker Engine：容器、网络、镜像、exec、文件访问和资源限制。
- runtime-agent：跨宿主执行通道，协议由 `agentcontracts` 定义。
- Redis：运行时任务或锁按具体 adapter 使用。
- 宿主机网络：ACL 与端口绑定由 infrastructure 实现，业务模块不直接操作。

## 跨模块依赖

| 消费方 | 用途 | 接线 |
| --- | --- | --- |
| `challenge` | 镜像探针、发布自检、容器文件能力 | `composition.BuildChallengeModule` |
| `instance` | 实例 runtime service、proxy / maintenance 依赖 | `composition.BuildInstanceModule` |
| `practice` | 通过 instance 和 runtime capability 完成训练运行 | `composition.BuildPracticeModule` |
| `contest` | AWD checker、preview、runtime placement | `composition.BuildContestModule` |
| `ops` | runtime stats、dashboard、节点可观测性 | `composition.BuildOpsModule` |

## Guardrail

- `code/backend/internal/module/container_runtime/architecture_test.go`：禁止 runtime 依赖业务 owner 模块。
- `code/backend/internal/app/composition/runtime_node_execution_router_test.go`：覆盖按 `runtime_node_id` 路由和离线节点错误。
- `code/backend/internal/module/container_runtime/infrastructure/docker_sandbox_executor_test.go`：约束 sandbox 容器安全边界。
- `code/backend/internal/module/container_runtime/ports/*`：端口按 capability 拆分，避免宽接口。

## 已知限制

- runtime node failover 只承诺重新入队和重建；不承诺原容器 live migration 或已有 SSH/WebSocket 会话透明迁移。
