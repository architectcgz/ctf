<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# Runtime Node Health And Failover Rebuild Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Backend behavior changes must use `test-driven-development`: write the failing test, verify RED, implement minimal GREEN, then refactor while keeping tests green.

**Goal:** 把 `runtime_nodes` 从“调度元数据”提升为可用性 owner：节点心跳过期后自动摘除新调度，并把该节点上的可恢复实例重新入队，由现有 pending scheduler / AWD desired reconciler 在健康节点上重建。

**Architecture:** 继续沿用 `runtime-agent + node_id + PostgreSQL + Redis`，不引入 live migration、Kubernetes 或新的调度服务。`container_runtime` owner 维护 runtime node health / heartbeat / capacity snapshot；`instance` owner 继续负责 active runtime recovery；`practice` pending scheduler 与 AWD desired reconciler 复用现有重建链路。

**Tech Stack:** Go, GORM, PostgreSQL migration, runtime-agent bridge, container runtime module, instance maintenance service, practice scheduler, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-runtime-node-health-and-failover-rebuild`
- Parent Task Group: `2026-06-12-true-ha-group`
- Slice Index: `5/5`
- Depends On: `2026-06-12-shared-storage-owner-convergence, 2026-06-12-distributed-event-bus-and-outbox-relay, 2026-06-12-ssh-gateway-ha-and-draining`
- Started At: `2026-06-12T15:48:05Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-runtime-node-health-and-failover-rebuild`
- Branch: `task/2026-06-12-runtime-node-health-and-failover-rebuild`
- Plan Type: `slice`

## Plan Status

- Status: `ready-to-merge`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 为 `runtime_nodes` 补 `last_seen_at` 持久化事实，并提供 repository 方法维护 `health_status`、`last_seen_at`、`capacity_snapshot`、`schedulable`。
  - 让默认 runtime node selector 和 execution router 只把新实例 / 默认执行路由导向 `schedulable + ready/degraded + heartbeat fresh` 的节点。
  - 新增 runtime node health evaluator 后台任务，周期性探测每个 runtime node，成功时写 ready heartbeat 与 capacity snapshot，失败或心跳过期时标记 offline；`schedulable=false` 只摘除新调度，不阻断健康节点上的显式旧容器操作。
  - 节点转为 offline 时，重排该节点上 `creating/running` 且未过期的实例到 `pending`，清空旧 runtime identity，复用现有 practice scheduler 重建；AWD service 缺口继续由 desired reconciler 补齐。
  - 同步架构与运维事实源，明确节点故障只承诺重建，不承诺透明会话迁移。
- Non-Goals:
  - 不实现容器 live migration、原 TCP/SSH/WebSocket 会话保持或跨 node 原地迁移。
  - 不引入 Kubernetes / Swarm / 新调度器，也不实现复杂负载均衡或资源 bin-pack。
  - 不把所有 runtime-agent 运维 API 扩成完整 node telemetry 协议；本轮 capacity snapshot 先基于已有 `ListManagedContainerStats` 汇总。
  - 不改变外部 HTTP API / WebSocket 协议；本轮是后端 runtime 可用性行为变化。

## Problem Statement

- Current behavior / structure:
  - `runtime_nodes` 已有 `health_status` 与 `capacity_snapshot`，但没有 `last_seen_at`，无法区分刚注册、健康、过期和离线。
  - `RuntimeNodeRepository.FindFirstSchedulableNode` / `FindSchedulableNodeByName` 只检查 `schedulable = true`，会把新实例继续调度到已失联节点。
  - `runtimeNodeExecutionRouter` 对显式 `node_id` 只按 ID 找节点，不检查健康状态；对默认路由仍回落到 first schedulable。
  - `InstanceMaintenanceService.ReconcileLostActiveRuntimes` 已能把缺失 runtime 的实例 `RequeueLostRuntime` 回 `pending`，但当前依赖容器 inspect，节点整体失联时 inspect error 会跳过实例。
  - `startup_runtime_recovery` 只在宿主 `boot_id` 变化时触发整机 outage recovery，不覆盖运行期间单个 runtime-agent node 失联。
- Target behavior / structure:
  - runtime node health 由 `container_runtime` owner 维护，心跳成功写 `ready`，探测失败或超出 stale threshold 写 `offline` 并摘除调度。
  - 新实例调度只选择健康节点；显式绑定的旧实例仍可在健康节点上按 `node_id` 路由，节点离线则返回 `ErrRuntimeNodeUnavailable`。
  - offline transition 调用 instance owner 的 node-scoped requeue 能力，把该 node 上可恢复 active instances 变回 pending；已有 scheduler 在健康 node 上重建。
  - AWD desired runtime reconciliation 仍是 team x service 期望态 owner；node failover 只负责让旧 active runtime 不再阻塞 desired gap。
- Why this task is needed now:
  - T1/T2/T3/T4 已经把状态面、共享文件/密钥、SSH gateway、跨副本事件收口到 HA 基线；剩余 blocker 是执行面 node 故障后仍会继续调度到失效 node，且旧 runtime 不会自动重建。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
- Related architecture/contracts:
  - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- Related prior work:
  - T1 `redis-sentinel-and-postgres-ha-connectivity`
  - T2 `shared-storage-owner-convergence`
  - T3 `ssh-gateway-ha-and-draining`
  - T4 `distributed-event-bus-and-outbox-relay`

## Task Classification

- Classification: `结构性改动 / 非琐碎任务`
- Why:
  - 触达 schema、runtime node repository、execution router、background lifecycle、instance recovery、practice scheduler 复用边界和架构/运维事实源。
  - 改变运行时调度与故障恢复行为，需要测试先锁定 owner 与回归边界。
  - 这是 HA task group 的最后执行面 slice，必须完成独立 review gate。

## Files

- Create:
  - `code/backend/migrations/000019_add_runtime_node_last_seen_at.up.sql`
  - `code/backend/migrations/000019_add_runtime_node_last_seen_at.down.sql`
  - `code/backend/internal/module/container_runtime/application/node_health_service.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
  - `docs/reviews/backend/2026-06-12-backend-review-runtime-node-health-and-failover-rebuild.md`（独立 review 通过后归档）
