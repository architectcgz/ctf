<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# Runtime Node Health 与 Failover Rebuild Implementation Plan

> Status: Superseded. Current task delivery plan is `docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md`; this earlier task-group slice draft is kept only for historical planning context.

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 把 `runtime_nodes` 从“可选 metadata + first schedulable selector”升级为运行时可用性 owner：维护 node heartbeat / health / capacity，调度排除失效 node，并在 node 失联后把其上 active runtime 标记为可重建，由现有 actual / desired runtime owner 在健康 node 上重建。

**Architecture:** 复用当前 `runtime-agent` 被动 Health RPC 与 `runtime_nodes` 数据模型，先采用控制面 polling health job，而不是新增 agent 主动注册协议。`container_runtime` 拥有 node health / selector predicate，`instance` 拥有 lost active runtime 标记与 requeue，`practice` 拥有 pending instance provisioning 与 AWD desired reconcile；failover rebuild 不承诺容器 live migration 或 SSH/WebSocket 会话透明迁移。

**Tech Stack:** Go, GORM, PostgreSQL migration, runtime-agent Health RPC, Redis distributed lock / keepalive, existing instance maintenance and practice scheduler, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-runtime-node-health-and-failover-rebuild`
- Started At: `2026-06-12T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-runtime-node-health-and-failover-rebuild`
- Branch: `task/2026-06-12-runtime-node-health-and-failover-rebuild`

## Objective And Non-Goals

- Objective:
  - 为 `runtime_nodes` 增加 `last_seen_at` 并使用现有 `health_status`、`capacity_snapshot` 字段。
  - 新增 node health polling / evaluator job：周期性调用 runtime-agent Health RPC，更新 ready/degraded/offline 与 last seen。
  - 改造 selector：新实例只选 healthy + schedulable node；显式绑定 node 在 provisioning 时也要检查健康状态。
  - node offline 后标记该 node 上 active runtime 为 lost candidate 或 requeue，并确保重建不会再次回到 offline node。
  - 区分 Jeopardy 与 AWD：Jeopardy 只重建已存在 active instance；AWD 继续由 desired reconciler 补齐 contest/team/service scope。
- Non-Goals:
  - 不实现容器 live migration、内存态迁移或 TCP/SSH/WebSocket 会话透明漂移。
  - 不引入 Kubernetes / Nomad / 外部调度器。
  - 不实现 runtime-agent 主动注册/上报 API；第一版使用控制面 polling。
  - 不把 Jeopardy 改成有 desired owner 的自动创建模型。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-02-runtime-control-plane-agent-split-plan.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Related architecture/contracts:
  - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `code/backend/internal/infrastructure/redislock/lock.go`
  - `code/backend/internal/shared/lockkeepalive/lockkeepalive.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-02-runtime-control-plane-agent-split-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`

## Task Classification

- Classification: `结构性改动 / 非琐碎任务`
- Why:
  - 触达数据库 schema、node repository、runtime routing、practice provisioning、instance maintenance、startup recovery 和 runtime-agent health。
  - 这是真正执行面 failover 的核心切片，错误会导致实例重建到失效 node 或把不可迁移会话误写成已恢复。
  - 需要跨 actual runtime reconciliation 与 AWD desired reconciliation 明确 owner 边界。

## Files

- Create:
  - `code/backend/migrations/000019_add_runtime_node_last_seen.up.sql`
  - `code/backend/migrations/000019_add_runtime_node_last_seen.down.sql`
  - `code/backend/internal/module/container_runtime/application/node_health_monitor.go`
  - `code/backend/internal/module/container_runtime/application/node_health_monitor_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
