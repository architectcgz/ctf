<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# runtime instance port alias cleanup Implementation Plan

**Goal:** Remove `runtime/ports` re-exports of `instance/ports` types so instance access contracts are referenced through their owning package instead of through runtime aliases.

**Architecture:** Keep `runtime/ports` focused on container/runtime capabilities. Instance-owned access, proxy ticket, teacher instance query, and maintenance contracts remain in `instance/ports` or `instance/contracts`; existing runtime infrastructure may still implement some of those ports in this slice, but the dependency becomes explicit instead of hidden behind `runtime/ports`.

**Tech Stack:** Go, CTF modular monolith backend, existing module architecture tests, runtime/app composition tests.

---

## Task Metadata

- Task Slug: `2026-06-09-runtime-instance-port-alias-cleanup`
- Started At: `2026-06-09T04:24:22Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-09-runtime-container-runtime-port-file-split/2026-06-09-runtime-instance-port-alias-cleanup`
- Branch: `task/2026-06-09-runtime-instance-port-alias-cleanup`

## Objective And Non-Goals

- Objective:
  - Add a runtime architecture guard that rejects `runtime/ports` importing `instance/ports`.
  - Move `ManagedContainer` and `ManagedContainerState` ownership back to `runtime/ports`, with `instance/ports` consuming them as runtime capability shapes.
  - Replace production and test references to instance-owned runtime aliases, such as `runtimeports.ProxyTicketClaims`, `RuntimeCleaner`, `TeacherInstanceFilter`, and `ProxyTicketInstanceReader`, with direct `instanceports` imports.
  - Keep behavior, public HTTP API, proxy ticket JSON, runtime-agent protocol shape, and repository queries unchanged.
  - Move the running instance count query out of runtime ports so instance owns the `instances` table read and ops keeps the dashboard-facing consumer port.
  - Remove the stale `instance -> runtime` module dependency baseline by moving instance-side access URL parsing, maintenance container views, and AWD operation input shapes to instance-owned contracts / ports.
  - Remove unused runtime application response mappers that only referenced instance contracts and had no runtime business caller.
- Non-Goals:
  - Do not move the remaining runtime repository methods unrelated to running count or proxy ticket scope out of `runtime/infrastructure` in this slice.
  - Do not delete or rename `runtime/ports/http.go`; file placement cleanup can be a later mechanical slice after aliases are gone.
  - Do not change database schema, runtime-agent messages, HTTP route paths, or DTO response contracts.

## Inputs

- Source docs:
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Related architecture/contracts:
  - `code/backend/internal/module/runtime/ports/http.go`
  - `code/backend/internal/module/runtime/ports/metrics.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/runtime/api/http/handler.go`
  - `code/backend/internal/app/composition/runtime_http_service_adapter.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/instance/infrastructure/{repository.go,awd_target_proxy_repository.go,proxy_ticket_store.go}`
  - `code/backend/internal/module/architecture_baseline_test.go`
- Related prior work:
  - `26f118c5b refactor(runtime): 拆分容器运行时端口文件`
  - `892eadeeb test(runtime): 限制宿主执行器宽接口使用面`
  - `3b6b28afb docs(runtime): 同步容器运行时端口拆分事实`

## Task Classification

- Classification: `非琐碎任务`
- Why: This changes backend module boundary imports across runtime ports, runtime HTTP adapter, app composition, runtime infrastructure, generated mapper inputs, and tests.

## Files

- Create:
  - None expected.
- Modify:
  - `code/backend/internal/module/runtime/architecture_test.go`
  - `code/backend/internal/module/runtime/ports/http.go`
  - `code/backend/internal/module/runtime/ports/metrics.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/runtime/runtime/module.go`
  - `code/backend/internal/module/runtime/application/contracts.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/instance/infrastructure/awd_target_proxy_repository.go`
  - `code/backend/internal/module/instance/infrastructure/proxy_ticket_store.go`
  - `code/backend/internal/module/runtime/api/http/handler.go`
  - `code/backend/internal/app/composition/runtime_http_service_adapter.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - Tests and test adapters that still use runtime aliases for instance-owned types.
- Review:
  - `code/backend/internal/module/runtime/ports/container_inventory_runtime.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/runtime/application/{commands,queries}/response_mapper*.go`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Test:
  - `go test ./internal/module/runtime -run TestRuntimePortsDoNotReexportInstancePorts -count=1`
  - `go test ./internal/module/runtime/ports -count=1`
  - `go test ./internal/module/runtime/... -count=1`
  - `go test ./internal/app -run 'TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestAWDDefenseSSHGateway' -count=1`
  - `go test ./internal/module -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - `instance/ports` already owns instance lookup, proxy ticket, teacher query, runtime cleaner, and maintenance-facing container state contracts.
  - `runtime/ports` owns container provisioning, cleanup, file, image, inventory, stats, interactive, and host executor contracts.
  - The module dependency baseline still records `runtime -> instance`; `instance -> runtime` has been removed after the instance-owned contract cleanup.