- Modify:
  - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - `code/backend/internal/module/container_runtime/contracts/runtime_node.go`
  - `code/backend/internal/module/container_runtime/ports/node.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/runtime/module.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/config/{types.go,defaults.go,validate.go}`
  - `code/backend/configs/{config.yaml,config.prod.yaml}`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
- Review:
  - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/app/composition/runtime_module_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
- Test:
  - `code/backend/internal/module/container_runtime/application/*_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/*_test.go`
  - `code/backend/internal/module/instance/infrastructure/*_test.go`
  - `code/backend/internal/module/instance/application/commands/*_test.go`
  - `code/backend/internal/app/composition/*runtime*_test.go`
  - `code/backend/internal/config/config_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `runtime_nodes` schema already has `health_status` / `capacity_snapshot` but no `last_seen_at`.
  - `RuntimeNodeRepository` owns default bootstrap and schedulable lookup.
  - `runtimeNodeExecutionRouter` owns node-bound execution routing and client cache.
  - `InstanceMaintenanceService.ReconcileLostActiveRuntimes` already requeues lost runtimes through `RequeueLostRuntime`.
  - `practice_instance_scheduler` already consumes pending instances and persists selected `node_id`.
  - `ReconcileDesiredAWDInstances` already fills missing AWD `team x service` desired runtime.
- Reuse / extend / split / create-new decision:
  - Extend `container_runtime` with a focused node health application service and repository methods; do not put node health into `instance` or `practice`.
  - Extend existing instance repository with node-scoped recoverable query / requeue capability; do not duplicate instance state transitions in `container_runtime`.
  - Reuse `ListManagedContainerStats` for capacity snapshot; do not add a new runtime-agent health RPC unless tests show current RPC cannot support the behavior.
  - Reuse existing background job registration and Redis lock patterns where needed; do not invent a parallel scheduler framework.
- Owner boundary:
  - `container_runtime`: runtime node health facts, heartbeat/capacity snapshot, schedulable health selector, node health evaluator lifecycle.
  - `app/composition`: wiring concrete node clients and passing node-scoped requeue callback into the health evaluator.
  - `instance`: deciding which instances on an offline node are requeued and clearing runtime identity safely.
  - `practice`: consuming pending instances and desired AWD gaps; no new failover state machine.
- Why this is the narrowest safe surface:
  - It changes the single missing availability owner without redesigning scheduler, runtime-agent protocol, or instance lifecycle.
  - It uses the already-tested `pending -> creating -> running` path for rebuild instead of adding a second rebuild path.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - T5 adds runtime availability behavior across schema, routing and background jobs. The decision tree is about owner boundaries and failure semantics, not just adding fields.
- Evidence inspected:
  - HA group index and umbrella plan show T5 depends on T2/T3/T4 and should not start before outbox/fanout is landed.
  - Current branch HEAD includes `merge: 合并 T4 跨副本事件 outbox relay`.
  - `runtime_nodes` base table has `health_status` / `capacity_snapshot` but lacks `last_seen_at`.
  - selector methods only check `schedulable = true`.
  - startup recovery covers host boot ID change, not runtime node heartbeat expiry.
  - instance maintenance already requeues missing runtime identity and restarts stopped containers where possible.
- grill-with-docs findings:
  - Architecture docs currently say formal multi-node mode still primarily uses explicit binding or single-node default; this task must update that current fact after implementation.
  - `ReconcileLostActiveRuntimes` intentionally skips inspect errors, which is correct for transient Docker/API errors but insufficient for node-level offline state; node offline must bypass per-container inspect and requeue by node owner.
  - Desired AWD reconciliation should remain the expected-state owner. Node failover should remove stale active blockers, then let desired reconciliation fill the missing service.
  - `health_status = degraded` needs a clear schedulable meaning. This plan treats `ready` and `degraded` as selectable if heartbeat is fresh; `unknown/offline` are not selectable.
- Plan adjustments after challenge:
  - Add `last_seen_at` and explicit stale threshold instead of overloading `updated_at`.
  - Add node-scoped instance requeue rather than changing global `ListRecoverableActiveInstances` semantics.
  - Keep capacity snapshot as observability and future scheduling input; do not make load balancing depend on it in this slice.

## Execution Slices

### Slice 1: runtime node health persistence and selector semantics

- Goal:
  - Add durable `last_seen_at`, define selectable health semantics, and make repository lookups exclude offline / stale nodes.
- Dependencies:
  - T1 state HA baseline.
- Files:
  - Create:
    - `code/backend/migrations/000019_add_runtime_node_last_seen_at.up.sql`
    - `code/backend/migrations/000019_add_runtime_node_last_seen_at.down.sql`
    - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - Modify:
    - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
    - `code/backend/internal/module/container_runtime/contracts/runtime_node.go`
    - `code/backend/internal/module/container_runtime/ports/node.go`
    - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - Review:
    - `code/backend/internal/app/runtime_node_migration_test.go`
  - Test:
    - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
- Steps:
  - [x] Step 1: Write failing repository tests for `last_seen_at`, selectable `ready/degraded`, stale heartbeat exclusion, offline exclusion, and explicit default-node name fallback.
  - [x] Step 2: Run `cd code/backend && go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1` and verify RED.
  - [x] Step 3: Add migration/entity field and repository methods: `MarkNodeHeartbeat`, `MarkNodeOffline`, `ListSchedulableHealthyNodes`, `FindSchedulableHealthyNodeByName`, `FindHealthyByID`.
  - [x] Step 4: Update `NewDefaultRuntimeNodeSelector` to accept a stale threshold option and use health-aware lookups.
  - [x] Step 5: Re-run the same test command and verify GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1`
