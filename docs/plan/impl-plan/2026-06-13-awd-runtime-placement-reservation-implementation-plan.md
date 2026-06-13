<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# AWD Runtime Placement And Reservation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 AWD 比赛建立显式 runtime placement / reservation 契约，保证 `single_node + docker_bridge_alias` 比赛在开赛前先占用宿主容量，赛中不会因 node failover 或默认 selector 自动漂移到其他 runtime node。

**Architecture:** `contest` 拥有比赛级 placement / reservation 业务事实，`container_runtime` 拥有 node health、capacity snapshot 和底层 reservation 账本，`practice` 在 AWD 实例创建与 desired reconcile 中消费固定 placement。管理员 UI 在赛前 preflight 中展示 node 容量和健康状态，并提供显式 reserve / rebind / emergency recreate 入口。

**Tech Stack:** Go modular monolith, GORM/PostgreSQL migrations, Gin admin API, runtime-agent node routing, Vue 3 + Pinia/composables, OpenAPI v1 contracts, Vitest + Go tests.

---

## Task Metadata

- Task Slug: `2026-06-13-awd-runtime-placement-reservation`
- Parent Task Group: `无`
- Slice Index: `-`
- Depends On: `无`
- Started At: `2026-06-13T01:35:59Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-awd-runtime-placement-reservation`
- Branch: `task/2026-06-13-awd-runtime-placement-reservation`
- Plan Type: `slice`

## Plan Status

- Status: `ready-for-implementation`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - Persist AWD `network_mode=docker_bridge_alias` and `placement_mode=single_node` as explicit contest runtime contract.
  - Add preflight reservation so capacity is reserved before start / prewarm / checker / round operations.
  - Split node eligibility into execution, provisioning and placement semantics.
  - Make all AWD runtime creation paths obey contest placement.
  - Show node health and capacity in the first UI version.
- Non-Goals:
  - No distributed AWD overlay/proxy network in this slice.
  - No live migration of running containers.
  - No partial rebind for `single_node + docker_bridge_alias`.
  - No CPU/memory scheduling score; CPU/memory stay in `capacity_snapshot`.
  - No automatic start-time reservation hidden behind `GET /readiness`.

## Problem Statement

- Current behavior / structure:
  - `practiceRuntimeNodeSelectorAdapter.SelectRuntimeNode` ignores `InstanceScope` and delegates to the default runtime selector.
  - `RuntimeNodeRepository.ListSchedulableHealthyNodes` treats `ready/degraded + fresh + schedulable` as default-new-scheduling eligible.
  - AWD Docker bridge alias is created in `runtime_container_create.go` / `awd_runtime_rules.go`, but there is no persisted contest-level placement contract.
  - `WireRuntimeNodeFailover` calls `InstanceModule.HandleRuntimeNodeOffline` and then `ReconcileDesiredAWDInstances`; current docs say replacement scheduling may pick a healthy node.
  - `GET /admin/contests/:id/awd/readiness` is read-only challenge/checker readiness and does not reserve runtime capacity.
- Target behavior / structure:
  - `docker_bridge_alias` requires `single_node`; the whole contest binds to one runtime node.
  - `preflight reserve` creates or refreshes active placement/reservation; start/prewarm/round/checker paths require it.
  - Single-node AWD desired reconcile reads the placement node and leaves scopes pending when that node is unavailable.
  - Rebind is an explicit admin action; ordinary rebind is only pre-start/no-active-runtime, while active runtime movement is `emergency rebind/recreate all`.
  - UI exposes node candidates with health, freshness, schedulable, container count, active reservations and blocking reasons.
- Why this task is needed now:
  - Runtime node HA work made multi-node deployment possible, but AWD currently still depends on same Docker bridge network semantics.
  - Without reservation, a contest can appear ready and then fail at prewarm/start because no node has enough bindable capacity.
  - Without fixed placement, node failover can rebuild team services on different nodes and break alias-based AWD reachability.

## Inputs

- Source docs:
  - `docs/文档规范.md`
  - `docs/plan/README.md`
  - `docs/plan/impl-plan/README.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/architecture/backend/design/awd-engine-migration.md`
  - `docs/architecture/backend/design/instance-sharing.md`