- Reuse / extend / split / create-new decision:
  - Reuse `instance/ports` directly for instance-owned types.
  - Define `ManagedContainer` and `ManagedContainerState` in `runtime/ports` because they describe runtime-managed container inventory, then alias them from `instance/ports` for the maintenance use case.
  - Extend `runtime/architecture_test.go` instead of adding a separate test package.
- Owner boundary:
  - `runtime/ports`: container/runtime capabilities only.
  - `instance/ports`: instance access, proxy ticket, teacher query, maintenance, and instance-owned use-case contracts.
  - `app/composition`: adapts runtime container inventory to instance maintenance without hiding owner imports.
- Why this is the narrowest safe surface:
  - It removes the alias surface and adds a guard without moving persistence implementations, changing public contracts, or deleting files.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The decision is about owner boundaries and smallest coherent slice, not a behavior change.
- grill-with-docs findings:
  - Code confirms the stale alias surface is wider than `runtime/ports/http.go`: `runtime/ports/metrics.go` also aliases instance-owned `ManagedContainer` shapes.
  - Docs already describe `runtime/ports/http.go` instance aliases as the remaining runtime -> instance boundary debt.
  - Moving most implementations out of `runtime/infrastructure` would be a larger repository-owner migration and is out of scope for this first alias cleanup.
- Plan adjustments after challenge:
  - Include `ManagedContainer` / `ManagedContainerState` ownership correction in this slice.
  - Keep broad runtime infrastructure owner migration for later slices, but allow focused owner moves when the current task explicitly continues into those blocks.

## Validation

- Commands:
  - Red:
    - `go test ./internal/module/runtime -run TestRuntimePortsDoNotReexportInstancePorts -count=1`
  - Green:
    - `go test ./internal/module/runtime -run TestRuntimePortsDoNotReexportInstancePorts -count=1`
    - `go test ./internal/module/runtime/ports -count=1`
    - `go test ./internal/module/runtime/... -count=1`
    - `go test ./internal/module/instance/... -count=1`
    - `go test ./internal/app -run 'TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestAWDDefenseSSHGateway' -count=1`
    - `go generate ./internal/module/runtime/application/queries`
    - `go test ./internal/module -count=1`
    - `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
    - `python3 scripts/check-docs-consistency.py`
    - `git diff --check`
- Manual checks:
  - `rg -n 'ctf-platform/internal/module/instance/ports' code/backend/internal/module/runtime/ports -g '*.go'`
  - `rg -n 'runtimeports\\.(Instance|ProxyTicket|AWDTargetProxyScope|AWDDefenseSSH|TeacherInstance|UserVisibleInstance|RuntimeCleaner)' code/backend/internal -g '*.go'`
- Review focus:
  - Verify runtime ports no longer import or re-export instance ports.
  - Verify instance-owned proxy ticket and teacher query shapes are referenced through `instance/ports`.
  - Verify runtime container inventory shapes remain usable by instance maintenance without behavior changes.

## Task 1: Runtime Ports Guard

**Files:**
- Modify: `code/backend/internal/module/runtime/architecture_test.go`

- [x] **Step 1: Add a failing guard**

Add `TestRuntimePortsDoNotReexportInstancePorts`, scanning non-test Go files in `runtime/ports` and rejecting imports of `ctf-platform/internal/module/instance/ports`.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/module/runtime -run TestRuntimePortsDoNotReexportInstancePorts -count=1
```

Expected: FAIL because `runtime/ports/http.go` and `runtime/ports/metrics.go` still import `instance/ports`.

## Task 2: Remove Runtime -> Instance Port Aliases

**Files:**
- Modify the runtime, instance, app composition, infrastructure, handler, mapper, and test files listed above.

- [x] **Step 1: Move runtime-owned container inventory shapes to runtime ports**

Define `ManagedContainer` and `ManagedContainerState` in `runtime/ports/metrics.go` or the inventory file. Change `instance/ports` to alias those runtime-owned shapes for maintenance.

- [x] **Step 2: Replace instance-owned alias use sites**

Change production and test code from `runtimeports.<instance-owned type>` to `instanceports.<type>` where the type belongs to instance access, proxy tickets, teacher instance query, or maintenance.

- [x] **Step 3: Remove `instance/ports` import from runtime ports**

Delete the alias declarations from `runtime/ports/http.go` and `runtime/ports/metrics.go`; keep runtime-owned contracts such as `ContainerDirectoryEntry`, runtime workspace repositories, and proxy traffic recorder unchanged.

- [x] **Step 4: Verify GREEN**