- Review focus:
  - `updated_at` is not used as heartbeat.
  - `unknown/offline` are never selected for new scheduling.
  - `degraded` remains selectable only while `last_seen_at` is fresh.
- Done criteria:
  - Repository owns health-aware selection and heartbeat state, with migration coverage.

### Slice 2: node health evaluator application service and background lifecycle

- Goal:
  - Add a `container_runtime/application` service that probes nodes, stores heartbeat/capacity snapshot, marks offline nodes, and exposes a loop suitable for app background jobs.
- Dependencies:
  - Slice 1.
- Files:
  - Create:
    - `code/backend/internal/module/container_runtime/application/node_health_service.go`
    - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
  - Modify:
    - `code/backend/internal/module/container_runtime/runtime/module.go`
    - `code/backend/internal/app/composition/container_runtime_module.go`
    - `code/backend/internal/config/{types.go,defaults.go,validate.go}`
    - `code/backend/configs/{config.yaml,config.prod.yaml}`
  - Review:
    - `code/backend/internal/module/container_runtime/application/container_stats_service.go`
    - `code/backend/internal/app/composition/root.go`
  - Test:
    - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
    - `code/backend/internal/app/composition/runtime_module_test.go`
    - `code/backend/internal/config/config_test.go`
- Steps:
  - [x] Step 1: Write failing application tests for successful heartbeat, capacity snapshot aggregation, probe failure marking node offline, stale threshold marking offline, and loop cancellation.
  - [x] Step 2: Run `cd code/backend && go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1` and verify RED.
  - [x] Step 3: Implement `NodeHealthService` with explicit `ctx`, `now` injection for tests, and capacity snapshot JSON generated from managed container stats.
  - [x] Step 4: Add config fields under `container.runtime_node_health`: `enabled`, `poll_interval`, `probe_timeout`, `stale_after`, `failure_threshold`.
  - [x] Step 5: Register the node health background job from `BuildContainerRuntimeModule` when enabled.
  - [x] Step 6: Re-run node health, runtime module, and config tests and verify GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'RuntimeModule|BackgroundJob' -count=1`
  - `cd code/backend && go test ./internal/config -run 'RuntimeNodeHealth|Defaults|Validate' -count=1`
- Review focus:
  - Background job uses caller lifecycle context and stops cleanly.
  - Probe timeout/cancel functions are always released.
  - Capacity snapshot is observability-only in this slice.
- Done criteria:
  - Healthy nodes keep heartbeating; failed/stale nodes become offline and unschedulable by repository semantics.

### Slice 3: execution router and default scheduler health filtering

- Goal:
  - Make new scheduling and default execution routing respect health-aware repository lookups.
- Dependencies:
  - Slice 1.
- Files:
  - Modify:
    - `code/backend/internal/app/composition/runtime_node_execution_router.go`
    - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
    - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
    - `code/backend/internal/app/composition/runtime_module_test.go`
  - Review:
    - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
    - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - Test:
    - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
    - `code/backend/internal/app/composition/runtime_module_test.go`
- Steps:
  - [x] Step 1: Write failing tests showing `SelectDefaultNode` skips offline/stale nodes and `runtimeNodeExecutionRouter` returns `ErrRuntimeNodeUnavailable` for explicit offline node IDs.
  - [x] Step 2: Run `cd code/backend && go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)' -count=1` and verify RED.
  - [x] Step 3: Update router node resolution to call health-aware repository methods for default and explicit runtime execution.
  - [x] Step 4: Ensure existing explicit healthy node routing tests still pass.
  - [x] Step 5: Re-run the same test command and verify GREEN.