- Related architecture/contracts:
  - `docs/contracts/openapi-v1/`
  - `docs/contracts/openapi-v1.yaml`
  - `code/backend/migrations/000001_init_schema.up.sql`
  - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - `code/backend/internal/module/instance/entity/instance.go`
  - `code/backend/internal/module/contest/entity/contest_awd_service.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-12-true-ha-group/runtime-node-health-and-failover-rebuild.md`
  - `docs/plan/impl-plan/2026-06-12-true-ha-group/INDEX.md`
  - `docs/todos/2026-05-16-awd-runtime-followup.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Crosses backend schema, runtime node health semantics, contest APIs, practice provisioning, app composition failover, frontend contracts/UI, OpenAPI and architecture docs.
  - Changes user-visible admin preflight workflow and runtime failure behavior.
  - Requires independent review before merge because wrong behavior can break live AWD competitions.

## Files

- Create:
  - `code/backend/migrations/000020_create_awd_runtime_placement_reservations.up.sql`
  - `code/backend/migrations/000020_create_awd_runtime_placement_reservations.down.sql`
  - `code/backend/internal/module/container_runtime/entity/runtime_node_reservation.go`
  - `code/backend/internal/module/contest/entity/contest_runtime_placement.go`
  - `code/backend/internal/module/contest/entity/awd_runtime_reconcile_override.go`
  - `code/backend/internal/module/container_runtime/application/node_eligibility.go`
  - `code/backend/internal/module/container_runtime/application/runtime_node_reservation_service.go`
  - `code/backend/internal/module/contest/application/commands/awd_runtime_placement_service.go`
  - `code/backend/internal/module/contest/application/queries/awd_runtime_placement_query.go`
  - `code/frontend/src/features/contest-awd-admin/model/useAwdRuntimePlacement.ts`
  - `code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePlacementPanel.vue`
- Modify:
  - `code/backend/internal/config/container.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/ports/ports.go`
  - `code/backend/internal/module/container_runtime/runtime/module.go`
  - `code/backend/internal/module/contest/contracts/*`
  - `code/backend/internal/module/contest/ports/*`
  - `code/backend/internal/module/contest/infrastructure/*`
  - `code/backend/internal/module/contest/api/http/*`
  - `code/backend/internal/module/contest/runtime/wiring.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/app/composition/runtime_node_failover.go`
  - `code/backend/internal/app/router_admin_contest_awd_routes.go`
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/admin/contest-support.ts`
  - `code/frontend/src/api/admin/contest-operations.ts`
  - `code/frontend/src/api/admin/contest-awd-admin.ts`
  - `code/frontend/src/features/contest-workbench/model/useContestEditAwdWorkspace.ts`
  - `code/frontend/src/features/platform/contest-manage/ui/ContestAwdPreflightPanel.vue`
  - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
  - `docs/contracts/openapi-v1/**/*`
  - `docs/contracts/openapi-v1.yaml`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
- Review:
  - `code/backend/internal/module/contest/application/commands/awd_readiness_gate_trace.go`
  - `code/backend/internal/module/contest/application/jobs/status_awd_readiness.go`
  - `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
  - `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts`
  - `code/frontend/src/features/contest-awd-admin/model/useAwdContestSnapshotLoader.ts`
  - `code/frontend/src/features/platform/contest-manage/model/useAwdStartOverrideFlow.ts`
- Test:
  - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - `code/backend/internal/module/container_runtime/application/runtime_node_reservation_service_test.go`
  - `code/backend/internal/module/contest/application/commands/awd_runtime_placement_service_test.go`
  - `code/backend/internal/module/contest/application/queries/awd_runtime_placement_query_test.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/app/composition/runtime_node_failover_wiring_test.go`
  - `code/backend/internal/app/router_route_wiring_test.go`
  - `code/frontend/src/api/__tests__/admin-awd-runtime-placement.test.ts`
  - `code/frontend/src/features/contest-awd-admin/model/useAwdRuntimePlacement.test.ts`
  - `code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePlacementPanel.test.ts`
  - `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.awd-preflight.test.ts`

## 复用与 Owner 决策

- Existing patterns searched:
  - Runtime node health and selector: `container_runtime/application/node_health_service.go`, `container_runtime/infrastructure/node_repository.go`.
  - AWD instance scope and network alias: `practice/ports/ports.go`, `practice/application/commands/runtime_container_create.go`, `awd_runtime_rules.go`.
  - Node offline wiring: `app/composition/runtime_node_failover.go`, `instance/application/commands/maintenance_service.go`.
  - Admin AWD readiness/UI: `contest/application/queries/awd_readiness_query.go`, `router_admin_contest_awd_routes.go`, `features/contest-awd-admin`, `features/platform/contest-manage`.
- Reuse / extend / split / create-new decision:
  - Extend `contest` with placement command/query service instead of overloading readiness.
  - Extend `container_runtime` with generic node eligibility and reservation service; do not put contest/team/service semantics into runtime module.
  - Extend `practice` runtime selector adapter to use scope-specific placement; do not fork the scheduler.
  - Add UI panel beside existing preflight/readiness surface; do not hide reservation behind checker readiness summary.
- Owner boundary:
  - `container_runtime`: node health, freshness, schedulable, capacity snapshot, reservation rows, capacity math.
  - `contest`: contest-level placement contract, reserved units calculation, preflight reserve/rebind/emergency commands, audit semantics.
  - `practice`: instance provisioning consumes placement and applies one-scope degraded override.
  - `instance`: offline requeue remains node-scoped but must not clear contest placement reservation.
  - Frontend `features/contest-awd-admin`: runtime placement UI and API normalization; `features/platform/contest-manage` embeds it in preflight.
