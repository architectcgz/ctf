<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 拆分 runtime container capability port 文件 Implementation Plan

**Goal:** Split `runtime/ports/container_runtime.go` into capability-specific port files without changing exported interface names, behavior, or wiring.

**Architecture:** Keep `runtime` as the physical owner of Docker / runtime-agent host execution. This slice only improves the physical landing of existing capability ports by grouping them into smaller files; `RuntimeHostExecutor` remains the infrastructure aggregation boundary for local Docker and remote runtime-agent adapters.

**Tech Stack:** Go, existing CTF modular monolith backend, runtime ports / infrastructure / app composition guard tests.

---

## Task Metadata

- Task Slug: `2026-06-09-runtime-container-runtime-port-file-split`
- Started At: `2026-06-09T04:08:23Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-runtime-container-runtime-port-file-split`
- Branch: `task/2026-06-09-runtime-container-runtime-port-file-split`

## Objective And Non-Goals

- Objective:
  - Split provisioning, cleanup, file, image, inventory, stats, interactive, and host-executor port definitions into dedicated files under `code/backend/internal/module/runtime/ports/`.
  - Add an architecture guard that keeps `RuntimeHostExecutor` usage limited to the runtime host adapter and app composition boundary.
  - Preserve all exported type names and method signatures.
  - Keep all production wiring and tests behaviorally unchanged.
- Non-Goals:
  - Do not move runtime host execution out of `runtime`.
  - Do not migrate ops running-count ownership to `instance`.
  - Do not remove `runtime/ports/http.go` instance aliases.
  - Do not change runtime-agent protocol, Docker engine implementation, or app composition fields.

## Inputs

- Source docs:
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Related architecture/contracts:
  - `code/backend/internal/module/runtime/ports/container_runtime.go`
  - `code/backend/internal/module/runtime/runtime/module.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/module/runtime/infrastructure/engine.go`
  - `code/backend/internal/module/runtime/infrastructure/agentclient/bridge.go`
  - `code/backend/internal/module/runtime/infrastructure/agentserver/service.go`
- Related prior work:
  - `48a752359 refactor(runtime): 收口端口保留契约边界`
  - `fab062129 feat(challenge): 增加题包交付与文件 GC`

## Task Classification

- Classification: `非琐碎任务`
- Why: This deletes and creates Go files in a known runtime boundary debt surface, even though the intended behavior is unchanged.

## Files

- Create:
  - `code/backend/internal/module/runtime/ports/provisioning_runtime.go`
  - `code/backend/internal/module/runtime/ports/cleanup_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_file_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_image_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_inventory_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_stats_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_interactive_runtime.go`
  - `code/backend/internal/module/runtime/ports/runtime_host_executor.go`
- Modify:
  - `code/backend/internal/module/architecture_test.go`
  - If comments or tests reveal stale references, keep edits scoped to `runtime/ports` and architecture checks.
- Review:
  - `code/backend/internal/module/runtime/ports/http.go`
  - `code/backend/internal/module/runtime/ports/metrics.go`
  - `code/backend/internal/module/runtime/application/contracts.go`
  - `code/backend/internal/app/composition/runtime_module.go`
