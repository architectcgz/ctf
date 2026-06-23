<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# AWD Emergency Runtime Recreate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to execute this plan task-by-task in this task worktree. This plan intentionally stops before implementation until the task is bound through `code-workflow`. Production code changes require TDD: write each failing test, run it red, then implement the smallest code to pass.

**Goal:** Add a controlled emergency operation that explicitly moves one AWD contest to a different schedulable runtime node and recreates the contest runtime on that node.

**Architecture:** Keep `contest_runtime_placements` as the contest-owned placement fact, `instance` as the owner of per-instance runtime identity, and `practice` as the owner of desired AWD runtime reconciliation. Add a composition-level orchestration service plus an internal CLI/runbook entrypoint; do not expose an admin HTTP API or UI in v1. The operation is disruptive by design: it abandons old runtime identity, creates a fresh batch of AWD containers on the target runtime node, and treats old-node residual cleanup as a follow-up operational check.

**Tech Stack:** Go, GORM, PostgreSQL, Redis desired reconcile state, existing `container_runtime` runtime node health/schedulable repository, existing `contest` placement repository, existing `instance` maintenance service, existing `practice` desired AWD reconciler, package-level Go tests, internal `cmd/*` CLI pattern.

---

## Task Metadata

- Task Slug: `2026-06-21-awd-emergency-runtime-recreate`
- Created At: `2026-06-21T20:35:00+08:00`
- Draft Branch: `multi-instance`
- Plan Type: `formal implementation plan`
- Related TODO: `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`
- Parent Decision: `docs/plan/archive/impl-plan/2026-06/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`

## Plan Status

- Status: `draft-pending-task-gate`
- Coding may start only after:
  - [ ] Run `bash scripts/start-implementation.sh 2026-06-21-awd-emergency-runtime-recreate` to bind task slug, worktree / branch and startup gate.
  - [ ] Re-read this plan in the implementation worktree.
  - [ ] Confirm this v1 scope still excludes admin HTTP API, UI, capacity reservation and automatic residual cleanup.
  - [ ] Complete one independent plan review or explicitly record that independent review was unavailable.

Startup note:

- `2026-06-21`: Plan drafted from brainstorming + `grill-with-docs` discussion. This draft is not an implementation start signal.

## Confirmed Decisions

- Emergency recreate may run while an AWD contest is active; operators must accept a short runtime interruption.
- `target_runtime_node_id` means `runtime_nodes.id`, the platform control-plane runtime node primary key. It usually identifies the runtime-agent / Docker daemon host that will create the new containers.
- `target_runtime_node_id` is not a Docker Swarm node id, Docker engine id, `runtime_nodes.name`, `container_id`, or `network_id`.
- The target node must be eligible for new scheduling: healthy, heartbeat fresh, and `schedulable=true`.
- There is no `--force` bypass for `schedulable=false` in v1. If a node is cordoned, operators must explicitly make it schedulable through runtime node management before using it as the target.
- Emergency recreate does not operate on old containers on the target node. It abandons the old runtime identity and creates a new batch of AWD containers on the target node.
- Old node residual containers / networks are best-effort follow-up cleanup. They do not block recovery.
- Re-running the command with the same target node is allowed as a repair retry when a previous attempt already changed placement but failed during requeue or reconcile.

## Objective And Non-Goals

- Objective:
  - Add a schedulable-healthy target runtime node validator for emergency recreate.
  - Atomically replace a contest's active AWD runtime placement with a new active placement, or no-op when it is already on the target node.
  - Requeue the contest's live AWD instances by clearing old `runtime_node_id / container_id / network_id / runtime_details / access_url` and setting `running / creating` rows back to `pending`.
  - Record AWD service operations for affected instances with `operation_type=recreate`, `status=provisioning`, `reason=emergency_runtime_recreate`.
  - Clear desired AWD reconcile failure state for the contest so old node failures do not suppress new-node rebuild.
  - Trigger existing `ReconcileDesiredAWDInstances()` so `team x visible service` runtime is rebuilt on the new active placement.
  - Provide an internal dry-run-by-default CLI and a runbook.
- Non-Goals:
  - Do not add administrator HTTP API or frontend UI.
  - Do not add full runtime reservation, capacity scoring, capacity preflight, or capacity visualization.
  - Do not add Docker / Swarm / engine internal node identity fields.
  - Do not migrate containers, networks, TCP sessions, SSH sessions, WebSocket sessions, or runtime details from the old node.
  - Do not require old node cleanup to succeed before restoring the contest runtime.
  - Do not introduce a contest-level runtime operation table in v1.