- Why this is the narrowest safe surface:
  - The change follows existing module boundaries and only adds new APIs where a new admin operation is required.
  - It avoids replacing the scheduler or introducing distributed networking before the Docker bridge alias model is changed.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming` was used first to resolve semantic choices: `single_node`, `docker_bridge_alias`, explicit reserve, health eligibility, rebind behavior and node offline behavior.
- Why this pass fits:
  - The task is a behavior/design change with irreversible data/API consequences.
- grill-with-docs findings:
  - Current docs describe default new scheduling as `ready/degraded + fresh + schedulable`; this is too broad for AWD placement.
  - `degraded` needs separate execution/provisioning semantics.
  - Readiness is already a checker/service readiness concept; reservation must not be an implicit side effect of `GET readiness`.
  - Current node failover docs imply AWD can rebuild on any healthy node, which conflicts with `single_node + docker_bridge_alias`.
- Plan adjustments after challenge:
  - Add separate `runtime_node_reservations` and `contest_runtime_placements`.
  - Require explicit preflight reserve and block start/prewarm/round/checker without active valid reservation.
  - Make rebind explicit and split pre-start rebind from emergency active-runtime recreate.
  - Add first-version UI for node capacity and health.

## Confirmed Semantics

- `network_mode=docker_bridge_alias` and `placement_mode=single_node` are persisted for AWD from the first version.
- `docker_bridge_alias` requires all `team × service` containers of one contest to share the same Docker host network.
- `GET readiness` remains read-only.
- `POST preflight-reserve` is the only normal operation that creates or refreshes reservation.
- `start contest`, `prewarm`, `create round`, `run checker`, student start/restart and admin start/restart all require an active valid reservation.
- Force override for checker readiness must not bypass runtime reservation blockers.
- `reserved_units = approved_team_count * visible_awd_service_count`.
- `available_for_contest = degraded_container_threshold - current_managed_containers - active_reserved_units_on_target_node_excluding_this_contest`.
- `degraded_container_threshold=0` means degraded capacity gating disabled; node never becomes degraded due to container count threshold.
- `execution_eligible = ready/degraded + fresh`.
- `provisioning_eligible = ready + fresh + schedulable`.
- `placement_eligible = provisioning_eligible + available capacity`.
- `degraded + fresh` allows old container execution, not ordinary provisioning.
- If only degraded nodes exist, ordinary AWD provisioning stays pending.
- One-scope manual degraded override may allow one rebuild on the bound degraded/fresh node.
- Node offline does not release reservation.
- Cleaning some instances releases actual container capacity but not contest reservation.
- Ordinary rebind is whole-contest pre-start/no-active-runtime only.
- Runtime migration during a contest is `emergency rebind/recreate all`: clean old contest AWD runtime, release old reservation, create new reservation, rebuild all `team × service`, audited as an outage operation.

## Execution Slices

### Slice 1: Runtime Node Eligibility And Capacity Semantics

- Goal: Make node health semantics explicit before reservations consume them.
- Dependencies: none.
- Files:
  - Create:
    - `code/backend/internal/module/container_runtime/application/node_eligibility.go`
  - Modify:
    - `code/backend/internal/config/container.go`
    - `code/backend/internal/module/container_runtime/application/node_health_service.go`
    - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
    - `code/backend/internal/module/container_runtime/ports/ports.go`
  - Review:
    - `docs/architecture/backend/01-system-architecture.md`
    - `docs/architecture/backend/05-key-flows.md`
  - Test:
    - `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
    - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
