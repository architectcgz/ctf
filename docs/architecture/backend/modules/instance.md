# instance 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/instance/`、`code/backend/internal/app/composition/instance_module.go`
> 替代：无

## 定位

`instance` 是靶机实例生命周期和访问入口 owner，负责实例命令、查询、proxy ticket、maintenance、startup recovery 以及 AWD target / defense SSH 访问所需的实例边界。

## 本文档范围

| 本文档负责 | 本文档不负责（见其他文档） |
|-----------|-------------------------|
| instance 模块的职责边界、HTTP 入口和用例组织 | 实例共享策略和访问控制 → `docs/architecture/backend/design/instance-sharing.md` |
| 实例生命周期管理和 proxy 访问的实现 | 容器运行时能力 → `container_runtime` 模块文档 |
| 实例启动恢复和维护任务 | 容器调度和并发控制 → `docs/architecture/backend/design/容器调度与并发控制.md` |
| 模块内部组件协作和数据流 | 竞赛实例编排 → `contest` 模块文档 |

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

## 启动恢复协议

### 本章节范围

| 本文档负责 | 本文档不负责 |
|----------|------------|
| 说明 Instance 启动恢复机制、leader election、boot_id 检测和代码位置 | 记录 Redis lock concrete 实现细节；lock 模式见 runtime 或基础设施文档 |
| 记录恢复流程步骤：暂停补时、lost runtime 修复、AWD 期望态协调 | 替代 startup_runtime_recovery_service.go 代码；冲突时以代码为准 |
| 指向容器调度与并发控制专题文档 | 记录手动暂停赛事的管理命令设计 |

### 启动恢复机制

- **Leader Election**：
  - 进程启动后读取宿主 `/proc/sys/kernel/random/boot_id`（通过 `boot_id_reader.go` 端口）
  - 竞争 Redis 分布式锁 `ctf:platform:runtime:recovery:lock`，成功者成为 recovery leader
  - Leader 续租成功后读取 Redis `platform_runtime_state` 中的上次 `boot_id + last_heartbeat_at`
  - Standby 副本拿不到锁时只等待 leader readiness；看到同 boot_id 且 heartbeat 未过期后才返回启动完成
- **Boot ID 变化检测**：
  - 对比当前 `boot_id` 与上次保存的 `boot_id`
  - 如果不同，说明宿主机重启或进程迁移，需要执行完整恢复流程
  - 如果相同，只更新 heartbeat，跳过恢复步骤
- **恢复流程（boot_id 变化时）**：
  1. **AWD 赛事暂停补时**：按 `last_heartbeat_at -> started_at` 给 active AWD 比赛补暂停时长，并刷新这些比赛下活跃实例的 `expires_at`
  2. **Lost Active Runtime 修复**：调用 `ReconcileLostActiveRuntimes()` 修复停机前仍有 active row 但实际容器已丢失的运行态（标记 `lost`、清理残留或触发重建）
  3. **AWD 期望态补齐**：调用 app composition 注入的 `ReconcileDesiredAWDInstances()` 补齐 AWD `team × visible service` 期望态（确保每个队伍的每个服务都有实例）
  4. **恢复耗时补差**：按 `last_heartbeat_at -> recovered_at` 再补齐恢复耗时；`runtime_recovery_key + runtime_recovery_applied_seconds` 保证重试只补差值，避免重复补时
- **幂等性保证**：
  - `runtime_recovery_key` 基于 `boot_id` 生成，每次宿主重启产生新 key
  - `runtime_recovery_applied_seconds` 记录已补偿秒数，重复执行时只补差值
  - 同一 boot_id 下的多次心跳间隙不触发恢复流程
- **代码位置**：
  - 恢复编排：`code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - Boot ID 读取：`code/backend/internal/module/instance/infrastructure/boot_id_reader.go`
  - Runtime state 存储：`code/backend/internal/module/instance/infrastructure/platform_runtime_state_store.go`
  - Recovery lock：`code/backend/internal/module/instance/infrastructure/startup_runtime_recovery_lock_store.go`
  - Composition 接线：`code/backend/internal/app/composition/instance_module.go`（注入 AWD reconcile adapter）
- **专题文档引用**：完整容器调度与并发控制 → `docs/architecture/features/容器调度与并发控制.md`

## 调度器编排与并发控制

### 本章节范围

| 本文档负责 | 本文档不负责 |
|----------|------------|
| 说明 Instance 维护清理机制、Redis cleanup lock 和代码位置 | 记录 practice provisioning scheduler 的完整实现；scheduler owner 在 practice 模块 |
| 记录 Instance 与 practice/contest 的协作边界 | 记录 Docker concrete 或 runtime-agent 协议细节 |
| 指向容器调度与并发控制专题文档 | 替代专题文档中的并发槽位控制说明 |

### 维护清理机制

- **清理范围**：
  - 过期实例：`expires_at` 已过且状态不是 `stopping` 或 `stopped` 的实例
  - Stopping 实例：状态为 `stopping` 的实例，需要实际销毁容器和网络资源
  - Lost active runtime：停机前仍有 active row 但实际容器已丢失的运行态
- **并发控制**：
  - Redis 分布式锁 `ctf:container:cleanup:lock` 串行化 runtime cleanup
  - 避免多 API 副本重复销毁同一批资源，导致并发冲突或重复日志
  - 锁持有时长默认 30 秒，清理完成后释放
- **清理流程**：
  1. `maintenance_service.go` 定期扫描过期和 stopping 实例
  2. 获取 cleanup lock，成功后执行清理
  3. 调用 `container_runtime` capability 销毁容器和网络
  4. 更新实例状态为 `stopped`，记录清理结果
  5. 释放 cleanup lock
- **代码位置**：
  - 维护服务：`code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - Cleanup lock：`code/backend/internal/module/instance/infrastructure/stopping_cleanup_lock_store.go`
  - Runtime identity 修复：`code/backend/internal/module/instance/application/commands/runtime_maintenance_service.go`

### 与 Practice/Contest 协作边界

- **Instance 职责**：
  - 提供实例生命周期 contract（创建、停止、延期、访问）
  - 维护实例记录、runtime identity、proxy ticket
  - 启动恢复时修复 lost active runtime
  - 清理过期和 stopping 实例
- **Practice 职责**：
  - 拥有训练开题策略和 provisioning scheduler
  - 决定何时领取 `pending` 实例并调用 container runtime 创建容器
  - 消费 instance repository 和 runtime service，但不反向拥有实例访问 contract
- **Contest 职责**：
  - 拥有 AWD runtime placement 和轮次状态
  - 通过 app composition adapter 注入 `ReconcileDesiredAWDInstances()` 到 startup recovery
  - Instance 不导入 contest 生产代码，只通过 composition 接线协作
- **Composition 接线**：
  - `code/backend/internal/app/composition/instance_module.go`：Instance module 组装
  - `code/backend/internal/app/composition/instance_practice_adapters.go`：Practice 与 instance 适配器
  - `code/backend/internal/app/composition/instance_contest_adapters.go`：Contest AWD 期望态协调适配器
- **专题文档引用**：完整调度与并发模型 → `docs/architecture/features/容器调度与并发控制.md`

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
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`：覆盖 boot id 变化、leader/standby、暂停补差和重复恢复幂等。
- `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`：覆盖 lost active runtime reconcile、stopping cleanup 和 runtime identity 修复。

## 已知限制

- `InstanceModule` 仍是 app composition 组合视图，不是独立部署服务。
