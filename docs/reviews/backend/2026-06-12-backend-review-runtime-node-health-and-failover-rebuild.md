# Backend Review: Runtime Node Health And Failover Rebuild

## Findings

### Blocker 1: explicit node-bound operations are tied to `schedulable=true`

- Location:
  - `code/backend/internal/app/composition/runtime_node_execution_router.go:512`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go:515`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go:158`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go:160`
- Issue:
  - With `container.runtime_node_health.enabled=true`, every explicit node-bound runtime operation goes through `resolveNodeForExecution(nodeID)` and then `FindHealthyByID`.
  - `FindHealthyByID` requires `schedulable=true` in addition to `ready/degraded + fresh last_seen_at`.
  - This mixes two different concepts: "eligible for new scheduling" and "healthy enough to operate existing containers already bound to this node".
- Why it matters:
  - `schedulable=false` existed before this change as the selector-side switch for excluding a node from new placement.
  - After this diff, setting a healthy node to `schedulable=false` to cordon or drain new workload also makes old `node_id`-bound operations return `ErrRuntimeNodeUnavailable`.
  - The health evaluator scans only `ListSchedulableNodes` (`node_health_service.go:91`, `node_repository.go:238`-`246`), so an unschedulable but still alive node will stop receiving heartbeat updates, eventually look stale, and still will not trigger node-offline requeue because it is no longer in the health scan set.
  - That leaves existing instances in a bad middle state: no new scheduling lands there, but existing container access / cleanup / checker / file / SSH operations can fail without the offline failover path rebuilding them elsewhere.
- Architecture conflict:
  - The implementation plan says explicit old bindings should continue routing on healthy nodes and fail when the node is offline: `docs/plan/impl-plan/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md:57`.
  - Current docs describe `schedulable` as the eligibility condition for default new scheduling (`docs/architecture/backend/03-container-architecture.md:74`, `docs/operations/runtime-agent-deployment.md:150`), while explicit old operations are documented as failing for offline nodes (`docs/operations/runtime-agent-deployment.md:152`), not for merely unschedulable nodes.
- Required fix:
  - Split the lookup semantics:
    - default/new scheduling: require `schedulable=true`, healthy status, and fresh heartbeat;
    - explicit `node_id` runtime operations: require healthy status and fresh heartbeat, but do not require `schedulable=true`.
  - Keep heartbeats for unschedulable nodes if they can still host existing containers, for example by adding a health-scan repository method that lists runtime nodes independently of new-scheduling eligibility.
  - If the intended product meaning is instead that `schedulable=false` means "fully unavailable for all runtime operations", document that explicitly and make it trigger the same requeue/failover contract; the current implementation does neither.
- Required tests:
  - Add a router test where a node is `schedulable=false`, `health_status=ready`, and `last_seen_at` fresh:
    - default selection skips it;
    - explicit `node_id` operation still routes to it.
  - Add health-service or repository coverage proving unschedulable active nodes still get heartbeat evaluation, or proving the alternative "fully unavailable" contract requeues them.

## Classification Check

同意当前任务按 `结构性改动 / 非琐碎任务` 处理。该 diff 同时触达 schema、runtime node repository、execution router、background lifecycle、instance lifecycle repair、practice scheduler / desired AWD reconciler、config 与架构 / 运维事实源，需要独立 code-review gate。

## Gate Verdict

`blocked`

当前有 1 个 material blocker。修复前不建议进入 workflow completion。

## Implementation Response After Blocked Review

2026-06-13 实现上下文已按 blocker 修复 `schedulable` 与执行健康语义混淆：

- `RuntimeNodeRepository.FindHealthyByID` 不再要求 `schedulable=true`，显式 `node_id` 操作只要求 `ready/degraded + fresh last_seen_at`。
- 新调度路径仍使用 `ListSchedulableHealthyNodes` / `FindSchedulableHealthyNodeByName`，继续要求 `schedulable=true`。
- `RuntimeNodeRepository.ListHealthCheckNodes` 覆盖所有 runtime nodes；`NodeHealthService` 改用该接口，cordoned / unschedulable 节点仍会继续 heartbeat 或触发 offline failover。
- `runtimeNodeExecutionRouter.ListManagedContainers` 也改用 all-node inventory，避免 cordoned 节点上的已有容器从 maintenance / inventory cache 视图中消失。
- 新增回归测试覆盖 cordoned node 默认调度跳过、显式健康 nodeID 路由可用、offline nodeID 仍被拒绝、health service 探测 unschedulable 节点，以及 managed container inventory 仍扫描 cordoned 节点。

实现上下文已重跑 required re-validation 中的 Go 测试范围：

```bash
cd code/backend && timeout 90s go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1
cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1
cd code/backend && timeout 90s go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeNodeFailover|RuntimeModule' -count=1
cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition ./internal/config -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge|Defaults|Validate' -count=1
```

以上命令结果均为 PASS。该记录不是独立 re-review verdict；最终 gate verdict 仍需后续 reviewer 复核。

## Material Findings

- Blocker 1: explicit node-bound runtime operations incorrectly depend on `schedulable=true`, which can make healthy but cordoned nodes unusable without triggering offline failover.

## Senior Implementation Assessment

当前总体 owner 方向是合理的：`container_runtime` 维护 node health fact，`instance` 执行 node-scoped requeue，`practice` 继续消费 pending / desired AWD gap，composition 只负责 callback wiring。

主要问题是 repository 查询把调度 eligibility 直接复用于 explicit runtime execution。更低风险的形状是给 repository 暴露两个不同 contract：一个用于 new placement，一个用于 explicit node execution。这样不会把运维 cordon/drain 语义误伤到已有实例生命周期。

## Required Re-validation

修复 Blocker 1 后，至少重跑：

```bash
cd code/backend && timeout 90s go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1
cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1
cd code/backend && timeout 90s go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeNodeFailover|RuntimeModule' -count=1
cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition ./internal/config -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge|Defaults|Validate' -count=1
timeout 120s python3 scripts/check-docs-consistency.py
timeout 60s git diff --check -- code/backend docs/architecture/backend docs/operations docs/plan/impl-plan/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md docs/plan/impl-plan/2026-06-12-true-ha-group/INDEX.md docs/reviews/backend/2026-06-12-backend-review-runtime-node-health-and-failover-rebuild.md
```

`completion-full` 不需要由 reviewer 本轮重跑；实现上下文已有 PASS 证据。修复后如 touched surface 扩大，再由实现上下文按 code-workflow completion gate 判断是否重跑。

## Residual Risk

- 本轮 review 未执行真实多 runtime-agent 节点演练；多 node failover 仍依赖后续运维环境手工验证。
- `NodeHealthService` 的 offline handler 去重是单进程内状态。跨 API 进程或进程重启后可能重复调用 handler，但当前 node-scoped requeue 是条件更新，重复调用应只造成空 requeue / desired reconcile 再跑一次；这不是本轮 blocker。

## Touched Known-Debt Status

- `runtime_nodes.health_status` 缺少 heartbeat owner 的结构债已通过 `last_seen_at`、health-aware selector、NodeHealthService 和 docs 基本收口。
- `ReconcileLostActiveRuntimes` 依赖 per-container inspect、无法覆盖 node-level offline 的债务已通过 node-scoped requeue 方向收口。
- 本轮 blocker 是新的 contract 混淆：`schedulable` 被同时用作 new-scheduling eligibility 和 explicit runtime execution availability。该 debt 位于本次 touched surface 内，修复前不能作为 residual risk 留到后续。

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-runtime-node-health-and-failover-rebuild`
- Branch: `task/2026-06-12-runtime-node-health-and-failover-rebuild`
- Task slug: `2026-06-12-runtime-node-health-and-failover-rebuild`
- Plan: `docs/plan/impl-plan/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md`
- Diff source: current worktree uncommitted + untracked diff
- Reviewer mode: independent code-workflow gate review; production code not modified
- Review archive path: `docs/reviews/backend/2026-06-12-backend-review-runtime-node-health-and-failover-rebuild.md`