- Steps:
  - [ ] Step 1: Add failing unit tests for eligibility.
    - Cover `ready/fresh/schedulable` as provisioning eligible.
    - Cover `degraded/fresh` as execution eligible but not provisioning eligible.
    - Cover stale/unknown/offline as neither eligible.
    - Cover `degraded_container_threshold=0` disabled.
  - [ ] Step 2: Run targeted tests and confirm failure.
    - Run: `cd code/backend && go test ./internal/module/container_runtime/application ./internal/module/container_runtime/infrastructure -run 'TestRuntimeNodeEligibility|TestNodeHealth' -count=1`
    - Expected: FAIL due missing eligibility/degraded threshold behavior.
  - [ ] Step 3: Implement `NodeEligibility` helpers and config default.
    - Add typed reasons such as `node_not_ready`, `node_degraded`, `node_stale`, `node_unschedulable`, `insufficient_capacity`.
    - Keep CPU/memory in snapshot only.
  - [ ] Step 4: Update node health to mark degraded only when container threshold is configured and exceeded.
  - [ ] Step 5: Update default selector to use provisioning eligibility, not `ready/degraded`.
  - [ ] Step 6: Run targeted tests.
    - Run: `cd code/backend && go test ./internal/module/container_runtime/application ./internal/module/container_runtime/infrastructure -count=1`
    - Expected: PASS.
  - [ ] Step 7: Commit.
    - `git add code/backend/internal/config code/backend/internal/module/container_runtime`
    - `git commit -m "feat(runtime): 拆分节点执行与调度健康语义" -m "新增 runtime node eligibility 规则，degraded 节点只允许旧容器执行，不再进入默认 provisioning。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - `cd code/backend && go test ./internal/module/container_runtime/application ./internal/module/container_runtime/infrastructure -count=1`
- Review focus:
  - `degraded` 是否仍被默认新调度选中。
  - 心跳 freshness 是否仍由同一个 stale threshold 控制。
- Done criteria:
  - Node eligibility 有独立 helper、测试和 blocking reason。

### Slice 2: Placement / Reservation Persistence And Services

- Goal: Persist bottom-level reservation and contest-level placement.
- Dependencies: Slice 1.
- Files:
  - Create:
    - `code/backend/migrations/000020_create_awd_runtime_placement_reservations.up.sql`
    - `code/backend/migrations/000020_create_awd_runtime_placement_reservations.down.sql`
    - `code/backend/internal/module/container_runtime/entity/runtime_node_reservation.go`
    - `code/backend/internal/module/container_runtime/application/runtime_node_reservation_service.go`
    - `code/backend/internal/module/contest/entity/contest_runtime_placement.go`
    - `code/backend/internal/module/contest/entity/awd_runtime_reconcile_override.go`
    - `code/backend/internal/module/contest/application/commands/awd_runtime_placement_service.go`
  - Modify:
    - `code/backend/migrations/000001_init_schema.up.sql`
    - `code/backend/internal/app/migration_files_test.go`
    - `code/backend/internal/app/test_schema_test.go`
    - `code/backend/internal/module/container_runtime/runtime/module.go`
    - `code/backend/internal/module/contest/runtime/module.go`
  - Test:
    - `code/backend/internal/module/container_runtime/application/runtime_node_reservation_service_test.go`
    - `code/backend/internal/module/contest/application/commands/awd_runtime_placement_service_test.go`
- Steps:
  - [ ] Step 1: Write migration tests for table names, indexes and FK references.
    - Tables:
      - `runtime_node_reservations`
      - `contest_runtime_placements`
      - `awd_runtime_reconcile_overrides`
  - [ ] Step 2: Run migration tests and confirm failure.
    - Run: `cd code/backend && go test ./internal/app -run 'Test.*Migration|TestSchema' -count=1`
    - Expected: FAIL due missing migration files/schema registrations.
  - [ ] Step 3: Add migrations.
    - `contest_runtime_placements`: `contest_id`, `network_mode`, `placement_mode`, `node_id`, `reservation_id`, `reserved_units`, `status`, `blocked_reason`, timestamps.
    - `runtime_node_reservations`: `node_id`, `owner_type`, `owner_id`, `reserved_units`, `status`, `expires_at`, timestamps.
    - `awd_runtime_reconcile_overrides`: `contest_id`, `team_id`, `service_id`, `override_type`, `expires_at`, `consumed_at`, `requested_by`, `reason`, timestamps.
  - [ ] Step 4: Add entity structs and repository/service ports.
  - [ ] Step 5: Write failing service tests.
    - No active placement selects eligible node and creates reservation.
    - Active placement refreshes reserved units and does not change node.
    - Reservation excludes own active reservation during capacity refresh.
    - Node offline does not release reservation.
    - Contest ended/draft rollback/scope change releases reservation.
  - [ ] Step 6: Implement minimal services and repositories.
  - [ ] Step 7: Run targeted tests.
    - Run: `cd code/backend && go test ./internal/module/container_runtime/application ./internal/module/contest/application/commands ./internal/app -run 'TestRuntimeNodeReservation|TestAWDRuntimePlacement|Test.*Migration' -count=1`
    - Expected: PASS.
  - [ ] Step 8: Commit.
    - `git add code/backend/migrations code/backend/internal/app code/backend/internal/module/container_runtime code/backend/internal/module/contest`
    - `git commit -m "feat(awd): 持久化比赛运行节点绑定与容量预留" -m "新增 contest_runtime_placements 与 runtime_node_reservations，AWD 赛前显式 reserve 后才拥有宿主容量。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Migration tests and service tests pass.
- Review focus:
  - Reservation row status/index/FK design.
  - Capacity calculation excludes own reservation.
  - Race safety around reserve/rebind transactions.
- Done criteria:
  - DB and domain services can reserve, refresh, release and query placement without touching practice scheduler.

### Slice 3: Admin APIs, Readiness Gate, Rebind And Emergency Commands

- Goal: Expose explicit admin operations and block runtime actions without active reservation.
- Dependencies: Slice 2.
- Files:
  - Create:
    - `code/backend/internal/module/contest/application/queries/awd_runtime_placement_query.go`
  - Modify:
    - `code/backend/internal/module/contest/api/http/*`
    - `code/backend/internal/module/contest/application/commands/awd_readiness_gate_trace.go`
    - `code/backend/internal/module/contest/application/jobs/status_awd_readiness.go`
    - `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
    - `code/backend/internal/module/contest/ports/*`
    - `code/backend/internal/app/router_admin_contest_awd_routes.go`
    - `docs/contracts/openapi-v1/**/*`
    - `docs/contracts/openapi-v1.yaml`
  - Test:
    - `code/backend/internal/module/contest/application/queries/awd_runtime_placement_query_test.go`
    - `code/backend/internal/module/contest/application/commands/awd_runtime_placement_service_test.go`
    - `code/backend/internal/app/router_route_wiring_test.go`
- Steps:
  - [ ] Step 1: Add failing API/query tests for node candidates.
    - Candidate fields: `node_name`, `health_status`, `last_seen_at`, `fresh`, `schedulable`, `current_managed_containers`, `active_reserved_units`, `available_units`, `degraded_container_threshold`, `eligible_for_reservation`, `blocking_reasons`.
  - [ ] Step 2: Add failing command tests.
    - `preflight-reserve` creates/refreshes.
    - `rebind` with active runtime returns conflict unless emergency command is used.
    - `start contest`, `prewarm`, `create round`, `run current checker` fail with runtime reservation blocker even when readiness override is requested.
  - [ ] Step 3: Add routes.
    - `GET /api/v1/admin/runtime-nodes?purpose=awd_reservation&contest_id=<id>`
    - `GET /api/v1/admin/contests/:id/awd/runtime-placement`
    - `POST /api/v1/admin/contests/:id/awd/runtime-placement/preflight-reserve`
    - `POST /api/v1/admin/contests/:id/awd/runtime-placement/rebind`
    - `POST /api/v1/admin/contests/:id/awd/runtime-placement/emergency-recreate`
  - [ ] Step 4: Add handlers, DTOs, validation and audit resource types.
  - [ ] Step 5: Update OpenAPI split sources and regenerate bundle.
    - Run: `python3 tools/sync_openapi_from_contract.py`
  - [ ] Step 6: Run targeted backend tests.
    - Run: `cd code/backend && go test ./internal/module/contest/... ./internal/app -run 'TestAWDRuntimePlacement|TestRouter' -count=1`
  - [ ] Step 7: Commit.
    - `git add code/backend/internal/module/contest code/backend/internal/app docs/contracts`
    - `git commit -m "feat(awd): 增加运行节点预留与重绑定接口" -m "新增 preflight reserve、runtime placement 查询和 rebind/emergency recreate 管理入口，并把 runtime reservation 纳入 AWD 操作门禁。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Backend tests and OpenAPI sync pass.