- Validation:
  - `cd code/backend && go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)' -count=1`
- Review focus:
  - Explicit old node bindings should not silently fall back to a different node for direct container operations.
  - New instance selection may choose another healthy node, but old container operations must fail when the bound node is offline.
- Done criteria:
  - New scheduling excludes bad nodes, and old node-bound operations fail loudly instead of drifting to another node.

### Slice 4: node-scoped failover requeue and desired AWD rebuild trigger

- Goal:
  - When a node transitions offline, requeue that node's active instances to pending and invoke desired AWD reconciliation after stale active rows stop blocking expected state.
- Dependencies:
  - Slice 1
  - Slice 2
  - Slice 3
- Files:
  - Modify:
    - `code/backend/internal/module/instance/ports/ports.go`
    - `code/backend/internal/module/instance/infrastructure/repository.go`
    - `code/backend/internal/module/instance/infrastructure/repository_test.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
    - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
    - `code/backend/internal/app/composition/instance_module.go`
    - `code/backend/internal/app/composition/container_runtime_module.go`
  - Review:
    - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
    - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - Test:
    - `code/backend/internal/module/instance/infrastructure/repository_test.go`
    - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
- Steps:
  - [x] Step 1: Write failing repository test for `RequeueLostRuntimesByNode(ctx, nodeID)` covering `creating/running` + unexpired rows only, clearing runtime identity and keeping stopped/expired rows untouched.
  - [x] Step 2: Write failing maintenance/application test showing node offline callback requeues only that node and records AWD system recreate operations.
  - [x] Step 3: Run `cd code/backend && go test ./internal/module/instance/infrastructure ./internal/module/instance/application/commands -run 'Node|Requeue|RuntimeMaintenance' -count=1` and verify RED.
  - [x] Step 4: Implement node-scoped repository and service method, then wire the node health evaluator to call it after marking a node offline.
  - [x] Step 5: Wire desired AWD reconciler invocation after node offline requeue where app composition already has the reconciler; avoid cycles by passing a small callback.
  - [x] Step 6: Re-run the same tests and verify GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/instance/infrastructure ./internal/module/instance/application/commands -run 'Node|Requeue|RuntimeMaintenance' -count=1`