## Problem Statement

- Existing AWD placement intentionally prevents silent cross-node drift: an AWD contest reuses `contest_runtime_placements.runtime_node_id`, and if that node is unavailable, pending creates wait / back off.
- That is correct for normal failover, but an active contest still needs an explicit operational escape hatch when the bound runtime node is unrecoverable or waiting would harm the match.
- Current repository methods can ensure an active placement, but they do not provide a controlled replacement operation.
- Current explicit healthy-by-id runtime node lookup allows unschedulable nodes because it is used for old-container operations. Emergency recreate creates new containers, so it must use schedulable-new-scheduling semantics.
- Desired AWD reconcile suppression can preserve failures caused by the old node. Emergency recreate must clear those state entries before triggering rebuild.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
  - `docs/plan/README.md`
  - `docs/plan/impl-plan/README.md`
  - `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`
  - `code/backend/tests/README.md`
  - `harness/policies/reuse-first.yaml`
- Related architecture / operations docs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/operations/runtime-agent-deployment.md`
- Related code:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/app/composition/contest_runtime_placement_adapter.go`
  - `code/backend/internal/module/contest/entity/contest_runtime_placement.go`
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go`
  - `code/backend/internal/app/composition/runtime_node_failover.go`
  - `code/backend/cmd/storage-gc/main.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Touches runtime operations, contest placement persistence, instance lifecycle, Redis reconcile state and internal operator tooling.
  - A wrong implementation can move an active contest to the wrong Docker host or leave the contest stuck without runtime.
  - Requires TDD, code-workflow startup/review gates and focused validation.

## Files

- Create:
  - `code/backend/internal/app/composition/awd_emergency_runtime_recreate.go`
  - `code/backend/internal/app/composition/awd_emergency_runtime_recreate_test.go`
  - `code/backend/cmd/awd-emergency-runtime-recreate/main.go`
  - `code/backend/cmd/awd-emergency-runtime-recreate/main_test.go`
  - `docs/operations/awd-emergency-runtime-recreate.md`
- Modify:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`: add schedulable-healthy-by-id lookup for new scheduling targets.
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`: cover schedulable target eligibility.
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository.go`: add active placement replacement / no-op repair method.
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository_test.go`: cover replacement transaction and idempotent same-target retry.
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`: add contest-scoped AWD emergency requeue method and operation recording.
  - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`: cover emergency requeue behavior.
  - `code/backend/internal/module/instance/infrastructure/repository.go`: add contest-scoped AWD runtime requeue query.
  - `code/backend/internal/module/instance/infrastructure/repository_test.go`: cover status filters and runtime identity clearing.
  - `code/backend/internal/app/composition/instance_module.go`: expose emergency requeue through `InstanceModule`.
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`: add contest-scoped desired reconcile state clearing helper.
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`: cover state clearing before emergency reconcile.
  - `code/backend/internal/module/practice/ports/ports.go`: add port method only if current desired state store needs a wider bulk clear contract.
  - `code/backend/internal/app/composition/practice_module.go`: expose desired state clear / reconcile operation to orchestration service.
  - `docs/architecture/backend/01-system-architecture.md`: update current facts after implementation.
  - `docs/architecture/backend/03-container-architecture.md`: update runtime node placement and emergency recreate semantics after implementation.
  - `docs/architecture/backend/05-key-flows.md`: add emergency recreate flow after implementation.
  - `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`: mark emergency recreate as tracked or implemented.
- Review:
  - `code/backend/internal/app/router.go`: confirm no HTTP route is added in v1.
  - `code/backend/internal/app/composition/runtime_node_failover.go`: compare normal failover with emergency recreate and keep semantics distinct.
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`: ensure rebuild still uses active placement.
  - `harness/policies/script-layer-manifest.json`: only needed if the implementation chooses `tools/` instead of `cmd/`.
- Test:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository_test.go`
  - `code/backend/internal/module/instance/infrastructure/repository_test.go`
  - `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/app/composition/awd_emergency_runtime_recreate_test.go`
  - `code/backend/cmd/awd-emergency-runtime-recreate/main_test.go`

## Reuse And Owner Decisions