- Review focus:
  - Force override must not bypass runtime reservation.
  - Route ownership should stay under contest/admin AWD and runtime node candidate query should not leak contest rules into `container_runtime`.
- Done criteria:
  - Admin can inspect candidates, reserve, rebind before runtime exists, and perform audited emergency recreate.

### Slice 4: Practice Scheduler, Desired Reconcile And Offline Behavior

- Goal: Make all AWD runtime creation paths obey placement/reservation.
- Dependencies: Slice 3.
- Files:
  - Modify:
    - `code/backend/internal/module/practice/ports/ports.go`
    - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
    - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
    - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
    - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
    - `code/backend/internal/app/composition/runtime_node_failover.go`
  - Test:
    - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
    - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
    - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
    - `code/backend/internal/app/composition/runtime_node_failover_wiring_test.go`
- Steps:
  - [ ] Step 1: Add failing tests for scope-aware runtime selection.
    - AWD scope with active placement uses fixed `node_id`.
    - AWD scope without active placement stays pending / returns runtime placement required.
    - Non-AWD scope still uses default runtime selector.
  - [ ] Step 2: Add failing tests for node offline.
    - Offline node requeues instances but does not release reservation.
    - Single-node AWD desired reconcile does not rebuild elsewhere.
    - Original node restored `ready` lets pending scopes rebuild on original node.
    - Original node restored `degraded` blocks ordinary rebuild.
  - [ ] Step 3: Add failing tests for manual degraded override.
    - One `contest_id + team_id + service_id` override permits one rebuild on degraded/fresh bound node.
    - Override cannot be used for offline/stale/unknown.
    - Consumed override is audited and cannot be reused.
  - [ ] Step 4: Implement practice port extension for `RuntimePlacementResolver`.
  - [ ] Step 5: Update composition adapter to branch by AWD scope and consult contest placement service.
  - [ ] Step 6: Ensure desired reconcile reads fixed placement instead of default selector.
  - [ ] Step 7: Update offline failover wiring so AWD placement status becomes `unavailable/degraded/active` but node reservation remains active.
  - [ ] Step 8: Run targeted tests.
    - Run: `cd code/backend && go test ./internal/module/practice/application/commands ./internal/module/instance/application/commands ./internal/app/composition -run 'TestAWD|TestRuntimeNodeFailover|TestRuntimeMaintenance' -count=1`
  - [ ] Step 9: Commit.
    - `git add code/backend/internal/module/practice code/backend/internal/module/instance code/backend/internal/app/composition`
    - `git commit -m "feat(awd): 让实例补齐遵守比赛运行节点绑定" -m "AWD desired reconcile 与手动启动统一读取 contest runtime placement，节点离线后只回到原绑定节点恢复，不再自动漂移。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Practice/instance/composition targeted tests pass.
- Review focus:
  - All AWD creation paths are covered: desired reconcile, prewarm, admin start/restart, student start/restart, offline pending rebuild, degraded override, emergency recreate.
  - Non-AWD scheduling behavior remains unchanged.
- Done criteria:
  - No AWD path can silently fall back to default runtime selector when placement is required.

### Slice 5: Frontend Contracts And Admin Preflight UI

- Goal: Show runtime node health/capacity and expose reserve/rebind operations in first UI version.
- Dependencies: Slice 3 API contract; can run after Slice 4 if backend shape changed.
- Files:
  - Create:
    - `code/frontend/src/features/contest-awd-admin/model/useAwdRuntimePlacement.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePlacementPanel.vue`
  - Modify:
    - `code/frontend/src/api/contracts.ts`
    - `code/frontend/src/api/admin/contest-support.ts`
    - `code/frontend/src/api/admin/contest-operations.ts`
    - `code/frontend/src/api/admin/contest-awd-admin.ts`
    - `code/frontend/src/features/contest-workbench/model/useContestEditAwdWorkspace.ts`
    - `code/frontend/src/features/platform/contest-manage/ui/ContestAwdPreflightPanel.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
  - Test:
    - `code/frontend/src/api/__tests__/admin-awd-runtime-placement.test.ts`
    - `code/frontend/src/features/contest-awd-admin/model/useAwdRuntimePlacement.test.ts`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDRuntimePlacementPanel.test.ts`
    - `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.awd-preflight.test.ts`
- Steps:
  - [ ] Step 1: Add API normalization tests.
    - Verify candidate fields normalize string IDs, timestamps, booleans and blocking reason arrays.
  - [ ] Step 2: Add composable tests.
    - Load placement + candidates.
    - Reserve selected node.
    - Auto reserve when no selected node.
    - Rebind requires explicit action and blocks active runtime response.
  - [ ] Step 3: Add UI tests.
    - Show health/status/capacity rows.
    - Disable reserve button when no eligible node.
    - Display blocked reasons.
    - Show active placement summary and reserved units.
  - [ ] Step 4: Implement API functions and types.
  - [ ] Step 5: Implement composable and UI panel.
    - Use compact operational layout, not marketing copy.
    - Use existing feature/admin panel styling.
  - [ ] Step 6: Mount panel in AWD preflight and operations pre-runtime surfaces.
  - [ ] Step 7: Run targeted frontend tests.
    - Run: `cd code/frontend && npm run test:unit -- admin-awd-runtime-placement useAwdRuntimePlacement AWDRuntimePlacementPanel ContestEdit.awd-preflight`
  - [ ] Step 8: Run frontend guard.
    - Run: `bash scripts/check-frontend-test-guard.sh`
  - [ ] Step 9: Commit.
    - `git add code/frontend`
    - `git commit -m "feat(frontend): 展示 AWD 运行节点容量与预留操作" -m "赛前检查页新增 runtime placement 面板，管理员可查看节点健康容量并显式 reserve 或 rebind。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Targeted Vitest and frontend guard pass.
