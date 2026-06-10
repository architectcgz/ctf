# runtime 残余状态 owner 拆分方案（runtime 模块最终退役）

> 状态：In Progress
> 事实源：`code/backend/internal/module/runtime/` 当前实现、`docs/design/backend-module-boundary-target.md` 目标版图、`docs/todos/2026-05-17-project-tech-debt-from-migrations.md` P1 技术债
> 替代：无；各切片落地后，稳定结论回收到 `docs/architecture/backend/07-modular-monolith-refactor.md` 与对应模块边界事实

## Task Metadata

**Goal:** 拆分 `runtime` 旧模块剩余 mixed persistence/state owner，让 `runtime/infrastructure`、`runtime/entity`、`runtime/contracts` 最终清空并退役。

**Architecture:** 按 CTF 后端 modular monolith 边界收口：`container_runtime` owner 容器运行资源与节点状态，`contest` owner AWD 运行态与流量事件，`instance` owner 实例清理与启动恢复状态。

- Task slug: `2026-06-10-runtime-residual-state-owner-split`
- Branch: `task/2026-06-10-runtime-residual-state-owner-split`
- Started at: `2026-06-10T06:48:29Z`
- Worktree path: `/home/azhi/workspace/projects/ctf`
- Plan path: `docs/plan/impl-plan/2026-06-10-runtime-residual-state-owner-split-plan.md`

## Task Classification

- 类型：结构性后端 refactor / owner migration。
- 复杂度：非琐碎、跨模块、受 startup gate 保护。
- 主要风险：事务边界、Redis/cron 并发行为、AWD 类型与仓储 import 迁移、AutoMigrate model 注册遗漏、runtime baseline 边界守卫同步。
- 当前执行策略：按切片推进，先做低风险死代码清理，再迁 node/allocation/Redis 状态；AWD、proxy、inventory、ACL 收尾保持后续独立 reviewable 切片。

## Files

预计触达范围：

- `code/backend/internal/module/runtime/`：删除已确认死代码，逐步迁出 entity / infrastructure / contracts。
- `code/backend/internal/module/container_runtime/`：接收 runtime node、allocation repository，并新建 `entity` 目录承载持久化模型。
- `code/backend/internal/module/instance/`：接收 cleaner、stopping cleanup lock、startup recovery runtime state store。
- `code/backend/internal/module/contest/`：后续接收 AWD workspace / operation / scope 与 proxy traffic owner。
- `code/backend/internal/composition/`：调整模块装配、事务组合和窄 port adapter。
- `docs/architecture/backend/07-modular-monolith-refactor.md`、`docs/todos/2026-05-17-project-tech-debt-from-migrations.md`：完成后回收稳定结论。

## 复用与 Owner 决策

- 块7 `PlatformRuntimeStateStore` 确认归 `instance`，理由是直接消费者是 `instance` 的 startup recovery；boot/heartbeat/leader gate 仍作为实例恢复门控实现细节处理。
- 块3 AWD 类型迁移不保留长期 `runtime/contracts` alias 转发；改完 import 后直接删除 runtime 转发层。
- 块2/块4 迁入 `container_runtime` 时新建 `container_runtime/entity`，避免把持久化模型塞进 infrastructure 本地私有类型。
- 已存在实现优先复用：instance 侧实例契约与 repository、contest 侧 AWD repository / traffic event repository、container_runtime 侧 ports 与 contracts。

## Intake Analysis Gate

- Harness route: `HARNESS`，因为任务是跨模块结构性迁移。
- 使用技能：`executing-plans`、`development-pipeline`、`ctf-go-backend-architecture-reference`、`backend-engineer`、`go-backend`、`runtime-ops-safety`。
- 计划评估：当前 plan 已列出目标 owner、切片顺序、兼容性、验证和 review 重点；用户已确认三个待定 owner / alias / entity 目录决策。
- 执行约束：每个切片保持 reviewable，代码行为、schema、API、Docker / runtime-agent 协议不变。