- Review focus:
  - Node-scoped requeue must not clear `node_id` before the scheduler selects a replacement unless the existing start path would otherwise reuse the old node. If replacement selection requires clearing `node_id`, document and test it explicitly.
  - Offline node requeue must not touch `stopping`, `stopped`, `expired`, or already pending rows.
  - Desired AWD reconciler remains expected-state owner; no duplicated `team x service` scan in container_runtime.
- Done criteria:
  - Node offline transition results in pending rebuild candidates and desired AWD gaps can be filled by existing owner.

### Slice 5: docs, group status, validation, and review handoff

- Goal:
  - Update current fact sources and complete workflow validation before independent review.
- Dependencies:
  - Slice 1-4.
- Files:
  - Modify:
    - `docs/architecture/backend/01-system-architecture.md`
    - `docs/architecture/backend/03-container-architecture.md`
    - `docs/architecture/backend/05-key-flows.md`
    - `docs/operations/runtime-agent-deployment.md`
    - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
    - This plan file's `Validation Evidence` / `Independent Review Handoff`
  - Create:
    - `docs/reviews/backend/2026-06-12-backend-review-runtime-node-health-and-failover-rebuild.md`
  - Review:
    - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - Test:
    - Architecture/doc/workflow checks.
- Steps:
  - [x] Step 1: Update architecture docs to state current node health, selector, failover, and no-live-migration boundaries.
  - [x] Step 2: Update runtime-agent operations doc with heartbeat/stale/offline behavior and manual validation steps.
  - [x] Step 3: Update HA task group index status from T4/T5 pending to match current implementation state.
  - [x] Step 4: Run focused Go tests, architecture/docs checks, and `git diff --check`; record evidence below.
  - [x] Step 5: Prepare independent review handoff, run the review gate when tooling allows, fix material findings, and re-run impacted verification.
- Validation:
  - See `Validation Plan`.
- Review focus:
  - Docs must describe implemented behavior only.
  - Manual validation must be explicit about accepted session interruption and rebuild expectation.
- Done criteria:
  - Code, docs, task group status, validation evidence, and review evidence are all aligned.

## Impact And Compatibility

- API / DTO:
  - No public HTTP / WebSocket contract change.
- Data / migration:
  - Adds `runtime_nodes.last_seen_at timestamp with time zone NULL`.
  - Existing nodes start as `unknown` with `last_seen_at = NULL`; after first health probe they become `ready` or `offline`.