## Files Reviewed

- `code/backend/migrations/000019_add_runtime_node_last_seen_at.up.sql`
- `code/backend/migrations/000019_add_runtime_node_last_seen_at.down.sql`
- `code/backend/internal/module/container_runtime/entity/runtime_node.go`
- `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
- `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
- `code/backend/internal/module/container_runtime/application/node_health_service.go`
- `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
- `code/backend/internal/app/composition/container_runtime_module.go`
- `code/backend/internal/app/composition/runtime_node_execution_router.go`
- `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
- `code/backend/internal/app/composition/runtime_node_failover.go`
- `code/backend/internal/app/composition/runtime_node_failover_wiring_test.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
- `code/backend/internal/module/instance/infrastructure/repository.go`
- `code/backend/internal/module/instance/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/config/types.go`
- `code/backend/internal/config/defaults.go`
- `code/backend/internal/config/validate.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/configs/config.yaml`
- `code/backend/configs/config.prod.yaml`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/05-key-flows.md`
- `docs/operations/runtime-agent-deployment.md`
- `docs/plan/impl-plan/2026-06-12-true-ha-group/INDEX.md`
- `docs/plan/impl-plan/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md`

## Validation Evidence Reviewed

复用了实现上下文提供的证据：

- `cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealthService.*OfflineHandler' -count=1`: PASS
- `cd code/backend && timeout 90s go test ./internal/module/container_runtime/... -run 'NodeHealth|RuntimeNode' -count=1`: PASS
- `cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/application/commands ./internal/app/composition ./internal/config -run 'RuntimeNode|NodeHealth|Requeue|RuntimeMaintenance|Provisioning|DesiredAWD|StartChallenge|Defaults|Validate' -count=1`: PASS
- `timeout 120s python3 scripts/check-docs-consistency.py`: PASS
- `timeout 60s git diff --check -- code/backend docs/architecture/backend docs/operations docs/plan/impl-plan/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md docs/plan/impl-plan/2026-06-12-true-ha-group/INDEX.md`: PASS
- `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`: PASS