- Modify:
  - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/ports/node.go`
  - `code/backend/internal/app/runtime_node_migration_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Review:
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- Test:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - `code/backend/internal/module/container_runtime/application/node_health_monitor_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/app/runtime_node_migration_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `runtime_nodes` 已有 `schedulable`、`health_status`、`capacity_snapshot`，但没有 `last_seen_at`，selector 也未使用 health。
  - `runtime-agent` 已有被动 Health RPC，返回 `Ready` 与 capabilities。
  - `runtime_node_execution_router` 显式 `nodeID > 0` 时只 `FindByID`，不校验健康。
  - `InstanceMaintenanceService.ReconcileLostActiveRuntimes` 已能把丢失 runtime requeue pending，但当前 `RequeueLostRuntime` 不清 `node_id`。
  - cleaner、startup recovery、practice scheduler 已有 Redis lock / keepalive 模式。
- Reuse / extend / split / create-new decision:
  - 复用现有 `health_status` / `capacity_snapshot` 字段，新增 `last_seen_at`。
  - 复用 runtime-agent Health RPC，由控制面 polling 更新 node 状态，不新增 agent push API。
  - 复用 `redislock` + `lockkeepalive` 给 node health monitor 做单 owner job。
  - 扩展 instance maintenance requeue，而不是新增一套平行 runtime rebuild service。
- Owner boundary:
  - `container_runtime`：node health facts、healthy schedulable predicate、agent health polling owner。
  - `app/composition`：runtime execution routing 与 client cache owner。
  - `instance`：active runtime lost 标记 / requeue / cleanup owner。
  - `practice`：new instance selection、pending provisioning、AWD desired reconcile owner。
- Why this is the narrowest safe surface:
  - 不改变 runtime-agent 协议形态，不引入外部 orchestrator；用现有 node_id authority 和 maintenance/reconcile 链路完成故障后重建。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
  - `dispatching-parallel-agents`
- Why this pass fits:
  - runtime failover 是 umbrella plan 最难切片，必须先厘清已有 runtime_nodes、startup recovery、actual/desired reconciler 和 Redis lock 模式。
- grill-with-docs findings:
  - `runtime_nodes` schema 已有 health/capacity 概念，但没有 last_seen 和 selector health predicate，说明当前只是 metadata，不是可用性 owner。
  - 当前 actual lost runtime 会直接 requeue，但不清 `node_id`，node offline 后可能重建回同一失效 node。
  - AWD 有 desired owner；Jeopardy 没有 desired owner，不能承诺自动补齐不存在的 Jeopardy 实例。
- Plan adjustments after challenge:
  - 明确第一版用 control-plane polling，不新增 agent 主动 heartbeat。
  - 把“清 node_id 或选择 replacement node”列为 failover rebuild blocker。
  - 文档明确 session 可中断，不承诺透明迁移。

## Ordered Task Slices

### Slice 1: runtime node health data contract

- [ ] **Step 1: 写 migration contract 测试**
  - Modify: `code/backend/internal/app/runtime_node_migration_test.go`
  - 断言 `runtime_nodes.last_seen_at timestamp with time zone` 存在；已有 `health_status/capacity_snapshot` 保持。

- [ ] **Step 2: 新增 last_seen_at migration**
  - Create: `code/backend/migrations/000019_add_runtime_node_last_seen.up.sql`
  - Create: `code/backend/migrations/000019_add_runtime_node_last_seen.down.sql`
  - 字段 nullable；新增必要 index。

- [ ] **Step 3: 更新 entity 与 repository mapping**
  - Modify: `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `LastSeenAt *time.Time`，所有写入时间 `.UTC()`。

- [ ] **Step 4: 写 node repository health predicate 测试**
  - Create: `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - 覆盖 ready + schedulable + fresh last_seen 可选；offline/degraded/stale/unschedulable 被排除。

- [ ] **Step 5: 实现 healthy schedulable 查询方法**
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - 增加 `FindFirstHealthySchedulableNode`、`FindHealthySchedulableNodeByName/ID`、`ListHealthySchedulableNodes` 或等价方法。

### Slice 2: health monitor / evaluator job

- [ ] **Step 6: 写 node health monitor 测试**
  - Create: `code/backend/internal/module/container_runtime/application/node_health_monitor_test.go`
  - 覆盖 Health RPC success 更新 ready/last_seen/capacity，failure/stale 标记 offline，Redis lock 已被占用时跳过。

- [ ] **Step 7: 实现 node health monitor**
  - Create: `code/backend/internal/module/container_runtime/application/node_health_monitor.go`
  - 遍历 nodes，调用 agent client `Health(ctx)`；成功写 last_seen/health/capacity；失败按阈值标记 degraded/offline。

- [ ] **Step 8: 接入 Redis lock / keepalive**
  - Modify: health monitor 或 runtime module wiring。
  - 复用 `redislock.Acquire/Refresh/Release` 与 `lockkeepalive.Start`，确保多 API 副本只有一个 monitor owner 执行标记。

- [ ] **Step 9: 接入 composition lifecycle**
  - Modify: `code/backend/internal/app/composition/container_runtime_module.go`
  - root background job 启动 monitor；ctx 来自 root，禁止内部 `context.Background()`。

### Slice 3: selector and router health enforcement

- [ ] **Step 10: 写 selector 排除 unhealthy node 测试**
  - Modify/Create: `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - 覆盖新实例选择跳过低 ID offline node；显式 unhealthy node 在 provisioning path 返回 unavailable。

- [ ] **Step 11: 改 default selector predicate**
  - Modify: `code/backend/internal/module/container_runtime/ports/node.go`
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - Modify: `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - 新建/provisioning 只选 healthy + schedulable。

- [ ] **Step 12: 区分 existing runtime operation 与 provisioning operation**
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - existing cleanup/inspect 可按原 node 尝试；new create/provisioning 必须拒绝 unhealthy 或重新选择 healthy node。

- [ ] **Step 13: 处理 client cache 与 offline node**
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - node offline 后不继续用 cached client 做新建 runtime；cleanup 是否尝试原 node要显式注释/测试。

### Slice 4: node offline lost marker and rebuild

- [ ] **Step 14: 写 node offline requeue 测试**
  - Modify: `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - node offline 后该 node 上 active running/creating instance 被标记 lost/requeue；requeue 后不会保留 offline `node_id`。