- Existing patterns searched:
  - Runtime node health / schedulable selection: `ListSchedulableHealthyNodes`, `FindSchedulableHealthyNodeByName`, `FindHealthyByID`, `NewDefaultRuntimeNodeSelector`.
  - Placement owner: `ContestRuntimePlacementRepository.FindActiveContestRuntimePlacement`, `EnsureActiveContestRuntimePlacement`.
  - Instance requeue owner: `InstanceMaintenanceService.HandleRuntimeNodeOffline`, `Repository.RequeueLostRuntimesByNode`, `recordSystemAWDOperation`.
  - Desired reconcile owner: `serviceCore.ReconcileDesiredAWDInstances`, desired reconcile failure state store.
  - CLI pattern: `code/backend/cmd/storage-gc/main.go`.
- Reuse / extend / create-new decision:
  - `extend_existing`: add a schedulable-healthy-by-id lookup to runtime node repository instead of reimplementing health rules in orchestration.
  - `extend_existing`: add replacement semantics to contest placement repository; do not create a new placement table.
  - `extend_existing`: add contest-scoped emergency requeue to instance maintenance because instance owns runtime identity and AWD operation recording.
  - `extend_existing`: add contest-scoped desired reconcile state clearing to practice because practice owns desired reconcile state.
  - `create_new_with_reason`: add a composition orchestration service because the operation intentionally coordinates contest placement, instance runtime identity and practice reconcile without moving those owners.
  - `create_new_with_reason`: add a dedicated internal CLI command because v1 is runbook-driven and should not expose HTTP/UI.
- Owner boundary:
  - `container_runtime` owns runtime node health, heartbeat freshness and schedulable eligibility.
  - `contest` owns contest-level active/released placement persistence.
  - `instance` owns per-instance runtime identity and requeue / AWD operation records.
  - `practice` owns desired AWD reconcile state and `team x visible service` rebuild.
  - `app/composition` owns cross-module orchestration only.
- Why this is the narrowest safe surface:
  - The operation can reuse existing pending scheduler and desired reconciler for actual runtime creation.
  - No schema migration is required unless implementation discovers a missing index or audit field.
  - No UI/API contract is introduced before the operational procedure has been exercised.

## Intake Analysis Gate

- Relevant analysis passes:
  - `brainstorming`: resolved v1 product shape as internal application service + CLI/runbook, not API/UI.
  - `grill-with-docs`: sharpened terminology and rejected `--force` for unschedulable targets.
  - `writing-plans`: this formal implementation plan.
- Findings:
  - `target_runtime_node_id` must stay tied to `runtime_nodes.id`.
  - Docker / engine node identity is not needed for this task.
  - Emergency recreate creates new containers and therefore must use schedulable-new-scheduling eligibility.
  - Same-target re-execution must be supported as repair retry after partial failure.
  - Desired reconcile suppression state must be cleared or old node failures can delay recovery.

## Target Behavior

CLI dry-run example:

```bash
cd code/backend
go run ./cmd/awd-emergency-runtime-recreate \
  -env prod \
  -contest-id 42 \
  -target-runtime-node-id 7 \
  -reason "runtime node 3 disk failure"
```

Expected dry-run output:

```text
mode=dry-run contest_id=42 current_runtime_node_id=3 target_runtime_node_id=7 eligible=true
affected_instances=24 pending_instances=0 desired_scopes=24
execute=false no changes applied
```

CLI execute example:

```bash
cd code/backend
go run ./cmd/awd-emergency-runtime-recreate \
  -env prod \
  -contest-id 42 \
  -target-runtime-node-id 7 \
  -reason "runtime node 3 disk failure" \
  -execute
```

Expected execute behavior:

- Active placement is on `target_runtime_node_id=7`.
- `running / creating` AWD instances for contest `42` are reset to `pending` and have empty runtime identity.
- Existing `pending` AWD instances stay pending.
- `stopping / stopped / expired / failed` instances are not pulled back.
- Desired reconcile failure state for contest `42` is cleared.
- Existing desired reconciler is invoked.
- CLI exits non-zero if target node is absent, unhealthy, heartbeat stale, or `schedulable=false`.

## Execution Slices

### Slice 1: Target Runtime Node Eligibility

- Goal:
  - Add a single repository-level target eligibility method for emergency recreate and future new-scheduling-by-id flows.