Run the validation commands listed above and update docs if the active backlog wording changes.

## Task 3: Move Running Instance Count Owner

**Files:**
- Modify: `code/backend/internal/module/instance/ports/ports.go`
- Create: `code/backend/internal/module/instance/infrastructure/repository.go`
- Create: `code/backend/internal/module/instance/infrastructure/repository_test.go`
- Modify: `code/backend/internal/app/composition/runtime_ops_adapter.go`
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify/delete runtime count query service and tests.

- [x] **Step 1: Add instance-owned running count port and repository**

Define `RunningInstanceCountRepository` in `instance/ports` and implement `CountRunningInstances(ctx)` in `instance/infrastructure.Repository` against the `instances` table.

- [x] **Step 2: Adapt ops dashboard query from instance owner**

Keep `opsports.RuntimeQuery` as the dashboard-facing consumer interface, but adapt it from the instance running count repository in app composition.

- [x] **Step 3: Remove runtime count owner**

Remove `runtimeports.CountRunningRepository`, `runtime/application/queries.CountRunningService`, `Module.RuntimeQuery`, and `runtime/infrastructure.Repository.CountRunning`.

## Task 4: Move Proxy Ticket Infrastructure Owner

**Files:**
- Create: `code/backend/internal/module/instance/infrastructure/proxy_ticket_store.go`
- Create: `code/backend/internal/module/instance/infrastructure/awd_target_proxy_repository.go`
- Move tests from `runtime/infrastructure` to `instance/infrastructure`.
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_builder.go`
- Modify: `code/backend/internal/testutil/systemapp/practice_flow.go`
- Modify: `code/backend/internal/module/runtime/architecture_test.go`
- Delete runtime-owned proxy ticket implementation files after call sites move.

- [x] **Step 1: Add runtime infrastructure guard**

Add a failing architecture test that rejects proxy ticket store and AWD proxy/SSH scope reader implementations under `runtime/infrastructure`.

- [x] **Step 2: Move proxy ticket Redis store and scope reader**

Move Redis ticket storage and AWD target / defense SSH scope SQL reads into `instance/infrastructure`, preserving query behavior and public ticket JSON.

- [x] **Step 3: Rewire composition and tests**

Use `instanceinfra.NewRepository` and `instanceinfra.NewProxyTicketStore` for instance proxy ticket service wiring. Keep runtime infrastructure imports only where runtime container capability or runtime state persistence is still the owner.

- [x] **Step 4: Verify**

Run focused instance/runtime/app tests plus workflow checks before commit.

## Task 5: Remove Instance -> Runtime Baseline Entry

**Files:**
- Modify: `code/backend/internal/module/instance/ports/ports.go`
- Create: `code/backend/internal/module/instance/contracts/access_url.go`
- Modify: `code/backend/internal/module/instance/contracts/instance_output.go`
- Modify: `code/backend/internal/module/instance/application/commands/{instance_service.go,maintenance_service.go}`
- Modify: `code/backend/internal/module/instance/application/queries/instance_service.go`
- Modify: `code/backend/internal/module/instance/infrastructure/awd_target_proxy_repository.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/module/architecture_baseline_test.go`

- [x] **Step 1: Remove instance-side runtime import sites**

Define instance-owned `ManagedContainer`, `ManagedContainerState`, `AWDDefenseWorkspace`, and `AWDServiceOperation` port shapes. Move instance access URL rewriting / alias resolution to `instance/contracts` and keep runtime details parsing narrow to the instance access use case.

- [x] **Step 2: Adapt runtime implementations at composition edge**

Map runtime container inventory structs to instance maintenance views in `app/composition`, and wrap the existing runtime repository with an instance-maintenance adapter so runtime persistence rows are not exposed through instance application interfaces.

- [x] **Step 3: Remove stale baseline allowlist**

Delete `instance -> runtime` from `moduleDependencyBaseline`; `TestModuleDependencyBaselineIsCurrent` now protects against reintroducing that dependency unless a new architecture review explicitly changes the baseline.

## Task 6: Remove Unused Runtime Application Instance Mappers

**Files:**
- Delete: `code/backend/internal/module/runtime/application/commands/response_mapper.go`
- Delete: `code/backend/internal/module/runtime/application/commands/response_mapper_assign.go`
- Delete: `code/backend/internal/module/runtime/application/commands/response_mapper_gen.go`
- Delete: `code/backend/internal/module/runtime/application/queries/response_mapper.go`
- Delete: `code/backend/internal/module/runtime/application/queries/response_mapper_assign.go`
- Delete: `code/backend/internal/module/runtime/application/queries/response_mapper_gen.go`

- [x] **Step 1: Confirm the mapper island has no caller**

`runtimeResponseMapper`, `instanceResponseMapperImpl`, `ToInstanceResp`, `ToInstanceInfo`, and `CopyTime` only resolved inside the six mapper files and had no runtime application caller.

- [x] **Step 2: Delete the unused mapper island**

Remove the command/query response mapper declarations, build-tag assignment files, and generated mapper files so runtime application no longer carries this instance contract import noise.

- [x] **Step 3: Verify**

Run runtime/module architecture tests and workflow checks before committing.

## Remaining Runtime -> Instance Dependency Orchestration

The remaining production `runtime -> instance` edge should be reduced in small owner-preserving slices. Do not replace the dependency with aliases; each slice should move the owner or add a narrow composition adapter.

Current remaining production import sites:

- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/domain/resources.go`
- `code/backend/internal/module/runtime/api/http/{handler.go,access_response_types.go,teacher_instance_types.go,teacher_instance_mapper.go}`
- `code/backend/internal/module/runtime/infrastructure/repository.go`