- Review focus:
  - UI must not imply `GET readiness` already reserves capacity.
  - Active-runtime rebind wording must make outage/emergency semantics explicit.
- Done criteria:
  - Admin can complete reserve from UI and understand why a node is blocked.

### Slice 6: Runtime Operations, Release Hooks And Lifecycle Cleanup

- Goal: Release or preserve reservations at the correct lifecycle points.
- Dependencies: Slices 2-4.
- Files:
  - Modify:
    - `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
    - `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go`
    - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - Test:
    - `code/backend/internal/module/contest/application/statusmachine/*_test.go`
    - `code/backend/internal/module/contest/infrastructure/*runtime_cleaner*_test.go`
    - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- Steps:
  - [ ] Step 1: Add tests for release triggers.
    - Contest ended releases reservation after AWD runtime cleanup.
    - Draft rollback/cancel releases placement.
    - Visible service count/team approval changes require reserve refresh before runtime operations continue.
  - [ ] Step 2: Add tests that cleanup instances does not release contest reservation.
  - [ ] Step 3: Implement lifecycle hooks.
  - [ ] Step 4: Run targeted tests.
    - Run: `cd code/backend && go test ./internal/module/contest/... ./internal/module/practice/application/commands -run 'Test.*RuntimePlacement|Test.*Reservation|Test.*Ended' -count=1`
  - [ ] Step 5: Commit.
    - `git add code/backend/internal/module/contest code/backend/internal/module/practice code/backend/internal/module/instance`
    - `git commit -m "feat(awd): 对齐运行节点预留的生命周期释放" -m "比赛结束、草稿回滚和作用域变更显式处理 reservation，实例清理只释放容器容量不释放比赛预留。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Contest lifecycle and cleanup tests pass.
- Review focus:
  - No accidental release on node offline or instance cleanup.
  - Scope changes force reserve refresh rather than silently under-reserving.
- Done criteria:
  - Reservation lifecycle follows confirmed semantics.

### Slice 7: Documentation And Contracts

- Goal: Move implemented decisions from plan into current fact sources.
- Dependencies: implementation slices complete enough to reflect actual code.
- Files:
  - Modify:
    - `docs/architecture/backend/01-system-architecture.md`
    - `docs/architecture/backend/03-container-architecture.md`
    - `docs/architecture/backend/05-key-flows.md`
    - `docs/architecture/backend/design/awd-engine-migration.md`
    - `docs/operations/runtime-agent-deployment.md`
    - `docs/contracts/openapi-v1.yaml`
    - `docs/contracts/openapi-v1/**/*`
- Steps:
  - [ ] Step 1: Update architecture docs to state `single_node + docker_bridge_alias` placement/reservation.
  - [ ] Step 2: Replace existing “AWD rebuilds on healthy node” wording with “recovers only to original bound node unless emergency recreate”.
  - [ ] Step 3: Document eligibility split and degraded semantics.
  - [ ] Step 4: Document reserve/rebind/emergency operation behavior in operations doc.
  - [ ] Step 5: Regenerate OpenAPI bundle if contract source changed.
  - [ ] Step 6: Run doc checks.
    - Run: `python3 scripts/check-docs-consistency.py`
    - Run: `git diff --check`
  - [ ] Step 7: Commit.
    - `git add docs`
    - `git commit -m "docs(awd): 记录运行节点绑定与预留语义" -m "架构、运行说明和契约同步 single_node AWD placement、reservation、degraded eligibility 与 emergency recreate 行为。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Docs consistency and diff check pass.