- Test:
  - `go test ./internal/module/runtime/ports -count=1`
  - `go test ./internal/module/runtime/... -count=1`
  - `go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestBuildContainerRuntimeModuleDelegatesToSubBuilders' -count=1`
  - `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`
  - `go test ./internal/module -run TestRuntimeHostExecutorUsageIsRestricted -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - `runtime/ports` already has smaller files such as `node.go`, `topology.go`, `metrics.go`, `port_reservation.go`, and `http.go`.
  - `runtime/runtime.Module` already depends on individual capability interfaces rather than a public `Engine`.
  - Local Docker `Engine`, runtime-agent `Bridge`, and agent server still need a single host-executor aggregation boundary.
- Reuse / extend / split / create-new decision:
  - Split the existing file into multiple files in the same package; do not create new packages or rename exported interfaces.
  - Keep `RuntimeHostExecutor` as the only explicit wide aggregate because it represents the host execution adapter boundary, not a business-module dependency.
- Owner boundary:
  - `runtime` remains the physical owner of host execution.
  - `challenge`, `contest`, `ops`, `instance`, and `practice` continue to consume their existing app-composition adapters and consumer-side ports.
- Why this is the narrowest safe surface:
  - File-level split improves the physical landing of capability definitions while avoiding signature, import path, and wiring churn.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The user requested the first optimization slice after an architecture analysis; the decision space is about scope control rather than new behavior.
- grill-with-docs findings:
  - Existing docs still classify this as an active runtime boundary debt.
  - Code confirms `ContainerRuntimeModule` already exposes consumer-facing fields; the current slice should not move owners yet.
  - `RuntimeHostExecutor` remains necessary for the local Docker / remote runtime-agent bridge surface.
- Plan adjustments after challenge:
  - Limit this slice to physical file split.
  - Leave `RuntimeHostExecutor` guard, ops count migration, and runtime->instance alias cleanup as later slices.

## Validation

- Commands:
  - `go test ./internal/module/runtime/ports -count=1`
  - `go test ./internal/module/runtime/... -count=1`
  - `go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestBuildContainerRuntimeModuleDelegatesToSubBuilders' -count=1`
  - `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`
  - `git diff --check`
- Manual checks:
  - `rg -n "container_runtime.go|RuntimeHostExecutor|ContainerProvisioningRuntime|ContainerCleanupRuntime" code/backend/internal/module/runtime/ports code/backend/internal/app code/backend/internal/module/runtime -g '*.go'`
- Review focus:
  - Verify exported API names and method signatures are unchanged.
  - Verify no behavior or wiring changes are introduced.
  - Verify file split does not hide the remaining wider owner work.

## Task 1: Split Runtime Capability Port Files

**Files:**
- Create:
  - `code/backend/internal/module/runtime/ports/provisioning_runtime.go`
  - `code/backend/internal/module/runtime/ports/cleanup_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_file_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_image_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_inventory_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_stats_runtime.go`
  - `code/backend/internal/module/runtime/ports/container_interactive_runtime.go`
  - `code/backend/internal/module/runtime/ports/runtime_host_executor.go`
- Remove:
  - `code/backend/internal/module/runtime/ports/container_runtime.go`

- [x] **Step 1: Copy each interface into a capability-specific file**

Preserve comments, names, method signatures, and package name.

- [x] **Step 2: Remove the original aggregate file**

Delete `container_runtime.go` after all definitions are present in the new files.

- [x] **Step 3: Run focused runtime validation**

Run:

```bash
go test ./internal/module/runtime/ports -count=1
go test ./internal/module/runtime/... -count=1
```

- [x] **Step 4: Run app/module architecture guards**

Run:

```bash
go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestBuildContainerRuntimeModuleDelegatesToSubBuilders' -count=1
go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1
git diff --check
```

Validation note:

- First `go test ./internal/module/runtime/... -count=1` run observed an existing flaky duplicate-cleanup assertion in `TestServiceDestroyManagedInstanceMarksStoppingThenBackgroundCleanupRemovesRuntime`; the failing test passed when rerun directly, and the full `runtime/...` command passed on a clean rerun. This slice did not touch runtime cleanup code.

## Task 2: Guard RuntimeHostExecutor Usage

**Files:**
- Modify:
  - `code/backend/internal/module/architecture_test.go`

- [x] **Step 1: Add a production-source guard for `RuntimeHostExecutor` references**

Scan non-test Go files under `code/backend/internal/module` and `code/backend/internal/app`. Allow the wide host executor only in:

- `code/backend/internal/module/runtime/ports/runtime_host_executor.go`
- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/agentclient/bridge.go`
- `code/backend/internal/module/runtime/infrastructure/agentserver/service.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/runtime_node_execution_router.go`

- [x] **Step 2: Run focused architecture validation**

Run:

```bash
go test ./internal/module -run TestRuntimeHostExecutorUsageIsRestricted -count=1
go test ./internal/module -count=1
go test ./internal/app -run 'TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestBuildContainerRuntimeModuleDelegatesToSubBuilders' -count=1
git diff --check
```

- [x] **Step 3: Review guard behavior**

Confirm the guard blocks new production references to the wide host executor outside runtime infrastructure and app composition, while keeping test fakes outside the production scan.
