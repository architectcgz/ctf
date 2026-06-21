# AWD runtime node identity and placement follow-up

- Project: `ctf 仓库根目录`
- Created: `2026-06-21T19:03+08:00`
- Status: `Partially tracked by impl-plan`

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

- [x] Tracked by `docs/plan/impl-plan/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：迁移 `instances.node_id` 到 `instances.runtime_node_id`，同步 Go entity、repository、DTO、测试和文档命名；迁移期间明确旧字段只表示平台 `runtime_nodes.id`。
- [x] Tracked by `docs/plan/impl-plan/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：为 `instances(runtime_node_id, container_id)` 和 `instances(runtime_node_id, network_id)` 增加辅助索引，确保执行面按 runtime node + Docker 对象定位。
- [x] Tracked by `docs/plan/impl-plan/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：新增 AWD contest runtime placement 轻量持久化，记录一场 AWD 比赛绑定的 `runtime_node_id`。
- [x] Tracked by `docs/plan/impl-plan/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：调整 AWD 防守重启和 desired reconcile 的 runtime node 选择：已绑定 contest 只能在绑定 node 上重建；绑定 node 不可用时不自动漂移。
- [x] Tracked by `docs/plan/impl-plan/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`：同步更新架构文档和测试，明确 `runtime_node_id + container_id` 复合执行身份、Docker 内部 node id 命名边界、以及 AWD 赛中不自动跨 node 重建的规则。

## Open Items

- [ ] 设计 emergency recreate 的产品形态：先决定是内部 service / runbook，还是管理员 API + UI，再实现显式整场切换 node 和全量重建。
- [ ] 评估是否需要完整 runtime reservation / capacity preflight；仅在多场 AWD 并发、多 node 自动排布和容量可视化需求明确后再立项。