- Review focus:
  - Docs are current facts only after code lands.
  - No plan-only future behavior is written as implemented fact.
- Done criteria:
  - Architecture and operations docs no longer contradict placement semantics.

### Slice 8: Full Validation, Independent Review And Archive

- Goal: Prove the task is ready to merge and preserve review evidence.
- Dependencies: Slices 1-7.
- Files:
  - Create:
    - `docs/reviews/backend/2026-06-13-awd-runtime-placement-reservation-round-1.md`
    - `docs/reviews/backend/2026-06-13-awd-runtime-placement-reservation-round-2.md` if fixes are needed.
  - Modify:
    - `docs/plan/impl-plan/2026-06-13-awd-runtime-placement-reservation-implementation-plan.md`
- Steps:
  - [ ] Step 1: Run backend targeted suite.
    - Run: `cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/contest/... ./internal/module/practice/... ./internal/module/instance/... ./internal/app/... -count=1`
  - [ ] Step 2: Run frontend targeted suite.
    - Run: `cd code/frontend && npm run test:unit -- admin-awd-runtime-placement useAwdRuntimePlacement AWDRuntimePlacementPanel ContestEdit.awd-preflight`
  - [ ] Step 3: Run frontend guard.
    - Run: `bash scripts/check-frontend-test-guard.sh`
  - [ ] Step 4: Run OpenAPI sync check.
    - Run: `python3 tools/sync_openapi_from_contract.py --check` if supported; otherwise run sync and ensure no diff.
  - [ ] Step 5: Run workflow checks.
    - Run: `bash scripts/check-startup-gate.sh`
    - Run: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
    - Run: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - [ ] Step 6: Prepare independent review handoff.
    - Include commit range, this plan path, changed files, validation commands and risk focus.
  - [ ] Step 7: Run independent `code-reviewer` gate.
    - Save each formal review as `docs/reviews/backend/2026-06-13-awd-runtime-placement-reservation-round-<n>.md`.
    - Each review must bind the reviewed commit hash/range.
  - [ ] Step 8: Fix blocker findings, then run a new review round with `round-<n+1>`.
  - [ ] Step 9: Archive plan after implemented facts are absorbed.
    - Run: `bash harness/workflow-plugins/code-workflow/archive_task_artifacts.sh`
  - [ ] Step 10: Commit review/final plan changes.
    - `git add docs/reviews docs/plan .harness/session-gates`
    - `git commit -m "docs(review): 归档 AWD 运行节点预留实现评审" -m "保留多轮 review 记录并绑定对应提交，完成 implementation plan 收尾归档。" -m "Task: 2026-06-13-awd-runtime-placement-reservation"`
- Validation:
  - Completion-full and independent review pass.
- Review focus:
  - Reviewer must check runtime reservation cannot be bypassed by force override.
  - Reviewer must check every AWD runtime creation path.
  - Reviewer must check node offline behavior and docs alignment.
- Done criteria:
  - Review findings are either fixed or explicitly non-blocking with evidence; plan is archived through workflow script.

## Impact And Compatibility

- API / DTO:
  - New admin runtime node candidate and AWD runtime placement endpoints.
  - Existing readiness response may gain runtime placement summary only if named clearly as read-only state; reservation remains mutation endpoint.
  - Readiness override payloads cannot bypass runtime reservation errors.
- Data / migration:
  - Adds three tables with FK to contests/runtime_nodes/teams/services/users.
  - Existing contests have no placement until preflight reserve is executed.
- State / cache / queue / event:
  - Desired reconcile remains scheduler-owned; no new queue.
  - Manual degraded override can be persisted in DB rather than Redis because it is audited user/admin state.
- Runtime / config:
  - Adds `container.runtime_node_health.degraded_container_threshold`, default `0`.
  - No change to runtime-agent protocol unless candidate query needs additional managed container stats already present in capacity snapshot.
- Frontend route / state / UX:
  - AWD preflight and operations panels show runtime placement state.
  - Admin reserve/rebind actions are explicit buttons.
- Docs / contracts:
  - OpenAPI v1 and backend architecture docs must be updated in the same task.

## Plan Review / Architecture Fit

- Target owner boundary:
  - The plan keeps node eligibility/capacity in `container_runtime`, contest placement in `contest`, runtime consumption in `practice`, and node offline requeue in `instance`.
- Reuse points / landing zones:
  - Reuse existing `RuntimeNodeRepository`, node health service, AWD readiness gate shape, admin AWD route group, and contest-workbench preflight UI.
- Known structural debt touched:
  - Current `practiceRuntimeNodeSelectorAdapter` ignores scope.
  - Current docs overstate AWD rebuild freedom after node offline.
  - Current degraded semantics are too coarse.
- How this plan avoids behavior-only convergence:
  - It adds persisted placement/reservation facts and forces provisioning selectors to consume them; it does not just add UI warnings.
