# AWD runtime node identity and placement follow-up

- Project: `ctf 仓库根目录`
- Created: `2026-06-21T19:03+08:00`
- Status: `Open follow-up`

## Context

本记录沉淀 2026-06-21 对 AWD 多 runtime node 语义的决策：当前 `runtime_node` 在实现上代表平台登记的容器执行节点，通常对应一台运行 Docker daemon / runtime-agent 的宿主；它不是 Docker / Swarm / daemon 内部的 node identity。

AWD 的 `docker_bridge_alias` 依赖同一 Docker 宿主内的 bridge network。同一场 AWD 比赛应固定到一个平台 `runtime_node_id`，防守重启、desired reconcile 和 node failover 补建都不应静默跨 node 漂移。

## Confirmed Decisions

- `instances.node_id` 的语义应迁移为 `runtime_node_id`，引用 `runtime_nodes.id`。
- Docker / daemon / Swarm 内部节点身份必须使用独立字段名，例如 `docker_node_id` 或 `engine_node_id`，不得和平台 `runtime_node_id` 混用。
- 执行面容器引用按 `(runtime_node_id, container_id)` 理解；`container_id` 和 `network_id` 脱离所属 runtime node 不具备全局含义。
- 数据库业务主键继续使用 `instances.id`；复合 runtime 引用用于执行路由、清理、checker、文件写入和审计定位。
- 后续迁移应增加辅助索引：
  - `instances(runtime_node_id, container_id)`
  - `instances(runtime_node_id, network_id)`
- AWD 绑定使用轻量 placement 表承载，例如 `contest_runtime_placements(contest_id, runtime_node_id, status, ...)`，为后续 reservation 扩展留口。
- AWD 防守重启和 desired reconcile 应优先使用 contest 绑定的 `runtime_node_id`；绑定 node 不可用时阻塞或保持 pending，不 fallback 到默认 selector。

## Tracked By Implementation Plan

- [x] Tracked by `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：迁移 `instances.node_id` 到 `instances.runtime_node_id`，同步 Go entity、repository、DTO、测试和文档命名；迁移期间明确旧字段只表示平台 `runtime_nodes.id`。
- [x] Tracked by `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：为 `instances(runtime_node_id, container_id)` 和 `instances(runtime_node_id, network_id)` 增加辅助索引，确保执行面按 runtime node + Docker 对象定位。
- [x] Tracked by `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：新增 AWD contest runtime placement 轻量持久化，记录一场 AWD 比赛绑定的 `runtime_node_id`。
- [x] Tracked by `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：调整 AWD 防守重启和 desired reconcile 的 runtime node 选择：已绑定 contest 只能在绑定 node 上重建；绑定 node 不可用时不自动漂移。
- [x] Tracked by `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：同步更新架构文档和测试，明确 `runtime_node_id + container_id` 复合执行身份、Docker 内部 node id 命名边界、以及 AWD 赛中不自动跨 node 重建的规则。

## Open Items

- [ ] 完整 runtime reservation / capacity preflight 仍未落地。后续只有在明确需要多场 AWD 并发、多 node 自动排布或赛前容量可视化时再立项；立项时应重新确认是否需要 `runtime_node_reservations`、`reserved_units`、按 contest 排除自身 reservation 的容量计算，以及 reservation 生命周期释放。
- [ ] AWD runtime placement 管理 API / UI 仍未落地。当前 active placement 由后端在 AWD scope 首次选择 runtime node 时自动 `ensure`；没有 `preflight-reserve`、候选 node 列表、reserve/rebind 管理端点或管理员容量面板。
- [ ] runtime reservation gate 仍未落地。当前 start、prewarm、round、checker、学生 start/restart、管理员 start/restart 没有强制要求先存在有效 reservation；checker readiness override 也没有和 reservation blocker 绑定。
- [ ] emergency recreate 仍未实现。已有草案计划 `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-emergency-runtime-recreate-implementation-plan.md`，但生产代码中没有 placement replacement、contest-scoped AWD emergency requeue、desired reconcile state cleanup、内部 CLI 或 runbook。
- [ ] degraded / provisioning eligibility 需要在 capacity preflight 或 emergency recreate 立项时重新决策。当前事实是默认新调度允许 `ready/degraded + fresh + schedulable`；旧 6/13 计划中的 `degraded_container_threshold`、`provisioning_eligible = ready + fresh + schedulable` 和 degraded blocking reason 没有作为生产契约落地。
- [ ] reservation 相关文档和 OpenAPI 只在对应能力实现时再补。不要把 preflight reserve、reservation lifecycle、runtime placement UI 或 degraded capacity 规则写成当前事实。

## Covered By Later Work

以下内容已经由后续实现覆盖，不再作为本 todo 的 open item 跟踪：

- `2026-06-21-awd-runtime-node-identity-and-placement`：完成 `instances.node_id` 到 `instances.runtime_node_id` 的迁移、复合执行身份索引、`contest_runtime_placements` 轻量表、AWD scope 固定 active placement，以及绑定 node 不可用时不回退默认 selector。
- `2026-06-24-runtime-node-resource-pool-public-hosts`：完成 per-node `public_host/access_host`、`runtime_port_pool`、`runtime_subnet_pool` 和 node-scoped 资源分配；这解决的是 node 内端口/子网分配正确性，不等同于 contest-level runtime reservation。