- Files:
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`
- Steps:
  - [ ] Write failing test `TestRuntimeNodeRepositoryFindSchedulableHealthyByIDRejectsUnschedulableNode`.
  - [ ] Run: `cd code/backend && go test ./internal/module/container_runtime/infrastructure -run TestRuntimeNodeRepositoryFindSchedulableHealthyByIDRejectsUnschedulableNode -count=1`
  - [ ] Confirm RED: method does not exist or unschedulable node is accepted.
  - [ ] Implement `FindSchedulableHealthyByID(ctx, nodeID, staleThreshold, now)` using existing `ready/degraded`, `last_seen_at` freshness and `schedulable=true` rules.
  - [ ] Add positive coverage for a fresh ready schedulable node.
  - [ ] Re-run focused tests and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/container_runtime/infrastructure -run 'TestRuntimeNodeRepositoryFindSchedulableHealthyByID|TestRuntimeNodeRepositoryListSchedulableHealthyNodes' -count=1`
- Done criteria:
  - Emergency recreate has a target validator that cannot accidentally use old-container explicit-route semantics.

### Slice 2: Contest Placement Replacement

- Goal:
  - Replace a contest's active placement atomically and support same-target repair retry.
- Files:
  - Modify: `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository.go`
  - Modify: `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository_test.go`
- Steps:
  - [ ] Write failing repository test `TestContestRuntimePlacementRepositoryReplaceActivePlacementReleasesOldAndCreatesNew`.
  - [ ] Run: `cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestRuntimePlacementRepositoryReplaceActivePlacementReleasesOldAndCreatesNew -count=1`
  - [ ] Confirm RED: replacement method does not exist.
  - [ ] Implement `ReplaceActiveContestRuntimePlacement(ctx, contestID, runtimeNodeID)` with a DB transaction.
  - [ ] In the transaction, lock active placement rows for the contest, set old active row to `released` with `released_at`, and create the new active row.
  - [ ] Write failing test `TestContestRuntimePlacementRepositoryReplaceActivePlacementNoopsWhenAlreadyTarget`.
  - [ ] Run the same package focused tests and confirm RED for same-target handling.
  - [ ] Make same-target replacement return the existing active placement with `changed=false` and no extra active row.
  - [ ] Re-run focused tests and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestRuntimePlacementRepository -count=1`
- Done criteria:
  - A contest has at most one active placement.
  - Replacement is safe to retry after placement already points at the target node.

### Slice 3: Instance Emergency Requeue

- Goal:
  - Add contest-scoped AWD runtime identity abandonment and requeue under the `instance` owner.
- Files:
  - Modify: `code/backend/internal/module/instance/infrastructure/repository.go`
  - Modify: `code/backend/internal/module/instance/infrastructure/repository_test.go`
  - Modify: `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - Modify: `code/backend/internal/module/instance/application/commands/runtime_maintenance_service_test.go`
  - Modify: `code/backend/internal/app/composition/instance_module.go`
- Steps:
  - [ ] Write failing repository test `TestRepositoryRequeueAWDContestRuntimesOnlyRequeuesLiveRuntimeRows`.
  - [ ] Run: `cd code/backend && go test ./internal/module/instance/infrastructure -run TestRepositoryRequeueAWDContestRuntimesOnlyRequeuesLiveRuntimeRows -count=1`
  - [ ] Confirm RED: method does not exist.
  - [ ] Implement repository method that selects rows with `contest_id = ?`, `team_id IS NOT NULL`, `service_id IS NOT NULL`, `status IN ('creating','running')`, `expires_at > now`.
  - [ ] Update selected rows to `pending`, clear `runtime_node_id`, `container_id`, `network_id`, `runtime_details`, `access_url`, and update `updated_at`.
  - [ ] Keep existing `pending`, `stopping`, `stopped`, `expired`, `failed` rows unchanged.
  - [ ] Re-run repository focused test and confirm GREEN.
  - [ ] Write failing maintenance service test proving emergency requeue records one `recreate/provisioning` AWD operation per requeued instance with reason `emergency_runtime_recreate`.
  - [ ] Run: `cd code/backend && go test ./internal/module/instance/application/commands -run TestRuntimeMaintenanceServiceHandlesAWDEmergencyRuntimeRecreate -count=1`
  - [ ] Confirm RED.
  - [ ] Add `HandleAWDEmergencyRuntimeRecreate(ctx, contestID, reason)` to `InstanceMaintenanceService`.
  - [ ] Expose it through `InstanceModule`.
  - [ ] Re-run focused tests and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/instance/infrastructure ./internal/module/instance/application/commands -run 'TestRepositoryRequeueAWDContestRuntimes|TestRuntimeMaintenanceServiceHandlesAWDEmergencyRuntimeRecreate' -count=1`
- Done criteria:
  - Instance runtime identity is abandoned by the instance owner.
  - Requeued rows are auditable through existing AWD service operations.

### Slice 4: Desired Reconcile State Reset

- Goal:
  - Clear stale failure/backoff/suppression state for a contest before emergency reconcile.
- Files:
  - Modify: `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - Modify: `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - Modify: `code/backend/internal/module/practice/ports/ports.go` only if needed.
  - Modify: `code/backend/internal/app/composition/practice_module.go`
- Steps:
  - [ ] Write failing test `TestDesiredAWDReconcilerClearsContestFailureStateForEmergencyRecreate`.
  - [ ] Run: `cd code/backend && go test ./internal/module/practice/application/commands -run TestDesiredAWDReconcilerClearsContestFailureStateForEmergencyRecreate -count=1`
  - [ ] Confirm RED: no contest-scoped clear method exists.
  - [ ] Implement `ClearContestDesiredAWDReconcileFailures(ctx, contestID)` by listing contest teams and AWD services, then deleting per `contestID/teamID/serviceID` state keys through the existing store.
  - [ ] Do not delete `AWDScopeControl` rows; manual desired-reconcile suppression remains an operator control and is not the same as failure backoff state.
  - [ ] Expose the method through the practice runtime / composition module.
  - [ ] Re-run focused test and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestDesiredAWDReconciler.*Emergency|TestReconcileDesiredAWD' -count=1`