- Hidden second-redesign risk:
  - Distributed AWD will still require a new network mode and proxy/overlay design. This plan explicitly persists `network_mode` and `placement_mode` so that future distributed support is a new mode, not a reinterpretation.
- Decision after review:
  - Plan is fit for implementation. The first implementation slice should start at node eligibility because every later capacity and reservation decision depends on it.

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/architecture/backend/design/awd-engine-migration.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/contracts/openapi-v1/`
- Fact sources to update after implementation:
  - Same as above, plus OpenAPI bundle.
- Plan-only notes that must not become architecture source:
  - Unimplemented emergency recreate internals.
  - Future distributed AWD overlay/proxy network.
  - Any capacity scoring beyond container count threshold.
- Archive condition:
  - Code, tests, contracts and architecture docs are merged; review rounds are saved under `docs/reviews/...round-<n>.md`; plan is moved with `archive_task_artifacts.sh`.

## Validation Plan

- Per-slice commands:
  - Slice 1: `cd code/backend && go test ./internal/module/container_runtime/application ./internal/module/container_runtime/infrastructure -count=1`
  - Slice 2: `cd code/backend && go test ./internal/module/container_runtime/application ./internal/module/contest/application/commands ./internal/app -run 'TestRuntimeNodeReservation|TestAWDRuntimePlacement|Test.*Migration' -count=1`
  - Slice 3: `cd code/backend && go test ./internal/module/contest/... ./internal/app -run 'TestAWDRuntimePlacement|TestRouter' -count=1`
  - Slice 4: `cd code/backend && go test ./internal/module/practice/application/commands ./internal/module/instance/application/commands ./internal/app/composition -run 'TestAWD|TestRuntimeNodeFailover|TestRuntimeMaintenance' -count=1`
  - Slice 5: `cd code/frontend && npm run test:unit -- admin-awd-runtime-placement useAwdRuntimePlacement AWDRuntimePlacementPanel ContestEdit.awd-preflight`
  - Slice 7: `python3 scripts/check-docs-consistency.py && git diff --check`
- Integration commands:
  - `cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/contest/... ./internal/module/practice/... ./internal/module/instance/... ./internal/app/... -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - AWD preflight shows candidates and reserve result.
  - Start/prewarm/round/checker fail before reserve and succeed after valid reserve.
  - Simulated node offline leaves AWD scopes pending until original node recovers or emergency recreate runs.
- Commands intentionally skipped and why:
  - Full E2E browser smoke may be skipped if no dev server/runtime-agent/Docker fixture is available; record the environment reason in validation evidence.

## Validation Evidence

- Command: `待执行`
  - Result: `待执行`
  - Notes: Plan-only change so far; implementation validation is assigned per slice.

## Independent Review Handoff

- Review target:
  - Commit range for `task/2026-06-13-awd-runtime-placement-reservation`.
  - Plan path: `docs/plan/impl-plan/2026-06-13-awd-runtime-placement-reservation-implementation-plan.md`.
- Validation evidence summary:
  - Include per-slice targeted tests, completion-full, OpenAPI sync and docs checks.
- Architecture / contract inputs:
  - Backend architecture docs listed in `Documentation Owner`.
  - OpenAPI v1 split sources and bundle.
  - Runtime node health/failover prior plans.
- Known risks / review focus:
  - `degraded` accidentally entering provisioning.
  - Force override bypassing runtime reservation.
  - AWD desired reconcile falling back to default selector.
  - Node offline releasing reservation or auto-moving contest.
  - UI implying readiness GET performs reservation.
- Project-local checks to consider:
  - `bash scripts/check-startup-gate.sh`
  - `bash scripts/check-frontend-test-guard.sh`
  - `python3 scripts/check-docs-consistency.py`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## Rollback / Recovery

- Safe revert boundary:
  - Each slice has its own commit. Backend migration slice must be reverted with matching down migration before code depending on tables is removed.
- Data / config / runtime recovery notes:
  - Existing contests without placement remain valid drafts but cannot perform AWD runtime operations until `preflight-reserve`.
  - If reservation tables contain bad rows, release rows by contest owner through admin command rather than deleting node rows.
  - `degraded_container_threshold=0` disables degraded-by-container-count if threshold tuning causes false degraded states.
- Irreversible operations:
  - None planned, but emergency recreate destroys and rebuilds active contest AWD runtime. It must require explicit admin action and audit reason.

## Residual Risks

- Risk: The first version does not support distributed AWD networking.
  - Why acceptable: `docker_bridge_alias` cannot satisfy cross-host alias reachability; persisting `network_mode` makes future distributed mode explicit.
  - Follow-up owner, if any: Future architecture/design task for overlay/proxy AWD mode.
- Risk: Capacity uses container count threshold rather than CPU/memory.
  - Why acceptable: CPU/memory metrics remain observable in snapshot, while first implementation needs deterministic reservation semantics.
  - Follow-up owner, if any: Runtime capacity scoring task after production metrics exist.
- Risk: Emergency recreate causes contest outage.
  - Why acceptable: It is explicit, audited, whole-contest, and matches the lack of live migration guarantees.
  - Follow-up owner, if any: Operations runbook update after first emergency drill.