- State / cache / queue / event:
  - Existing pending scheduler queue is reused through `instances.status = pending`.
  - No new Redis event stream is required by T5.
- Runtime / config:
  - Adds `container.runtime_node_health.*` configuration.
  - Runtime node capacity snapshot is stored in `runtime_nodes.capacity_snapshot`.
- Frontend route / state / UX:
  - No frontend route changes. Users may see interrupted sessions on node loss; subsequent rebuild/reconnect should work.
- Docs / contracts:
  - Backend architecture and runtime-agent operations docs must be updated.

## Plan Review / Architecture Fit

- Target owner boundary:
  - Runtime node health belongs to `container_runtime`; instance lifecycle repair belongs to `instance`; desired AWD fill belongs to `practice`; composition only wires callbacks.
- Reuse points / landing zones:
  - `RuntimeNodeRepository` for durable node facts.
  - `InstanceMaintenanceService` and `Repository.RequeueLostRuntime` pattern for lost runtime repair.
  - `practice_instance_scheduler` for actual rebuild.
- Known structural debt touched:
  - `runtime_nodes.health_status` existed without heartbeat semantics. This task closes that debt by adding `last_seen_at`, stale rules, selector tests and docs.
  - `ReconcileLostActiveRuntimes` previously skipped inspect errors. This task keeps that local-container behavior but adds node-scoped offline requeue so node outage does not depend on per-container inspect.
- How this plan avoids behavior-only convergence:
  - It does not just add a periodic ping. It aligns schema, selector, router, requeue owner, scheduler reuse and docs in one slice.
- Hidden second-redesign risk:
  - Full capacity-aware scheduling is intentionally out of scope; capacity snapshot is stored now so future scheduler improvements can use a stable field without reworking heartbeat owner.
- Decision after review:
  - Proceed with the above five slices. Do not broaden into bin-pack scheduling or live migration.

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
- Fact sources to update after implementation:
  - Same four files, plus `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md` for task group progress.
- Plan-only notes that must not become architecture source:
  - Test names, step ordering, red/green commands, review handoff details.
- Archive condition:
  - After completion-full, independent review, workflow governance, and task handoff, use `harness/workflow-plugins/code-workflow/archive_task_artifacts.sh` to archive this plan.

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1`
  - `cd code/backend && go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeModule' -count=1`
  - `cd code/backend && go test ./internal/module/instance/infrastructure ./internal/module/instance/application/commands -run 'Node|Requeue|RuntimeMaintenance' -count=1`
  - `cd code/backend && go test ./internal/config -run 'RuntimeNodeHealth|Defaults|Validate' -count=1`
- Integration commands:
  - `cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD' -count=1`
  - `bash scripts/check-architecture.sh --full`
  - `python3 scripts/check-docs-consistency.py`
  - `git diff --check -- code/backend docs/architecture/backend docs/operations docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
- Manual checks:
  - Two runtime nodes registered; stop node B's runtime-agent; after stale threshold, node B is `offline` and no new instance selects it.
  - Existing `running` instance on node B becomes `pending`, then practice scheduler rebuilds it on node A.
  - AWD `team x service` runtime on node B no longer blocks desired reconciliation and is recreated on a healthy node.
  - SSH/WebSocket sessions on failed node are allowed to drop; reconnect uses rebuilt runtime.
- Commands intentionally skipped and why:
  - Full multi-host failover rehearsal depends on external runtime-agent deployment and will remain a manual operations check unless the environment is available during this task.

## Validation Evidence

- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1`
  - Result: PASS
  - Notes: Runtime node migration-adjacent repository behavior, heartbeat persistence and healthy selector semantics.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1`
  - Result: PASS
  - Notes: Node health evaluator heartbeat, capacity snapshot, offline transition and cancellation behavior.
- Command: `cd code/backend && timeout 90s go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeModule|RuntimeNodeFailover' -count=1`
  - Result: PASS
  - Notes: Health-aware router / selector, background job registration and failover callback wiring.
