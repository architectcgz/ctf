# instance 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/instance/`、`code/backend/internal/app/composition/instance_module.go`
> 替代：无

## 定位

`instance` 是靶机实例生命周期和访问入口 owner，负责实例命令、查询、proxy ticket、maintenance、startup recovery 以及 AWD target / defense SSH 访问所需的实例边界。

## 事实来源

- HTTP 入口：`code/backend/internal/module/instance/api/http/handler.go`
- 命令用例：`code/backend/internal/module/instance/application/commands/`
- 查询用例：`code/backend/internal/module/instance/application/queries/`
- 对外契约：`code/backend/internal/module/instance/contracts/`
- 端口定义：`code/backend/internal/module/instance/ports/`
- 持久化与运行状态适配：`code/backend/internal/module/instance/infrastructure/`
- 装配视图：`code/backend/internal/app/composition/instance_module.go`

## 当前设计

- `instance/api/http`
  - 负责：实例访问响应、proxy 访问、教师实例视图等 HTTP 入口。
  - 不负责：直接调用 Docker、Redis lock、runtime-agent 或 contest repository。
- `instance/application/commands`
  - 负责：实例命令、维护清理、startup runtime recovery 编排。
  - 不负责：练习开题策略、竞赛 AWD scope、Docker concrete 或宿主 boot id 文件读取。
- `instance/application/queries`
  - 负责：实例查询和 proxy ticket 查询。
  - 不负责：训练进度、竞赛排行榜或教师复盘聚合。
- `instance/contracts`
  - 负责：实例 service contract、persistence contract、访问 URL、teacher instance DTO、事件和错误。
  - 不负责：承载 runtime adapter 便利方法或业务规则实现。
- `instance/infrastructure`
  - 负责：实例仓储、proxy ticket store、runtime state store、boot id reader、cleaner、ACL migration state、container inventory。
  - 不负责：写 practice / contest 业务状态，或拥有容器执行逻辑。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `GET /api/v1/instances` | `instance/api/http.Handler.ListInstances` | `queries.InstanceService`，当前用户实例列表 |
| `DELETE /api/v1/instances/:id` | `Handler.DestroyInstance` | `commands.InstanceService.DestroyInstance`，标记实例停止并触发清理 |
| `POST /api/v1/instances/:id/extend` | `Handler.ExtendInstance` | `commands.InstanceService`，延长实例有效期 |
| `POST /api/v1/instances/:id/access` | `Handler.AccessInstance` | `commands.InstanceService` / proxy ticket，生成访问入口 |
| `GET /api/v1/instances/:id/proxy`、`ANY /api/v1/instances/:id/proxy/*proxyPath` | `Handler.ProxyInstance` | `queries.ProxyTicketService`，校验 ticket 并代理根路径或子路径流量 |
| `GET /api/v1/teacher/instances` | `Handler.ListTeacherInstances` | teacher instance query，教师实例视图 |
| `DELETE /api/v1/teacher/instances/:id` | `Handler.DestroyTeacherInstance` | `commands.InstanceService.DestroyTeacherInstance`，按教师班级范围或管理员权限停止实例 |
| `POST /api/v1/contests/:id/awd/services/:sid/targets/:team_id/access` | `Handler.AccessAWDTarget` | AWD target access contract |
| `GET /api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy`、`ANY /api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath` | `Handler.ProxyAWDTarget` | proxy ticket + runtime route，支持根路径和子路径代理 |
| `POST /api/v1/contests/:id/awd/services/:sid/defense/ssh` | `Handler.AccessAWDDefenseSSH` | defense SSH access ticket |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Instance command service | `instance/application/commands/instance_service.go` | 实例停止、延期、访问入口、runtime identity 维护 | 训练开题策略、Docker concrete |
| Maintenance service | `instance/application/commands/maintenance_service.go` | 过期、停止中、运行时残留清理 | 业务题目或赛事状态 |
| Startup runtime recovery service | `instance/application/commands/startup_runtime_recovery_service.go` | 启动时 runtime 状态恢复、锁和 boot id 语义编排 | Redis lock concrete、`os.ReadFile` |
| Instance query service | `instance/application/queries/instance_service.go` | 用户/教师实例列表和详情 | 训练进度聚合 |
| Proxy ticket service | `instance/application/queries/proxy_ticket_service.go` | proxy ticket 校验和查询 | token/session 签发 |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `instances` | `instance/entity.Instance` | 实例身份、用户/队伍/竞赛/服务关联、container/network/runtime node、访问 URL、过期和状态 | instance command / practice provisioning |
| Redis proxy ticket | `proxy_ticket_store.go` | 短期访问票据和代理校验状态 | instance command/query |
| Redis runtime state / lock | `platform_runtime_state_store.go`、`stopping_cleanup_lock_store.go` | startup recovery 锁、运行状态恢复和停止清理锁 | startup recovery / maintenance |
| 宿主 boot id | `boot_id_reader.go` | 区分宿主重启前后的 runtime 状态 | startup recovery infrastructure |
| Container inventory | `container_inventory_repository.go` | 从 runtime inventory 辅助恢复实例状态 | instance infrastructure |

`instances.runtime_node_id` 是当前运行时路由 authority；绑定离线节点的旧容器操作返回 runtime node unavailable，不自动漂移到其他节点。

## 边界

- `instance` 拥有实例记录、实例访问、proxy ticket、startup recovery 和 runtime identity 修复。
- `container_runtime` 拥有容器执行能力；`instance` 只通过显式 runtime capability 消费。
- `practice` 拥有训练开题和 provisioning scheduler；它消费 instance repository / runtime service，不反向拥有实例访问 contract。
- `contest` 拥有 AWD runtime placement 和轮次状态；`instance` 不导入 contest 生产代码。

## 主要用例

- 启动、停止、重启、维护和清理实例。
- 创建和校验 proxy ticket，生成访问 URL。
- 教师查看学生实例或实例运行态。
- 进程启动时读取 runtime 状态，恢复旧 active 实例与 runtime identity。
- runtime node offline 后对可恢复 active 实例进行重排准备。

## 数据与副作用

- PostgreSQL：实例记录、运行时 identity、ACL migration state。
- Redis：proxy ticket、启动恢复锁、stopping cleanup lock。
- 宿主机：boot id 由 `boot_id_reader.go` 读取，application 只依赖端口。
- 容器运行时：通过 `ContainerRuntimeModule` 注入的 capability 执行，不在 instance infrastructure 自建 Docker client。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `container_runtime` | 容器执行、inventory、runtime state | `composition.BuildInstanceModule` |
| `practice` | 训练开题消费实例能力 | `composition.BuildPracticeModule` |
| `contest` | AWD runtime state adapter 在 app 层接线 | `code/backend/internal/app/composition/instance_contest_adapters.go` |

## Guardrail

- `code/backend/internal/module/instance/architecture_test.go`：禁止 commands / queries 依赖 HTTP 或 runtime infrastructure，生产代码禁止导入 contest module。
- `code/backend/internal/app/router_route_wiring_test.go`：约束实例路由使用 instance handler。
- `code/backend/internal/app/composition/instance_practice_runtime_adapter_test.go`：覆盖 instance 与 practice runtime adapter。

## 已知限制

- `InstanceModule` 仍是 app composition 组合视图，不是独立部署服务。
