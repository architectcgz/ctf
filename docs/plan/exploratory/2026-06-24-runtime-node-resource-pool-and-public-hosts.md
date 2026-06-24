# Runtime Node 资源池与访问地址探索记录

> 状态：已提升为正式实施计划  
> 正式计划：`docs/plan/impl-plan/2026-06-24-runtime-node-resource-pool-public-hosts-implementation-plan.md`

## 定位

本文记录 runtime node 端口 / 网段资源池、node 级访问地址和 provisioning 进度事件的探索结论。它不是当前执行计划；后续实现、review 和 task gate 绑定以正式实施计划为准。

## 已收敛结论

- 端口和 Docker bridge 网段必须先选定 `runtime_node_id`，再从该 node 的资源池中分配。
- 端口唯一性从全局 `port` 收敛为 `(runtime_node_id, port)`。
- 网段唯一性从全局 `subnet` 收敛为 `(runtime_node_id, subnet)`。
- `runtime_nodes.endpoint` 只表示 API/gateway 访问 runtime-agent 的控制面地址。
- `runtime_nodes.public_host` 表示学生访问该 node 发布端口时使用的 host。
- `runtime_nodes.access_host` 表示 API/gateway/checker 内部访问该 node 发布端口时使用的 host。
- 资源池分配应使用数据库行锁，例如 `FOR UPDATE SKIP LOCKED`，避免每次生成候选后再反复查 DB / insert-conflict 探测。
- 实例启动需要同时保存当前 `provisioning_stage`，并通过 append-only `instance_provisioning_events` 记录阶段历史。
- 前端应展示“正在分配访问端口”“正在分配隔离网络”“正在创建靶机容器”“正在重新调度”等细粒度状态。
- 普通实例可以在 retry budget 内 cleanup/quarantine 后自动重调度；AWD service 不应静默单实例漂移到其他 node。

## 范围取舍

第一版不做 runtime node 管理 UI/API。原因是本轮要先稳定 runtime 分配正确性、Access URL 正确性、重调度行为和 provisioning 可观测性；管理 UI/API 会额外引入权限、校验 UX、审计、前端页面和运维工作流。v1 中 node hosts 通过 config / bootstrap / ops seed path 维护，后续可在资源分配模型稳定后单独补管理能力。

## 后续入口

- 正式实施计划：`docs/plan/impl-plan/2026-06-24-runtime-node-resource-pool-public-hosts-implementation-plan.md`
- 需要实现时先绑定 task slug：`runtime-node-resource-pool-public-hosts`