- Done criteria:
  - Emergency recreate does not inherit old-node failure backoff, while explicit operator suppressions remain intact.

### Slice 5: Composition Orchestration Service

- Goal:
  - Coordinate target validation, placement replacement, instance requeue, desired state clearing and reconcile without moving owners.
- Files:
  - Create: `code/backend/internal/app/composition/awd_emergency_runtime_recreate.go`
  - Create: `code/backend/internal/app/composition/awd_emergency_runtime_recreate_test.go`
  - Modify: `code/backend/internal/app/composition/contest_runtime_placement_adapter.go` if adapters are reused.
- Steps:
  - [ ] Write failing composition test `TestAWDEmergencyRuntimeRecreateDryRunReportsImpactWithoutMutation`.
  - [ ] Run: `cd code/backend && go test ./internal/app/composition -run TestAWDEmergencyRuntimeRecreateDryRunReportsImpactWithoutMutation -count=1`
  - [ ] Confirm RED.
  - [ ] Define input/output structs:

```go
type AWDEmergencyRuntimeRecreateRequest struct {
    ContestID           int64
    TargetRuntimeNodeID int64
    Reason              string
    Execute             bool
}

type AWDEmergencyRuntimeRecreateResult struct {
    ContestID            int64
    PreviousRuntimeNodeID int64
    TargetRuntimeNodeID  int64
    PlacementChanged     bool
    AffectedInstances    int
    RequeuedInstances    int
    ReconcileTriggered   bool
    DryRun               bool
}
```

  - [ ] Implement dry-run path that validates contest mode, active placement and target eligibility, then reports impact without DB changes.
  - [ ] Write failing test `TestAWDEmergencyRuntimeRecreateExecuteSwitchesPlacementThenRequeuesAndReconciles`.
  - [ ] Run the focused composition test and confirm RED.
  - [ ] Implement execute path in this order: validate target, replace/no-op placement, emergency requeue, clear desired state, trigger reconcile.
  - [ ] Write failing test `TestAWDEmergencyRuntimeRecreateAllowsSameTargetRepairRetry`.
  - [ ] Implement same-target repair path: no placement change, still requeue / clear / reconcile.
  - [ ] Write failing test `TestAWDEmergencyRuntimeRecreateRejectsUnschedulableTarget`.
  - [ ] Ensure target validation uses schedulable-healthy-by-id, not healthy-by-id.
  - [ ] Re-run focused tests and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/app/composition -run TestAWDEmergencyRuntimeRecreate -count=1`
- Done criteria:
  - Cross-module orchestration is testable without HTTP.
  - Failure halfway through can be retried against the same target.

### Slice 6: Internal CLI

- Goal:
  - Provide dry-run-by-default operator entrypoint for v1.
- Files:
  - Create: `code/backend/cmd/awd-emergency-runtime-recreate/main.go`
  - Create: `code/backend/cmd/awd-emergency-runtime-recreate/main_test.go`
- Steps:
  - [ ] Write failing CLI argument test proving missing `-contest-id`, `-target-runtime-node-id`, or `-reason` returns an error.
  - [ ] Run: `cd code/backend && go test ./cmd/awd-emergency-runtime-recreate -run TestRunValidatesRequiredFlags -count=1`
  - [ ] Confirm RED.
  - [ ] Implement `flag.NewFlagSet` parsing with `-env`, `-contest-id`, `-target-runtime-node-id`, `-reason`, and `-execute`.
  - [ ] Write failing dry-run output test proving default mode does not execute.
  - [ ] Implement command wiring: load config, open Postgres and Redis, build composition root/modules, build emergency recreate service, call it with `Execute=false` unless `-execute` is set.
  - [ ] Keep output structured enough for runbook copy/paste: mode, contest id, previous node, target node, affected and requeued counts.
  - [ ] Re-run CLI package tests and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./cmd/awd-emergency-runtime-recreate -count=1`