Target sequencing:

1. Runtime cleanup view first.
   - Runtime cleanup only needs `InstanceID`, optional `NodeID`, fallback `ContainerID`, fallback `NetworkID`, fallback `HostPort`, and encoded `RuntimeDetails`.
   - Add a runtime-owned cleanup target for `RuntimeCleanupService` and `runtime/domain` resource extraction.
   - Keep instance/practice/contest-facing ports unchanged initially, and convert from instance records to cleanup targets in `app/composition`.
   - Completion criteria: runtime application commands and runtime domain no longer import `instance/contracts`.
2. Instance persistence owner second.
   - Move instance table reads/writes out of `runtime/infrastructure.Repository` into `instance/infrastructure.Repository`.
   - Keep runtime allocation persistence (`PortAllocation`, `NetworkAllocation`, runtime nodes, runtime workspaces) in runtime infrastructure.
   - Use a narrow composition adapter where one use case must update instance state and release runtime allocations together.
   - Completion criteria: instance command/query/maintenance services are wired primarily from `instanceinfra.Repository`; runtime infrastructure no longer implements broad instance query/command ports.
3. HTTP compatibility surface third.
   - `runtime/api/http` currently handles instance HTTP routes and returns instance DTOs.
   - After application and persistence owner cleanup, either migrate this handler to `instance/api/http` or rename/recompose it as an instance handler while preserving routes and response contracts.
   - Completion criteria: runtime HTTP no longer imports instance contracts for instance route DTOs, or the file physically moves to the instance module with router compatibility preserved.
4. Baseline cleanup last.
   - Delete `runtime -> instance` from `moduleDependencyBaseline` only after production imports are gone.
   - Keep `TestModuleDependencyBaselineIsCurrent` as the enforcement gate.

## Task 7: Introduce Runtime Cleanup Target

**Files:**
- Modify: `code/backend/internal/module/runtime/architecture_test.go`
- Create: `code/backend/internal/module/runtime/contracts/cleanup_target.go`
- Modify: `code/backend/internal/module/runtime/domain/resources.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
- Modify: `code/backend/internal/app/composition/instance_practice_runtime_adapter.go`
- Modify: `code/backend/internal/app/composition/runtime_challenge_adapter.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Update focused tests that construct cleanup payloads.

- [x] **Step 1: Add failing architecture guard**

Add a focused guard that rejects `instance/contracts` imports from:

- `runtime/application/commands/runtime_cleanup_service.go`
- `runtime/domain/resources.go`

Run:

```bash
go test ./internal/module/runtime -run TestRuntimeCleanupCoreDoesNotDependOnInstanceContracts -count=1
```

Expected: FAIL because both files currently import `ctf-platform/internal/module/instance/contracts`.

- [x] **Step 2: Add runtime-owned cleanup target**

Define a small cleanup target under runtime-owned contracts:

```go
type RuntimeCleanupTarget struct {
    InstanceID     int64
    NodeID         *int64
    ContainerID    string
    NetworkID      string
    HostPort       int
    RuntimeDetails string
}
```

Also add a helper only where useful, for example `NewRuntimeCleanupTarget(...)` at the composition edge, not in runtime core.

- [x] **Step 3: Move runtime core to cleanup target**

Change:

- `RuntimeCleanupService.CleanupRuntime(ctx, target runtimeports.RuntimeCleanupTarget)`
- `runtimedomain.ExtractManagedResources(target runtimeports.RuntimeCleanupTarget)`

Use `target.InstanceID` for port/subnet release and logging. Use `target.RuntimeDetails` plus fallback fields for resource extraction.

- [x] **Step 4: Adapt composition edges**

Keep existing instance/practice/contest-facing ports stable for this slice. Add conversion adapters in `app/composition` so callers that still pass `*instancecontracts.Instance` do not force runtime core to import instance contracts.

- [x] **Step 5: Verify focused runtime cleanup tests**

Run:

