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
  - `code/backend/internal/module/runtime/application/queries/response_mapper.go`
  - `code/backend/internal/module/runtime/application/queries/response_mapper_gen.go`
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
  - The module dependency baseline already records `runtime -> instance` and `instance -> runtime`; this slice reduces hidden re-export coupling rather than removing the full implementation dependency.
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