- [ ] **Step 15: 扩展 instance repository**
  - Modify: `code/backend/internal/module/instance/infrastructure/repository.go`
  - 增加按 node 查询 active instances；扩展 `RequeueLostRuntime` 支持 reason=`runtime_node_offline` 且清 `node_id` 或写 replacement policy。

- [ ] **Step 16: 扩展 maintenance service**
  - Modify: `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - 增加 `RequeueRuntimesOnOfflineNodes` 或在 `ReconcileLostActiveRuntimes` 前处理 offline nodes。

- [ ] **Step 17: 确认 Jeopardy / AWD rebuild 语义**
  - Modify tests in `practice/application/commands/awd_desired_runtime_reconciler_test.go` 和 `instance_provisioning_test.go`
  - Jeopardy：只重建已有 active instance。
  - AWD：desired reconciler 对缺失 active scope 继续 create/recreate。

### Slice 5: startup recovery integration and docs

- [ ] **Step 18: 接入 startup recovery / cleaner order**
  - Modify: `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - Modify: `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - 确认 node offline marker 在 actual runtime reconcile 前后顺序合理；lock owner 不冲突。

- [ ] **Step 19: 更新事实源和运维文档**
  - Modify: `docs/architecture/backend/01-system-architecture.md`
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/operations/runtime-agent-deployment.md`
  - 明确 node health 状态、last_seen cutoff、调度排除、重建边界、会话中断边界。

- [ ] **Step 20: 运行最小验证**
  - Run: `cd code/backend && go test ./internal/module/container_runtime/... ./internal/app/composition ./internal/module/instance/application/commands ./internal/module/instance/infrastructure ./internal/module/practice/application/commands -run 'RuntimeNode|NodeHealth|ReconcileLost|DesiredAWD|Provisioning|StartupRuntimeRecovery|Cleaner' -count=1`

- [ ] **Step 21: Commit**
  - Run: `git add code/backend/migrations code/backend/internal/module/container_runtime code/backend/internal/app/runtime_node_migration_test.go code/backend/internal/app/composition code/backend/internal/module/instance code/backend/internal/module/practice docs/architecture/backend/01-system-architecture.md docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md && git commit -m "feat(backend): 增加 runtime node 健康与故障重建" -m "为 runtime_nodes 增加 last_seen 与健康调度 predicate，控制面通过 runtime-agent Health RPC 维护 node 可用性。" -m "node 失联后 active runtime 进入可重建链路，并避免重新调度到失效 node；AWD 与 Jeopardy 重建边界保持区分。" -m "Task: 2026-06-12-runtime-node-health-and-failover-rebuild"`

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/container_runtime/... -count=1`
  - `cd code/backend && go test ./internal/app/composition -run RuntimeNode -count=1`
  - `cd code/backend && go test ./internal/app -run RuntimeNodeMigration -count=1`
  - `cd code/backend && go test ./internal/module/instance/application/commands -run 'ReconcileLost|StartupRuntimeRecovery|RuntimeMaintenance' -count=1`
  - `cd code/backend && go test ./internal/module/instance/infrastructure -run Cleaner -count=1`
  - `cd code/backend && go test ./internal/module/practice/application/commands -run 'DesiredAWD|Provisioning|StartChallenge' -count=1`
  - `git diff --check -- code/backend/migrations code/backend/internal/module/container_runtime code/backend/internal/app/composition code/backend/internal/module/instance code/backend/internal/module/practice docs/architecture/backend/01-system-architecture.md docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md`
- Manual checks:
  - 停掉一个 runtime-agent node 后，health monitor 在 cutoff 后标记 offline。
  - 新实例不再调度到 offline node。
  - offline node 上已有 active Jeopardy instance 被 requeue，并在健康 node 上重建；原 SSH/WebSocket 会话中断后用户可重连。
  - AWD contest 的 team/service scope 在 node offline 后由 desired reconciler 补齐到健康 node。
- Review focus:
  - `health_status` / `last_seen_at` predicate 是否只有一个 owner，没有散落在多个 layer。
  - node offline requeue 是否清理或替换 `node_id`，避免 rebuild 回失效 node。
  - existing runtime cleanup 是否不会因 offline node 不可达而阻塞所有重建。
  - Jeopardy / AWD 重建边界是否清楚，没有承诺 live migration。
  - health monitor lock / keepalive 是否能防多 API 副本重复标记。