## Validation

每个已实现切片至少执行窄范围验证，并在最终回复列出实际结果：

- `cd code/backend && go test ./internal/module/<迁入模块>/... -count=1`
- `cd code/backend && go test ./internal/module/runtime/... -count=1`（当 runtime 删除/迁移触达守卫时）
- `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
- `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`

本轮切片 0-3 已执行：

- `cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/composition -count=1`：通过。
- `cd code/backend && go test ./internal/module/practice/... ./internal/module/contest/... ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`：通过。
- `cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/module/practice/... ./internal/module/contest/... ./internal/app/composition ./internal/testutil/systemapp -count=1`：通过。
- `bash scripts/check-startup-gate.sh`：通过。
- `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`：通过。
- `git diff --check`：通过。

## 执行清单

- [x] 切片 0：删除 runtime 死代码并调整测试引用。
- [x] 切片 1：runtime node 迁入 `container_runtime/entity` + infrastructure。
- [x] 切片 2：allocation 迁入 `container_runtime/entity` + infrastructure。
- [x] 切片 3：实例清理与启动恢复 Redis 状态迁入 `instance/infrastructure`。
- [ ] 切片 4：AWD 运行态迁入 `contest`，不保留 runtime alias。
- [ ] 切片 5：proxy traffic recorder 收口到 `contest`。
- [ ] 切片 6：inventory / node-index 拆成 instance + contest providers。
- [ ] 切片 7：ACL 收口、`RuntimeManagedInstance` 消除、runtime 退役。

## 定位

本文档是 runtime 旧模块剩余 mixed persistence/state 的 owner 归属与迁移切片方案，目标是让 `internal/module/runtime/infrastructure` 与 `entity`、`contracts` 最终清空，runtime 模块退役。

- 负责：把 7 块残余状态分别定 owner，给出迁移切片、兼容回退、验证计划和 review 重点。
- 不负责：改外部 HTTP 路由、API DTO、数据库 schema、Docker / runtime-agent 协议、实例生命周期行为；不重做已完成的 `container_runtime` capability 迁移和 `instance` 业务 owner 落地。

## 背景与事实源

这是 boundary-target 阶段 2 的收尾。前序工作已经把容器能力（provisioning/cleanup/file/image/stats/interactive/host-executor）迁到 `internal/module/container_runtime`，实例命令/查询/proxy ticket/maintenance 迁到 `internal/module/instance`，AWD 业务主链路落在 `internal/module/contest`。

runtime 模块当前只剩：

- `runtime/contracts/persistence.go`：`RuntimeManagedInstance`（`instances` 表映射）+ AWD 类型别名转发。
- `runtime/entity/`：`AWDDefenseWorkspace`、`AWDServiceOperation`、`AWDScopeControl`、`RuntimeNode`、`PortAllocation`、`NetworkAllocation`。
- `runtime/infrastructure/`：13 个 repository / store / cleaner。
- `runtime/ports/`、`runtime/doc.go` 与一批 `service_*_test.go` 行为测试。

`runtime/architecture_test.go` 已经是防回流守卫（`TestLegacyRuntimeModulePackageIsRetired`、`TestRootPackageKeepsOnlyDocFile`、`TestRuntimeInfrastructureDoesNotOwnInstanceProxyTickets` 等）。当前 `moduleDependencyBaseline` 仍保留 `contest -> runtime`、`practice -> runtime`、`runtime -> container_runtime`、`runtime -> contest` 四条边，正是本方案要消解的对象。

## 目标

- 让 7 块残余状态各自落到唯一业务/平台 owner，`runtime/infrastructure`、`runtime/entity`、`runtime/contracts` 清空，runtime 模块退役或只保留顶层防复活守卫。
- 消解 `runtime -> contest`、`runtime -> container_runtime`、`contest -> runtime`、`practice -> runtime` 四条 baseline 边。
- 消除 runtime 与 contest 之间的重复实现（AWD service operation / proxy traffic）。
- 删除已确认的死代码（`ManagedInstanceRepository`、`awd_scope_control_query.go`）。
- 全过程不改 schema、API、路由、Docker / agent 协议和实例生命周期行为。

## 非目标

- 不拆微服务、不动部署形态，继续单进程模块化单体。
- 不在本方案内重构 `container_runtime` 容器能力或 `contest` AWD 业务规则本身。
- 不引入新的 `internal/shared` 持久化层来"中转" runtime 状态。
- 不为目录整齐而给目标模块机械新建空 `domain` / `ports` / `contracts`。
- 不一次性大提交；每块按可 review 切片推进，每个实现切片各自走 `bash scripts/start-implementation.sh` 绑定 task slug。

## 输入与现状

以下是 7 块逐项现状，全部来自当前代码事实。

### 块1：instances 运行态视图

- `RuntimeManagedInstance`（`code/backend/internal/module/runtime/contracts/persistence.go:9`）：`instances` 表的 GORM 映射，与 `instance` 模块自有的 `instancecontracts.Instance` + `instanceinfra.Repository` 功能重复。
- `ManagedInstanceRepository`（`runtime/infrastructure/managed_instance_repository.go`）：只有 `FindByID`。生产代码零调用，仅 3 处测试引用（`router_composition_typed_deps_test.go` 源码守卫、`practice/.../instance_start_service_test.go`、`runtime/service_test.go` stub）。**死代码**。
- ACL migration：`ACLMigrationStateRepository`（`runtime/infrastructure/acl_migration_state_repository.go`，`ListInstancesNeedingACLHandleMigration` + `UpdateInstanceRuntimeDetails`）+ `composition/runtime_acl_migration.go`，在 `BuildContainerRuntimeModule` 启动时把 legacy ACL 规则迁成 ACL handle，解码用 `container_runtime/contracts.DecodeInstanceRuntimeDetails`，读写 `instances.runtime_details`。这是仍在生产执行的一次性 bootstrap 迁移。

### 块2：port/subnet allocation

- `AllocationRepository`（`runtime/infrastructure/allocation_repository.go`）：端口预留/绑定/释放 + 子网预留/释放，依赖 `runtime/entity.PortAllocation`（`port_allocations`）、`NetworkAllocation`（`network_allocations`），二者绑 `instance_id`。
- 主消费者：`container_runtime.Build`（`ProvisioningRepository` / `CleanupRepository`，见 `composition/runtime_module.go:66`）、`runtimeNodeExecutionRouter`、以及 instance lifecycle 事务里的释放路径（`composition/instance_module.go` 的 `withInstanceRuntimeLifecycleTx` → `ReleaseRuntimeAllocationsForInstance`）。
- 性质：端口/子网是容器运行资源分配；绑 `instance_id` 只是 owner key，不构成对 instance 业务语义的依赖。

### 块3：AWD 运行态持久化

- `AWDRepository`（`runtime/infrastructure/awd_repository.go`）：`awd_defense_workspaces` + `awd_service_operations` 读写。
- entity：`AWDDefenseWorkspace`、`AWDServiceOperation`、`AWDScopeControl`（`runtime/entity/`），维度全是 `contest_id/team_id/service_id/instance_id`。
- `awd_scope_control_query.go`：包级 helper `joinAWDActiveScopeControls` / `applyAWDActiveScopeFilter`，runtime 包内**零调用方**（`instance/infrastructure` 已有自用同名副本并在使用）。**死代码**。
- contest 已是成熟 AWD owner：`contest/infrastructure/` 已有 `awd_service_operation_repository.go`（contest 版 `AWDRepository.ListLatestServiceOperationsByContest` / `HasSystemRecoveryOperationAt`）、`awd_query_repository.go`、`awd_round_repository.go`、`awd_service_instance_repository.go`、`contest_awd_runtime_recovery_repository.go` 等一整套。
- runtime AWDRepository 的消费方：instance maintenance（`FindRunningAWDDefenseWorkspaceByInstanceID`、`CreateAWDServiceOperation`、`FinishAWDServiceOperation`、`FinishActiveAWDServiceOperationForInstance`，经 `instanceMaintenanceRepositoryAdapter` / `practiceInstanceRepositoryAdapter` 转成 `instanceports` 类型）；practice 通过 `runtime/contracts` 的 AWD 类型别名消费。
- 跨模块牵连：`runtime/contracts` 的 AWD 类型/常量被 `contest/ports/awd.go`、`contest/infrastructure/awd_service_operation_repository.go`、`practice/application/commands/awd_*.go`、`practice/ports/ports.go` 大量依赖，对应 baseline 的 `contest -> runtime`、`practice -> runtime` 边。

### 块4：runtime node state

- `RuntimeNode`（`runtime/entity/runtime_node.go`，`runtime_nodes`）：平台节点注册/调度元数据（endpoint/tls/schedulable/labels/health/capacity）。
- `RuntimeNodeRepository` + `defaultRuntimeNodeSelector`（`runtime/infrastructure/node_repository.go`）：**已实现 `container_runtime/ports.RuntimeNodeSelector`，已依赖 `container_runtime/contracts`**（`RuntimeNodeBootstrapSpec`、`RuntimeNodeBinding`）。
- 背景：`2026-06-02-runtime-control-plane-agent-split-plan.md` 把节点选择/执行面归在 container_runtime 范畴，API 节点只做节点选择与编排。

### 块5：跨状态查询聚合

- `ActiveContainerInventoryRepository.ListActiveContainerIDs`（`runtime/infrastructure/active_container_inventory_repository.go`）：跨 `instances`（含 `runtime_details` 解码出的 containers）+ `awd_defense_workspaces` 收集活跃容器 ID，给 orphan 清理用。
- `ContainerNodeIndexRepository.FindRuntimeNodeIDByContainerID`（`runtime/infrastructure/container_node_index_repository.go`）：容器 ID → node ID 反查，同样跨 `instances` + `awd_defense_workspaces`，给 `runtimeNodeExecutionRouter` 路由用。
- 性质：两个 repo 都"一个查询跨两个 owner 表"，属 boundary-target 指出的过宽聚合查询。

### 块6：proxy traffic recorder

- `ProxyTrafficEventRecorder`（`runtime/infrastructure/proxy_traffic_recorder.go`）：`RecordRuntimeProxyTrafficEvent`（实例 proxy 入口）+ `RecordAWDProxyTrafficEvent`（AWD proxy 入口），**两个方法最终都写 `contestcontracts.AWDTrafficEvent`**，整体依赖 `contest/contracts`，对应 baseline 的 `runtime -> contest` 边。
- 重复实现：`contest/infrastructure/awd_traffic_event_repository.go` 的 contest 版 `AWDRepository.RecordRuntimeProxyTrafficEvent` 与之**几乎逐行相同**（写 `contestentity.AWDTrafficEvent`），并额外有 `ListTrafficEvents`。
- 事实澄清：当前不存在独立的"普通实例访问流量"持久化；所有 proxy traffic 都是 AWD 攻击流量事件（`awd_traffic_events`）。instance 侧只持有代理入口（proxy ticket、proxy handler、`instance/infrastructure/awd_target_proxy_repository.go`），不 own 流量事件持久化。

### 块7：Redis/job 状态

- `StoppingCleanupLockStore`（`runtime/infrastructure/stopping_cleanup_lock_store.go`）+ `Cleaner`（`runtime/infrastructure/cleaner.go`，cron）：驱动 `cleanerService`（`ReconcileLostActiveRuntimes` / `CleanExpiredInstances` / `CleanupOrphans`），实现者是 `instancecmd.NewInstanceMaintenanceService`（见 `composition/instance_module.go:113`）。明确服务实例清理。
- `PlatformRuntimeStateStore`（`runtime/infrastructure/platform_runtime_state_store.go`）：`boot_id` / `last_heartbeat_at` / startup recovery lock，服务 `instancecmd.NewStartupRuntimeRecoveryService`（消费者在 `instance/startuprecovery`）。
- Redis key 集中在 `runtime/infrastructure/cachekeys/redis_keys.go`。

## Owner 决策总表

| 块 | 对象 | 目标 owner | 处理方式 | 理由 |
| --- | --- | --- | --- | --- |
| 1 | `ManagedInstanceRepository` | — | 删除 | 死代码，仅测试引用 |
| 1 | `RuntimeManagedInstance` 表映射 | `instance` | 消除，统一用 `instancecontracts.Instance` / `instanceinfra` | 与 instance 自有实例契约重复 |
| 1 | ACL handle migration | `container_runtime`（bootstrap）+ `instance`（窄读写 port） | ACL 迁移逻辑归 container_runtime，`instances` 读写经 instance 窄 port | ACL 是容器适配能力；实例数据归 instance owner |
| 2 | allocation（port/subnet） | `container_runtime` | 迁 `AllocationRepository` + 两个 entity；释放经窄 port，由 composition lifecycle tx 编排 | 端口/子网是容器运行资源，主消费者已在 container_runtime |
| 3 | AWD workspace/operation/scope | `contest` | 消除 runtime 侧，合并进 contest 已有 AWD 仓储；entity/类型迁 contest，instance 需要的窄能力经 contest port 适配 | contest 已是 AWD owner 且有等价实现 |
| 3 | `awd_scope_control_query.go` | — | 删除 | 死代码 |
| 4 | runtime node | `container_runtime` | 迁 `RuntimeNode` entity + repository + selector | 平台节点注册/调度状态，selector 已实现 container_runtime/ports |
| 5 | inventory / node-index | `instance` + `contest` 双 provider | 拆成实例容器 provider（instance）+ AWD workspace 容器 provider（contest），在 cleanup / nodeRouter 组合 | 消除跨 owner 表的过宽聚合查询 |
| 6 | proxy traffic recorder | `contest` | 消除 runtime 侧重复，收口 contest `awd_traffic_event_repository`；instance proxy 入口经 contest 窄 port 记录 | 所有 proxy traffic 都是 AWD 流量事件 |
| 7 | StoppingCleanupLockStore + Cleaner | `instance` | 迁 `instance/infrastructure` | 实例清理是 instance owner 职责 |
| 7 | PlatformRuntimeStateStore | `instance` | 跟随 startupRecovery 消费者迁 `instance` | 消费者是 instance startupRecovery；boot/heartbeat/leader gate 在本阶段作为实例启动恢复门控处理 |

已确认项：

- 块7 `PlatformRuntimeStateStore` 归 `instance`。
- 块3 AWD 类型迁移不保留 `runtime/contracts` type alias 转发，import 改完即删。
- `container_runtime` 接收块2/块4 时新建 `container_runtime/entity`。

## 阶段切片

按风险从低到高、依赖从被依赖到依赖排序。每个切片是一个独立可 review 的实现任务，实现时各自 `bash scripts/start-implementation.sh <slug>` 绑定 task slug 与 startup gate。

### 切片 0：清理死代码

- 删除 `runtime/infrastructure/managed_instance_repository.go`、`runtime/infrastructure/awd_scope_control_query.go`。
- 调整 3 处测试引用（改用 instance 侧等价或直接移除 stub 字段）。
- 风险：极低，无生产调用方。

### 切片 1：runtime node → container_runtime（块4）

- 迁 `RuntimeNode` entity 到 `container_runtime/entity`，迁 `RuntimeNodeRepository` + `defaultRuntimeNodeSelector` 到 `container_runtime/infrastructure`。
- selector 已实现 `container_runtime/ports.RuntimeNodeSelector`，迁移后 `composition/runtime_module.go` 的 `buildDefaultRuntimeNodeSelector` 改依赖 container_runtime 包。
- 风险：低，依赖方向已对齐。

### 切片 2：allocation → container_runtime（块2）

- 迁 `PortAllocation` + `NetworkAllocation` 到 `container_runtime/entity`，迁 `AllocationRepository` 到 `container_runtime/infrastructure`。
- 暴露窄 release 能力给 composition lifecycle tx；`withInstanceRuntimeLifecycleTx` 继续在同一事务里编排 instance 状态写入 + allocation release。
- 风险：中，要保证 `WithDB` 事务句柄替换语义和 instance 状态释放的同事务性不变。

### 切片 3：实例清理 + 启动恢复 Redis 状态 → instance（块7）

- 迁 `StoppingCleanupLockStore` + `Cleaner` 到 `instance/infrastructure`，相关 cachekeys 一并迁移。
- 迁 `PlatformRuntimeStateStore` 到 `instance/infrastructure`（按待确认结论，必要时改 `ops`）。
- `composition/instance_module.go` 的 cleaner / startupRecovery / stopping cleanup 装配改依赖 instance 包。
- 风险：中，注意 cron 行为、锁 keepalive、leader recovery gate 行为不变（参考 `multi-instance-startup-recovery-gate-fix`、`runtime-stopping-cleanup-optimization`）。

### 切片 4：AWD 运行态 → contest（块3，最大切片）

- entity：`AWDDefenseWorkspace`、`AWDServiceOperation`、`AWDScopeControl` 迁 `contest/entity`（与现有 `contest/entity/awd.go` 合并）。
- 类型/常量：`runtime/contracts` 的 AWD 别名迁 `contest/contracts`；`practice`、`contest` 改 import contest/contracts，消解 `contest -> runtime`、`practice -> runtime`。
- 仓储：runtime `AWDRepository` 的方法合并进 contest 已有 AWD 仓储（去重）；instance maintenance 需要的窄能力（按 instance_id 查 running workspace、建/结 service operation）通过 `contest/ports` 暴露窄 port，由 composition 适配进 instance maintenance，instance 不直接依赖 contest infrastructure。
- 风险：高，跨 practice/contest/instance 三方消费，注意 `BumpAWDDefenseWorkspaceRevision` 的事务语义和那条函数级 baseline。

### 切片 5：proxy traffic → contest（块6）

- 消除 `runtime/infrastructure/proxy_traffic_recorder.go`，能力收口到 contest `awd_traffic_event_repository`（已存在等价实现）。
- instance proxy 入口（`InstanceModule` 的 `runtimeProxyTrafficRecorder`）改经 `contest/ports` 窄 port 记录，由 composition 适配，消解 `runtime -> contest`。
- 风险：中，注意"实例 proxy 入口反查 contest scope"逻辑收口到 contest 后行为一致。

### 切片 6：inventory / node-index 拆 provider（块5）

- 拆 `ActiveContainerInventoryRepository`：实例容器枚举 provider 归 instance（扫 `instances`），AWD workspace 容器枚举 provider 归 contest（扫 `awd_defense_workspaces`）。
- 拆 `ContainerNodeIndexRepository`：容器→节点反查由 `runtimeNodeExecutionRouter`（container_runtime）通过 instance / contest 各自窄 lookup port 组合。
- orphan 清理的"活跃容器全集"在 instance maintenance / composition 组合两个 provider。
- 依赖切片 1/4 已落地（node、awd workspace 已有清晰 owner）。
- 风险：中高，注意去重逻辑与 LIKE 匹配 runtime_details 的行为保持。

### 切片 7：ACL 收口 + RuntimeManagedInstance 消除 + runtime 退役（块1 收尾）

- ACL handle migration：迁移逻辑归 container_runtime bootstrap，`instances` 读写改用 instance 窄 port（list 待迁实例 + 回写 runtime_details）。
- 消除 `RuntimeManagedInstance` 表映射，调用点改用 `instancecontracts.Instance`。
- 此时 `runtime/infrastructure`、`runtime/entity`、`runtime/contracts` 应已清空；删除 runtime 模块剩余文件与 `service_*_test.go`（或迁移有效行为测试到对应 owner）。
- 把 `runtime/architecture_test.go` 的防回流守卫上移为顶层 `internal/module` 架构测试（防止 runtime 模块复活），消解 `runtime -> container_runtime` 等剩余 baseline 边。

## 兼容性与回退

- 行为不变：全过程不改表名/列名、API/路由、Docker / runtime-agent 协议、实例生命周期与 AWD 业务行为；只换 GORM 映射与仓储的 owner 包。
- schema 不变：纯代码 owner 迁移，不新增 migration。**验证点**：确认 GORM AutoMigrate 的 model 注册列表（`ctf-api-auto-migrate` 相关装配）在 entity 包迁移后仍注册全部表，不漏 `port_allocations` / `network_allocations` / `runtime_nodes` / `awd_*`。
- 类型别名：块3 迁移不保留 `runtime/contracts` type alias 转发，import 改完即删，不长期留兼容 wrapper（符合 boundary 完成标准）。
- 事务边界：块2 释放路径必须保持 instance 状态写入与 allocation release 同事务（`withInstanceRuntimeLifecycleTx`）。
- 回退：每个切片是独立 commit/worktree，可单独 revert；切片间无强制不可逆耦合，顺序可在"死代码 → node → allocation → 清理状态 → AWD → proxy → inventory → ACL 收尾"范围内微调，但块5 必须在块3/块4 之后、块7 必须最后。

## 验证计划

每个切片至少执行（最小充分范围）：

```bash
cd code/backend && go test ./internal/module/<迁入模块>/... -count=1
cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1
cd code/backend && bash ../../scripts/check-backend-architecture.sh --full
```

涉及 mapper / DTO / contract 边界迁移的切片，按项目 Completion Validation Gate 追加：

```bash
cd code/backend && go generate ./...   # 至少覆盖触达的 mapper 包
cd code/backend && go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1
```

整体收尾（切片 7）追加：

- `go test ./internal/module/runtime/...`：确认 teardown 守卫通过 / runtime 已删除。
- `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`。
- 确认 `moduleDependencyBaseline` 中 `runtime -> *`、`* -> runtime` 边已消解。
- 数据库 AutoMigrate / 启动恢复 / cleaner cron 行为回归（runtime 集成测试目录 `code/backend/tests/runtime`）。

## Review 重点

- owner 方向正确：不得把竞赛/用户业务状态塞进 `container_runtime`；node/allocation 只作为平台容器资源适配落入。
- 去重彻底：块3/块6 必须删 runtime 侧实现，收口到 contest 已有 owner，不能再平行留一份。
- 死代码确实删除而非搬迁：块1 `ManagedInstanceRepository`、块3 `awd_scope_control_query.go`。
- 事务与并发：allocation release 同事务、cleaner 锁 keepalive、startup recovery leader gate 行为不变。
- 边界不反向：instance/practice 不直接依赖 contest/container_runtime 的 infrastructure，跨模块只经 contract / 窄 port。
- 类型别名不长期残留；baseline 边按真实 import 消失而移除，不靠放宽守卫。
- schema / AutoMigrate 不丢表。

## 完成判定

- `runtime/infrastructure`、`runtime/entity`、`runtime/contracts` 清空，runtime 模块退役或只剩顶层防复活守卫。
- 7 块各自落到 `instance` / `container_runtime` / `contest`（块7 PlatformRuntimeStateStore 按最终结论落 instance 或 ops），无 mixed owner 残留。
- `moduleDependencyBaseline` 中 runtime 相关四条边消解。
- contest 与 runtime 的 AWD service operation / proxy traffic 重复实现合并为单一 owner。
- 全部切片验证通过，行为、schema、API、协议无变化。
- 稳定结论回收到 `docs/architecture/backend/07-modular-monolith-refactor.md`，并更新 `docs/todos/2026-05-17-project-tech-debt-from-migrations.md` P1 条目。