```bash
go test ./internal/module/runtime -run 'Test.*Cleanup|TestRuntimeCleanupCoreDoesNotDependOnInstanceContracts' -count=1
go test ./internal/app/composition -run 'Test.*Cleanup|Test.*RuntimeNode' -count=1
```

Expected: PASS.

- [x] **Step 6: Verify module boundary**

Run:

```bash
go test ./internal/module/runtime/... -count=1
go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1
```

Expected: PASS.

## Task 8: Move Remaining Instance Persistence Owner

**Files:**
- Modify: `code/backend/internal/module/instance/infrastructure/repository.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Create: `code/backend/internal/app/composition/instance_runtime_lifecycle_tx.go`
- Modify focused repository and composition tests.

- [x] **Step 1: List runtime repository methods by owner**

Classify each `runtime/infrastructure.Repository` method as:

- instance table query/write
- runtime allocation/state persistence
- mixed instance state plus runtime allocation transaction

- [x] **Step 2: Move pure instance methods**

Move pure instance methods to `instance/infrastructure.Repository`, preserving SQL behavior and tests.

- [x] **Step 3: Adapt mixed methods at composition**

For mixed methods, compose `instanceinfra.Repository` and a narrow runtime allocation repository in `app/composition` rather than making either module own both concepts.

- [x] **Step 4: Verify owner split**

Run focused instance/runtime repository tests and module architecture guards before committing.

Current residual surface after Task 8:

- Pure `instances` table methods have moved to `instance/infrastructure.Repository`, including `FindByUserAndChallenge`, `RefreshInstanceExpiry`, `UpdateRuntime`, `PersistProvisionedRuntime`, `FindExpired`, `ListRecoverableActiveInstances`, `ListStoppingInstances`, `RefreshActiveAWDInstanceExpiryByContest`, `RequeueLostRuntime`, `ListPendingInstances`, `TryTransitionStatus`, and `CountInstancesByStatus`.
- `app/composition/instance_module.go`, `app/composition/contest_module.go`, and `app/composition/instance_runtime_lifecycle_tx.go` now own the cross-owner DB transaction orchestration for `FailProvisioning`, `UpdateStatusAndReleasePort`, `FinalizeStoppedRuntime`, and `ExpireInstanceRuntime`; instance state mutation stays in `instanceinfra.Repository`, while allocation release now goes through `runtimeinfra.AllocationRepository`.
- `runtime/infrastructure.Repository` now remains for runtime-owned allocations, AWD workspace / operation persistence, and active container inventory support; it no longer owns mixed instance-state + runtime-allocation lifecycle transactions.

## Task 9: Re-home Instance HTTP Surface

**Files:**
- Review: `code/backend/internal/module/runtime/api/http/*`
- Review: router wiring under `code/backend/internal/app`

- [x] **Step 1: Decide physical owner**

After Task 7 and Task 8, choose whether to move instance route handlers to `instance/api/http` or keep a compatibility wrapper in runtime while the implementation lives under instance.

- [x] **Step 2: Preserve route and response contracts**

Move or recompose without changing public HTTP paths, JSON fields, proxy ticket behavior, or audit logging.

- [x] **Step 3: Remove final baseline**

Delete `runtime -> instance` from `moduleDependencyBaseline` only after production imports are gone and module guard tests pass.

Current remaining surface after Task 9:

- Production `InstanceModule` handler ownership and HTTP tests have moved to `instance/api/http`; runtime test fixtures and system DTO decode paths now point at the instance module.
- `runtime/infrastructure.Repository` no longer imports `instance/contracts`; remaining cross-owner lifecycle consistency is now orchestrated at the composition edge instead of staying inside runtime persistence.

## Task 10: Move Pure Runtime Shapes To Runtime Contracts

**Files:**
- Create:
  - `code/backend/internal/module/runtime/contracts/container_directory.go`
  - `code/backend/internal/module/runtime/contracts/container_inventory.go`
  - `code/backend/internal/module/runtime/contracts/runtime_node.go`
  - `code/backend/internal/module/runtime/contracts/topology_create.go`
- Modify:
  - `code/backend/internal/module/runtime/ports/container_file_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_inventory_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_stats_runtime.go`
  - `code/backend/internal/module/runtime/ports/node.go`
  - `code/backend/internal/module/runtime/ports/http.go`
  - `code/backend/internal/module/runtime/application/contracts.go`
  - `code/backend/internal/app/composition/*`
  - `code/backend/internal/module/practice/*`
  - `code/backend/internal/module/runtime/infrastructure/*`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Delete:
  - `code/backend/internal/module/runtime/ports/metrics.go`
  - `code/backend/internal/module/runtime/ports/topology.go`

- [x] **Step 1: Move pure shapes out of capability ports**

Move topology create request/result, managed container inventory/stat/state, container directory entry, and runtime node bootstrap/binding shapes from `runtime/ports` to `runtime/contracts`.

- [x] **Step 2: Keep `runtime/ports` focused on capability interfaces**

Update container file / inventory / stats / node ports to reference `runtime/contracts` for shapes while keeping capability interface ownership in `runtime/ports`.

- [x] **Step 3: Rewire all consumers without alias shims**

Change composition adapters, runtime infrastructure, practice ports, runtime agent protocol, and related tests to use `runtimecontracts.*` directly for the moved shapes. Do not add compatibility aliases in `runtime/ports`.

- [x] **Step 4: Verify**

Run:

```bash
go generate ./internal/module/practice/application/queries
go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestAWDDefenseSSHGateway' -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/module/practice/... -count=1
go test ./internal/app/composition -count=1
go test ./internal/module -count=1
go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1
python3 scripts/check-docs-consistency.py
bash scripts/check-code-changes.sh
git diff --check
```

Validation note:

- `go test ./internal/module/practice/... -count=1` first hit a pre-existing-looking readiness probe flake at `TestInstanceReadinessProbeFallsBackToTCPForMalformedHTTPStatusCode`; the focused rerun on `./internal/module/practice/infrastructure -run TestInstanceReadinessProbeFallsBackToTCPForMalformedHTTPStatusCode -count=5` passed, and the full `practice/...` rerun then passed cleanly.

Current remaining surface after Task 10:

- `runtime/ports` 已经只保留 capability interface 与错误哨兵；纯数据形状不再要求 consumer 依赖 capability port 包。
- `container_runtime` 的剩余问题已经收敛成更明确的一条：底层 capability interface 与 Docker / runtime-agent 实现仍然物理落在 `runtime` 模块，后续需要继续判断是否拆到独立 `container_runtime` / platform adapter owner。

## Task 11: Move Runtime Persistence Construction To Composition

**Files:**
- Modify:
  - `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
  - `code/backend/internal/module/runtime/runtime/module.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] **Step 1: Add failing guards**

Require:

- `runtime/runtime.Module` to declare injected runtime persistence dependencies instead of constructing `runtimeinfra.Repository`.
- `app/composition/runtime_module.go` to inject those runtime-owned persistence adapters explicitly.

Run:

```bash
go test ./internal/app -run 'TestRuntimeModuleDoesNotConstructRuntimeInfrastructure|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule' -count=1
```

Expected: FAIL before the slice because `runtime/runtime/module.go` still imported `runtimeinfra` and `gorm/redis` and composition did not inject runtime persistence adapters.

- [x] **Step 2: Export narrow command dependency interfaces**

Export `runtimecmd.ProvisioningRepository` and `runtimecmd.RuntimeCleanupRepository` from the commands package so the runtime module can depend on named narrow persistence ports instead of a concrete repository constructor.

- [x] **Step 3: Remove runtime module infrastructure construction**

Change `runtimemodule.Deps` to accept injected runtime-owned persistence adapters and remove `DB` / `Cache` / `runtimeinfra.NewRepository(...)` from `runtime/runtime/module.go`.

- [x] **Step 4: Move injection to composition**

Make `internal/app/composition/runtime_module.go` own the `runtimeinfra.NewRepository(root.DB())` construction and pass it into `runtimemodule.Build(...)` as the provisioning / cleanup persistence adapter.

- [x] **Step 5: Verify**

Run:

```bash
go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestRuntimeModuleDoesNotConstructRuntimeInfrastructure|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule|TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestAWDDefenseSSHGateway' -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/app/composition -count=1
go test ./internal/module -count=1
```

Current remaining surface after Task 11:

- `runtime/runtime.Module` 已经不再直接依赖 `runtime/infrastructure` 的具体构造，也不再把 DB / cache 作为自己的输入；container-facing service builder 与 runtime-owned persistence 构造边界已经分开。
- 剩余 debt 主要是 capability interface、host adapter 和 `ContainerRuntimeModule` 这组物理包落点仍留在 `runtime`，以及 `runtime/infrastructure.Repository` 内部仍有待继续拆分的 runtime-owned persistence 面。

## Task 12: Narrow Composition Runtime Repo Dependencies

**Files:**
- Modify:
  - `code/backend/internal/app/composition/runtime_acl_migration.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_e2e_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] **Step 1: Add failing guards**

Require `runtime_acl_migration.go` and `runtime_node_execution_router.go` to stop depending on `*runtimeinfra.Repository` directly.

Run:

```bash
go test ./internal/app -run 'TestRuntimeNodeExecutionRouterUsesNarrowRuntimePersistenceDeps|TestRuntimeACLMigrationUsesNarrowRuntimeStateDeps' -count=1
```

Expected: FAIL before the slice because both files still referenced `*runtimeinfra.Repository`.

- [x] **Step 2: Introduce narrow composition-side interfaces**

Define:

- `runtimeACLMigrationRepository`
- `runtimeNodeAllocationRepository`
- `runtimeNodeStateRepository`

Keep them local to `app/composition`, because this slice is about narrowing the composition wiring surface, not creating a new shared port package yet.

- [x] **Step 3: Rewire router and ACL migration**

Change runtime node router and ACL migration helpers to accept the narrow interfaces. Keep `runtimeinfra.NewRepository(root.DB())` as the concrete provider at composition entry, but stop threading `*runtimeinfra.Repository` through the container runtime orchestration path.

- [x] **Step 4: Verify**

Run:

```bash
go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestRuntimeModuleDoesNotConstructRuntimeInfrastructure|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule|TestRuntimeNodeExecutionRouterUsesNarrowRuntimePersistenceDeps|TestRuntimeACLMigrationUsesNarrowRuntimeStateDeps|TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestAWDDefenseSSHGateway' -count=1
go test ./internal/app/composition -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/module -count=1
```

Current remaining surface after Task 12:

- 容器能力装配链已经不再把 `*runtimeinfra.Repository` 作为 concrete 类型向下传递；composition 面上已经明确分成 module 注入、router allocation/state、ACL migration state 三类窄依赖。

## Task 13: Split Runtime Allocation Repository

**Files:**
- Create:
  - `code/backend/internal/module/runtime/infrastructure/allocation_repository.go`
- Modify:
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/contest_module.go`
  - `code/backend/internal/app/composition/instance_runtime_lifecycle_tx.go`
  - `code/backend/internal/app/composition/practice_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_e2e_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - practice/runtime/contest test adapters that release runtime allocations or construct a practice runtime port owner.
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] **Step 1: Add composition guard for allocation owner**

Extend typed dependency tests so `runtimeinfra.AllocationRepository` owns port/subnet reservation and allocation release methods, and broad `runtimeinfra.Repository` no longer declares them.

- [x] **Step 2: Move allocation persistence methods**

Move `ReserveAvailablePort*`, `ReserveAvailableSubnet*`, `SyncInstanceHostPortForRestart`, `ReleaseRuntimeAllocationsForInstance`, release helpers, and allocation conflict helpers from `runtime/infrastructure.Repository` into `runtime/infrastructure.AllocationRepository`.

- [x] **Step 3: Rewire composition and test adapters**

Use `runtimeinfra.NewAllocationRepository(root.DB())` for provisioning, cleanup, node router allocation, practice runtime port owner, and mixed lifecycle transaction release paths. Keep `runtimeinfra.Repository` only for AWD workspace / operation persistence, runtime state index lookups, runtime node state, and migration support.

- [x] **Step 4: Verify**

Run:

```bash
go test ./internal/app -run 'TestPracticeModuleWiresRuntimePortOwnerFromCompositionRoot|TestRuntimeRepositoryDoesNotOwnAllocationPersistence' -count=1
go test ./internal/module/runtime/infrastructure -count=1
go test ./internal/app/composition -count=1
go test ./internal/module/practice/... -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/module -count=1
```

Current remaining surface after Task 13:

- `runtime/infrastructure.Repository` no longer owns allocation persistence; port/subnet reservation and allocation release now have a concrete owner in `runtime/infrastructure.AllocationRepository`.
- `runtime/infrastructure.Repository` still groups AWD defense workspace / AWD service operation persistence with runtime state index and migration-facing state lookups. The next runtime persistence cleanup should split those remaining owner groups instead of adding methods back to the broad repository.
- `container_runtime` capability interface / host adapter / `ContainerRuntimeModule` physical owner is still undecided and remains a separate structure question.

## Task 14: Split Runtime AWD Persistence Repository

**Files:**
- Create:
  - `code/backend/internal/module/runtime/infrastructure/awd_repository.go`
- Modify:
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/infrastructure/awd_defense_workspace_repository_test.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/contest_module.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - runtime / practice / contest test adapters that read AWD workspace or finish AWD service operations.
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] **Step 1: Add AWD persistence owner guard**

Extend typed dependency tests so `runtimeinfra.AWDRepository` owns AWD defense workspace and AWD service operation persistence, and broad `runtimeinfra.Repository` no longer declares those methods.

- [x] **Step 2: Move AWD workspace / operation methods**

Move `FindAWDDefenseWorkspace`, `UpsertAWDDefenseWorkspace`, `BumpAWDDefenseWorkspaceRevision`, `FindRunningAWDDefenseWorkspaceByInstanceID`, `CreateAWDServiceOperation`, `FinishActiveAWDServiceOperationForInstance`, and `FinishAWDServiceOperation` from `runtime/infrastructure.Repository` into `runtime/infrastructure.AWDRepository`.

- [x] **Step 3: Rewire composition and test adapters**

Use `runtimeinfra.NewAWDRepository(root.DB())` for instance maintenance AWD workspace / operation paths, practice active AWD operation finish, and contest ended-runtime workspace state store. Keep `runtimeinfra.Repository` for active container inventory, container -> node state lookup, ACL migration state, proxy traffic recorder, and other state/index reads not moved in this slice.

- [x] **Step 4: Verify**

Run:

```bash
go test ./internal/app -run 'TestRuntimeRepositoryDoesNotOwnAWDPersistence|TestRuntimeRepositoryDoesNotOwnAllocationPersistence' -count=1
go test ./internal/module/runtime/infrastructure -count=1
go test ./internal/app/composition -count=1
go test ./internal/module/practice/... -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/module/contest/infrastructure -count=1
```

Current remaining surface after Task 14:

- `runtime/infrastructure.Repository` no longer owns allocation persistence or AWD workspace / operation persistence.
- Remaining `runtime/infrastructure.Repository` responsibilities are active container inventory, container-to-node state lookup, ACL migration state update, runtime managed instance lookup, and proxy traffic event recorder support. These should be split in a later state/index slice instead of using the broad repository as a default landing zone.
- `container_runtime` capability interface / host adapter / `ContainerRuntimeModule` physical owner is still undecided and remains a separate structure question.
- 下一步剩余 debt 已进一步收敛成 concrete runtime persistence 本身的继续拆分，以及 capability interface / host adapter / `ContainerRuntimeModule` 的最终物理 owner 迁移。

## Task 15: Split Runtime State Repository

**Files:**
- Create:
  - `code/backend/internal/module/runtime/infrastructure/managed_instance_repository.go`
  - `code/backend/internal/module/runtime/infrastructure/active_container_inventory_repository.go`
  - `code/backend/internal/module/runtime/infrastructure/container_node_index_repository.go`
  - `code/backend/internal/module/runtime/infrastructure/acl_migration_state_repository.go`
- Modify:
  - `code/backend/internal/module/runtime/infrastructure/proxy_traffic_recorder.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_e2e_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `code/backend/internal/module/runtime/service_test.go`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Delete:
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/infrastructure/state_repository.go`

- [x] **Step 1: Add runtime state owner guard**

Extend typed dependency tests so runtime managed instance lookup, active container inventory, container-to-node state lookup, and ACL migration state persistence each have their own concrete repository owner. Other runtime infrastructure files should not continue to expose those methods.

- [x] **Step 2: Move runtime state/index methods**

Move `FindByID`, `ListActiveContainerIDs`, `FindRuntimeNodeIDByContainerID`, `ListInstancesNeedingACLHandleMigration`, and `UpdateInstanceRuntimeDetails` out of the transitional `RuntimeStateRepository`, and split them into `ManagedInstanceRepository`, `ActiveContainerInventoryRepository`, `ContainerNodeIndexRepository`, and `ACLMigrationStateRepository`.

- [x] **Step 3: Rewire composition and recorder owner**

Use `runtimeinfra.NewContainerNodeIndexRepository(root.DB())` for runtime node router lookup, `runtimeinfra.NewACLMigrationStateRepository(root.DB())` for legacy ACL migration state, and `runtimeinfra.NewActiveContainerInventoryRepository(root.DB())` for instance maintenance active container inventory reads. At the same time, let proxy traffic recording own its own concrete `ProxyTrafficEventRecorder` instead of hanging off the retired broad repository.

- [x] **Step 4: Verify**

Run:

```bash
go test ./internal/app -run 'TestRuntimeRepositoryDoesNotOwnStatePersistence|TestRuntimeStatePersistenceWiredFromCompositionRoot|TestRuntimeRepositoryDoesNotOwnAWDPersistence|TestRuntimeRepositoryDoesNotOwnAllocationPersistence|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule|TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildInstanceModuleDelegatesToSubBuilders' -count=1
go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter.*' -count=1
go test ./internal/module/runtime/... -count=1
```

Current remaining surface after Task 15:

- `runtime/infrastructure.Repository` 与过渡态 `RuntimeStateRepository` 都已删除；allocation、AWD、managed instance、active container inventory、container-node index、ACL migration、proxy traffic recorder 都已经有各自 concrete owner。
- runtime 侧剩余的结构问题主要从“宽仓储继续拆”收敛成“`runtime/ports/*` capability interface、host adapter、`ContainerRuntimeModule` 的最终物理 owner 是否继续留在 runtime，还是迁到独立 `container_runtime` / platform adapter”。
- 当前只剩 `ManagedInstanceRepository` 是否应继续保留为 production owner 需要观察；如果后续只剩测试或极窄读面，应继续下沉或消除，而不是再回并成新的宽 state repo。