- Command: `cd code/backend && timeout 90s go test ./internal/module/instance/infrastructure ./internal/module/instance/application/commands -run 'Node|Requeue|RuntimeMaintenance' -count=1`
  - Result: PASS
  - Notes: Node-scoped requeue, runtime identity clearing and instance maintenance callback behavior.
- Command: `cd code/backend && timeout 90s go test ./internal/config -run 'RuntimeNodeHealth|Defaults|Validate' -count=1`
  - Result: PASS
  - Notes: Runtime node health defaults and validation.
- Command: `cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD' -count=1`
  - Result: PASS
  - Notes: Cross-module runtime node health / failover / scheduler integration scope.
- Command: `cd code/backend && timeout 120s go test ./internal/module/practice/application/commands ./internal/module/practice/ports ./internal/app/composition -run 'ProcessPendingInstance(PersistsReselectedRuntimeNodeBeforeCreate|ReselectsRuntimeNodeBeforeProvisioning|ReturnsToPendingWhenNoRuntimeNodeAvailable)|RuntimeNode|Practice|Context' -count=1`
  - Result: PASS
  - Notes: Review-driven scheduler fixes: pending rows reselect healthy node before create, persist replacement `node_id` before runtime create, and return to `pending` instead of `failed` when no healthy node exists.
- Command: `cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition ./internal/config -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge|Defaults|Validate' -count=1`
  - Result: PASS
  - Notes: Post-review regression scope after requeue idempotency, offline callback idempotency and scheduler node-binding fixes.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1`
  - Result: PASS
  - Notes: Review-driven cancellation fix: parent context cancellation during shutdown no longer counts as a runtime node probe failure or triggers offline failover.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/... -run 'NodeHealth|RuntimeNode' -count=1`
  - Result: PASS
  - Notes: Runtime node and node health regression scope after cancellation semantics fix.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealthService.*OfflineHandler' -count=1`
  - Result: PASS
  - Notes: Review-driven offline handler retry fix: successful offline handling is not repeated in the same health service process, but a transient handler failure is retried on the next failed probe instead of being lost after the DB row becomes `offline`.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/... -run 'NodeHealth|RuntimeNode' -count=1`
  - Result: PASS
  - Notes: Runtime node and node health regression scope after offline handler retry semantics fix.
- Command: `cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition ./internal/config -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge|Defaults|Validate' -count=1`
  - Result: PASS
  - Notes: Cross-module runtime node health / failover / scheduler / config regression scope after offline handler retry fix.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1`
  - Result: PASS
  - Notes: Review blocker fix: new scheduling still filters `schedulable=true`, while explicit healthy lookup accepts cordoned but fresh nodes and health scan lists unschedulable nodes.
- Command: `cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1`
  - Result: PASS
  - Notes: Review blocker fix: node health evaluator probes unschedulable nodes instead of using the new-scheduling node list.
- Command: `cd code/backend && timeout 90s go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeNodeFailover|RuntimeModule' -count=1`
  - Result: PASS
  - Notes: Review blocker fix: default routing skips cordoned nodes, explicit healthy `node_id` operations still route to their original node, and managed container inventory still scans cordoned nodes so existing containers do not disappear from maintenance/cache views.
- Command: `cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition ./internal/config -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge|Defaults|Validate' -count=1`
  - Result: PASS
  - Notes: Cross-module runtime node health / failover / scheduler / config regression scope after schedulable-vs-health semantics split.
- Command: `timeout 180s bash scripts/check-architecture.sh --full`
  - Result: PASS
  - Notes: Backend module/app/test architecture and frontend architecture guards.
- Command: `timeout 120s python3 scripts/check-docs-consistency.py`
  - Result: PASS
  - Notes: Documentation references, architecture status and diagram source checks.
- Command: `timeout 60s git diff --check -- code/backend docs/architecture/backend docs/operations docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
  - Result: PASS
  - Notes: No whitespace errors in touched code/docs diff.
- Command: `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Implementation-context self-check after review-driven offline handler retry fix.
- Command: `timeout 30s bash -lc "! rg -n '\\[ \\]' docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md"`
  - Result: PASS
  - Notes: Active task plan checklist has no remaining unchecked items after independent review gate and post-review doc follow-up.