Reviewer did not rerun `completion-full`.

## Re-review 2026-06-13

### Re-review Scope / Date

- Date: `2026-06-13`
- Reviewer mode: independent backend code-workflow gate re-review after blocked finding fix; production code was not modified.
- Diff source: current worktree uncommitted + untracked diff.
- Focus:
  - Verify previous blocker fix for `schedulable=false` cordoned nodes.
  - Re-check runtime node repository selector semantics, health evaluator scan set, explicit node-bound runtime routing, managed container inventory, node offline requeue, scheduler reselection, and updated architecture / operations docs.
- Files reviewed in this pass:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/runtime_node_failover.go`
  - `code/backend/internal/app/composition/runtime_node_failover_wiring_test.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
  - this review archive

### Findings

#### Blockers

None.

The previous material blocker is fixed in the current diff:

- New/default scheduling still requires `schedulable=true + ready/degraded + fresh last_seen_at` through `ListSchedulableHealthyNodes` and `FindSchedulableHealthyNodeByName`.
- Explicit `node_id` execution uses `FindHealthyByID`, which checks `ready/degraded + fresh last_seen_at` and no longer checks `schedulable=true`.
- `NodeHealthService` uses `ListHealthCheckNodes`, so cordoned / unschedulable nodes stay in the health scan set.
- `runtimeNodeExecutionRouter.ListManagedContainers` also uses `ListHealthCheckNodes`, so existing containers on cordoned nodes remain visible to maintenance / inventory cache paths.
- Regression tests cover the fixed behavior: default routing skips cordoned nodes, explicit healthy node-bound routing still works, offline explicit routing is rejected, health service evaluates unschedulable nodes, and inventory scans cordoned nodes.

#### Suggestions

- Minor doc precision issue: `docs/architecture/backend/05-key-flows.md` says `WireRuntimeNodeFailover` is triggered only when a node enters `offline`. The implementation intentionally has a per-process successful-handler guard: a fresh API process can call the offline handler again for an already-offline node, and a failed handler attempt is retried on later failed probes. Because `RequeueLostRuntimesByNode` is conditional and returns only rows still matching `node_id + creating/running + unexpired`, this is idempotent and not material. The docs can be tightened later to say "after an offline mark / until successful handling in the current process" rather than strict cross-process transition-only semantics.

### Gate Verdict

`pass with minor issues`

### Material Findings

None.

### Required Re-validation

No additional blocker re-validation is required beyond the focused commands rerun in this re-review and the implementation context's existing broader evidence.

Commands independently rerun by reviewer:

```bash
cd code/backend && timeout 90s go test ./internal/module/container_runtime/infrastructure -run 'RuntimeNode' -count=1
cd code/backend && timeout 90s go test ./internal/module/container_runtime/application -run 'NodeHealth' -count=1
cd code/backend && timeout 90s go test ./internal/app/composition -run 'RuntimeNode.*(Offline|Healthy|Selector|Router)|RuntimeNodeFailover|RuntimeModule' -count=1
```

Result: all PASS.

### Residual Risk

- This re-review did not run a real two-node runtime-agent outage rehearsal. Multi-node behavior still depends on the operations manual validation path.
- `NodeHealthService` offline handler success tracking is process-local. Repeated handling after API restart is acceptable because instance requeue is conditionally scoped to current rows still bound to the offline node.
- The default bootstrap path still owns the configured default runtime node record on startup. That is outside this blocker fix, but operators should avoid using the bootstrap default node as a long-lived manual cordon control without confirming startup behavior.

### Touched Known-Debt Status

- The touched known debt around `schedulable` vs health semantics is closed for this task: cordon now means "no new scheduling" and no longer disables healthy explicit old-container operations.
- The touched known debt around health evaluator coverage is closed: all runtime nodes are scanned, not only schedulable nodes.
- The touched known debt around maintenance / inventory disappearance for cordoned nodes is closed: all runtime nodes are included in managed container inventory.

## Post-review Follow-up 2026-06-13

- The minor documentation precision suggestion was applied in `docs/architecture/backend/05-key-flows.md`: node offline failover is now described as process-local success dedupe with retry after failed handling, and as idempotent across API restart because node-scoped requeue conditionally updates only rows still bound to the offline node.
- The implementation plan handoff summary was updated with the same wording before archival.