- Done criteria:
  - Operators can preview safely before executing.
  - CLI does not expose an HTTP surface or frontend route.

### Slice 7: Docs, Todo And Guardrails

- Goal:
  - Record the operation as an operator runbook and update current architecture facts.
- Files:
  - Create: `docs/operations/awd-emergency-runtime-recreate.md`
  - Modify: `docs/architecture/backend/01-system-architecture.md`
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/architecture/backend/05-key-flows.md`
  - Modify: `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`
- Steps:
  - [ ] Write runbook with prerequisites: target node must be `healthy + heartbeat fresh + schedulable=true`, operator accepts runtime interruption, old node cleanup is follow-up.
  - [ ] Include dry-run and execute commands.
  - [ ] Include post-checks: active placement target, pending/recreated instance count, desired reconcile logs, residual old-node cleanup check.
  - [ ] Update architecture docs to distinguish normal AWD placement no-drift behavior from explicit emergency recreate.
  - [ ] Update todo item to point at this implementation plan or mark implemented after code lands.
  - [ ] Run documentation consistency checks relevant to changed files.
- Validation:
  - `python3 scripts/check-docs-consistency.py`
  - `bash scripts/check-workflow-governance.sh`
- Done criteria:
  - Future operators can execute and verify the operation without reading the implementation plan.

## Full Validation Gate

Run after all slices:

```bash
cd code/backend && go test ./internal/module/container_runtime/infrastructure \
  ./internal/module/contest/infrastructure \
  ./internal/module/instance/infrastructure \
  ./internal/module/instance/application/commands \
  ./internal/module/practice/application/commands \
  ./internal/app/composition \
  ./cmd/awd-emergency-runtime-recreate -count=1
```

Then run project checks:

```bash
python3 scripts/check-docs-consistency.py
bash scripts/check-workflow-governance.sh
```

If mapper / DTO / contract boundaries are touched unexpectedly, also run:

```bash
cd code/backend && go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1
```

## Architecture-Fit Evaluation

- Target architecture boundary is explicit:
  - `container_runtime` validates node eligibility.
  - `contest` replaces placement.
  - `instance` abandons old runtime identity and records per-instance operations.
  - `practice` clears desired reconcile state and rebuilds through existing reconciler.
  - `app/composition` orchestrates only.
- Shared layers and owners are named:
  - No direct placement SQL in practice.
  - No direct instance status mutation in contest.
  - No scheduler logic in the CLI.
- Structural convergence is not deferred:
  - The only deferred items are intentional non-goals: admin API/UI, capacity reservation, residual cleanup automation and contest-level operation table.
- Immediate second redesign risk:
  - Low for v1 runbook usage because service boundary can be reused by a future admin API without moving persistence ownership.
  - If UI/API is requested next, add an API contract slice rather than changing the service semantics.

## Residual Risks

- If Redis is unavailable, desired reconcile state clearing may fail. The execute path should fail before or after placement replacement based on implementation order; the plan requires same-target repair retry to make this recoverable.
- If target node passes health checks but lacks capacity, existing scheduler/provisioning may still fail. This is accepted in v1 because capacity reservation/preflight is a non-goal.
- If old node later comes back, old residual containers may still exist. The runbook must require a residual cleanup inspection.
- If current contest has manually suppressed desired reconcile controls, emergency recreate must not override them.

## Execution Handoff

Plan execution options after task gate binding:

1. Subagent-driven implementation: execute one slice at a time with review between slices.
2. Inline implementation: use `superpowers:executing-plans` in the bound task worktree and update checkboxes as each step passes.