- Command: `timeout 120s python3 scripts/check-docs-consistency.py`
  - Result: PASS
  - Notes: Post-review documentation precision follow-up kept architecture docs and references consistent.
- Command: `timeout 60s git diff --check -- code/backend docs/architecture/backend docs/operations docs/reviews docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
  - Result: PASS
  - Notes: No whitespace errors after review archive and doc precision follow-up.
- Command: `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Final implementation-context self-check after independent re-review and post-review documentation follow-up.
- Command: `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
  - Result: PASS
  - Notes: Post-review workflow governance checks, documentation consistency, OpenAPI sync check, script layer guards and startup workflow scaffold checks.

## Independent Review Handoff

- Review target:
  - Task slug `2026-06-12-runtime-node-health-and-failover-rebuild`
  - Plan `docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md`
  - Diff basis: `main...HEAD`
- Validation evidence summary:
  - Focused runtime node repository, NodeHealthService, app composition, instance requeue/maintenance and config tests passed.
  - Cross-module integration-ish Go tests passed for `RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge`.
  - `check-architecture.sh --full`, `check-docs-consistency.py` and `git diff --check` passed.
- Review-driven fixes already folded into current diff:
  - Node offline failover callback is retried after failed handling and not repeated after successful handling in the same health service process; after API restart it may run again for an already-`offline` node, with idempotency provided by conditional node-scoped requeue.
  - Node-scoped requeue conditionally updates only rows that are still `creating / running` for the offline node, and returns only rows actually requeued.
  - Scheduler reselects a healthy runtime node after claiming `pending -> creating`, persists the replacement `node_id` before runtime create, and moves the row back to `pending` when no healthy node exists.
  - Node health probe errors caused by parent context cancellation are not counted as node failures, so API shutdown does not mark nodes offline or trigger failover.
  - Node offline handler idempotency now records success, not only DB `offline` state: successful handler execution is not repeated by the same health service process, while transient handler failures are retried on later failed probes.
  - `schedulable` semantics are split from explicit execution health: default/new scheduling requires `schedulable=true + ready/degraded + fresh heartbeat`; explicit `node_id` operations require healthy/fresh node but do not require `schedulable=true`; health probing and managed container inventory scan all runtime nodes so cordoned nodes keep heartbeating, existing containers remain visible to maintenance, and offline failover can still trigger.
- Architecture / contract inputs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
- Known risks / review focus:
  - Node health owner must not duplicate instance lifecycle owner.
  - Offline explicit node routing must fail rather than silently move old container operations to another node.
  - Requeue must not touch stopped/stopping/expired rows.
  - Config defaults and validation must prevent tight loops and false offline churn.
- Project-local checks to consider:
  - `bash scripts/check-architecture.sh --full`
  - `python3 scripts/check-docs-consistency.py`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## Rollback / Recovery

- Safe revert boundary:
  - Code and docs can be reverted as one task slice; DB migration down removes `runtime_nodes.last_seen_at`.
- Data / config / runtime recovery notes:
  - If health evaluator causes false offline in production, set `container.runtime_node_health.enabled: false` and restart API to return to previous schedulable behavior while preserving schema.
  - Offline nodes can be restored by fixing the agent and letting the next successful health probe mark heartbeat ready.
- Irreversible operations:
  - None expected. Requeueing active instances is a runtime state change, but it uses existing pending rebuild semantics and clears old runtime identity by design.

## Residual Risks

- Risk:
  - Capacity snapshot is not yet used for load-aware scheduling.
- Why acceptable:
  - This T5 goal is availability/failover correctness. Capacity-aware placement is a separate scheduler optimization and should build on the heartbeat owner after this lands.
- Follow-up owner, if any:
  - Future scheduler optimization under `container_runtime` / `practice` scheduling boundary.
