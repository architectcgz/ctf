<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# container_runtime 迁移尾项收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the remaining container execution capability owner from legacy `runtime` implementation packages into `container_runtime`, while leaving instance / contest / practice business ownership out of the bottom module.

**Architecture:** `container_runtime` becomes the owner of container capability contracts, ports, application services, domain resource extraction, local Docker host adapter, runtime-agent protocol, and host-executor adapter packages. Legacy `runtime` remains only as runtime state / persistence owner until those state records have a clearer business owner; moved capability import paths are not kept as compatibility shims.

**Tech Stack:** Go, modular monolith backend, GORM-backed runtime persistence, Docker SDK, runtime-agent gRPC bridge, architecture guard tests.

---

## Task Metadata

- Task Slug: `2026-06-10-container-runtime-tail-migration`
- Started At: `2026-06-10T05:19:39Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-container-runtime-tail-migration`
- Branch: `task/2026-06-10-container-runtime-tail-migration`

## Objective And Non-Goals

- Objective:
  - Move pure container runtime contracts, ports, application services, domain helpers, Docker host adapter, and runtime-agent protocol / bridge code into `internal/module/container_runtime`.
  - Retire old `internal/module/runtime/{application,agentcontracts}` capability implementation paths and keep `runtime/{contracts,ports,infrastructure}` only for runtime persistence/state where still needed.
  - Update app composition and module imports so production container capability assembly no longer depends on `runtime/application`, `runtime/application/commands`, or `runtime/ports`.
  - Update architecture guards, module dependency baseline, architecture docs, and the active migration TODO to reflect the new owner.
- Non-Goals:
  - Do not migrate instance business services, instance access handlers, practice flows, contest AWD workflows, or assessment / ops side effects.
  - Do not perform database schema changes.
  - Do not change Docker behavior, runtime-agent wire JSON fields, HTTP routes, API DTOs, or user-visible behavior.
  - Do not decide the final owner of every runtime persistence record in the same batch if a record still represents mixed instance / AWD state. Keep those concrete repositories in `runtime/infrastructure` unless the owner is already unambiguous.

## Inputs

- Source docs:
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-09-container-runtime-module-boundary-implementation-plan.md`
- Related architecture/contracts:
  - `code/backend/internal/module/container_runtime/architecture_test.go`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/app/architecture_rules_test.go`
- Related prior work:
  - `83b9825cd refactor(runtime): 建立 container_runtime 模块入口`
  - `docs/reviews/architecture/2026-06-10-container-runtime-module-boundary-review.md`

## Task Classification

- Classification: `非琐碎任务 / structural backend migration`
- Why:
  - Moves module ownership and import paths across `container_runtime`, `runtime`, app composition, tests, and architecture docs.
  - Touches a known active structural-debt surface from `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`.
  - Requires architecture guards, explicit no-compat migration decisions, focused Go validation, completion-full, and independent review.

## Files

- Create:
  - `code/backend/internal/module/container_runtime/application/*.go`
  - `code/backend/internal/module/container_runtime/application/commands/*.go`
  - `code/backend/internal/module/container_runtime/contracts/*.go`
  - `code/backend/internal/module/container_runtime/domain/*.go`
  - `code/backend/internal/module/container_runtime/ports/*.go`
  - `code/backend/internal/module/container_runtime/agentcontracts/*.go`
  - `code/backend/internal/module/container_runtime/infrastructure/*.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/*.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/*.go`
- Modify:
  - `code/backend/internal/module/container_runtime/runtime/module.go`
  - `code/backend/internal/module/container_runtime/architecture_test.go`
  - `code/backend/internal/module/runtime/{contracts,ports,application,agentcontracts,runtime}/*.go`
  - `code/backend/internal/module/runtime/domain/*.go`
  - `code/backend/internal/module/runtime/infrastructure/*.go`
  - `code/backend/internal/module/runtime/infrastructure/agentclient/*.go`
  - `code/backend/internal/module/runtime/infrastructure/agentserver/*.go`
  - `code/backend/internal/app/composition/*.go`
  - `code/backend/internal/app/*architecture*_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Review:
  - `code/backend/internal/module/practice/**`
  - `code/backend/internal/module/challenge/**`
  - `code/backend/internal/module/contest/**`
  - `code/backend/tests/**`
- Test:
  - `go test ./internal/module/container_runtime/... -count=1`
  - `go test ./internal/module/runtime/... -count=1`
  - `go test ./internal/app/composition -count=1`
  - `go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries|TestRuntimeHostExecutorUsageIsRestricted' -count=1`
  - `go test ./internal/app -run 'TestArchitectureRules.*|TestRouterCompositionStructure|TestRuntime.*|TestBuildContainerRuntimeModule.*' -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash scripts/check-workflow-complete.sh`

## 复用与 Owner 决策

- Existing patterns searched:
  - `container_runtime/runtime` already owns the module builder but currently imports `runtime/application`, `runtime/application/commands`, and `runtime/ports`.
  - `runtime/contracts` contains two categories: pure container capability shapes and runtime persistence / instance state shapes.
  - `runtime/infrastructure` contains both pure host adapters (`Engine`, runtime-agent bridge/server) and persistence/state repositories (`AllocationRepository`, `ManagedInstanceRepository`, `AWDRepository`, node repository, Redis state stores).
- Reuse / extend / split / create-new decision:
  - Move pure capability definitions and implementations into `container_runtime`.
  - Update production and test imports to the new owner instead of preserving old capability compatibility shims.
  - Leave persistence repositories in `runtime/infrastructure` unless they are purely host-adapter code. This avoids moving mixed instance/AWD state into the bottom container module.
- Owner boundary:
  - `container_runtime`: container execution capability contracts, ports, application services, local Docker host adapter, runtime-agent protocol/bridge/server, host-executor aggregate.
  - `runtime`: remaining runtime persistence/state stores and state contracts whose final owner is not yet resolved.
  - `instance`: instance lifecycle, visibility, access, proxy ticket, maintenance use cases.
  - `contest` / `practice` / `challenge`: business flows that consume container runtime through explicit ports/adapters.
- Why this is the narrowest safe surface:
  - It closes the concrete tail identified by the previous review without forcing a schema or business owner redesign.
  - It deletes the broad old `runtime` capability packages after switching production and test callers to the new owner, avoiding long-lived legacy shims.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - This is a design-boundary migration, not a bug; the main risk is choosing the wrong owner for mixed runtime state.
- Evidence inspected:
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/module/runtime/{contracts,ports,application,infrastructure}`
  - `code/backend/internal/module/container_runtime/{runtime,architecture_test.go}`
  - `code/backend/internal/module/architecture_{test,baseline_test}.go`
- Working design:
  - Move pure capability packages and host adapters to `container_runtime`.
  - Keep persistence/state repositories in `runtime/infrastructure` unless they are host adapter code.
  - Delete old capability import paths instead of preserving compatibility aliases; update all affected production and test callers in this batch.
- grill-with-docs findings:
  - The term `container_runtime` means platform/container execution capability, not instance business lifecycle.
  - `RuntimeManagedInstance` and AWD workspace/service operation records remain mixed runtime persistence views; moving them into `container_runtime` would incorrectly make the bottom module own business state.
  - The runtime-agent bridge also implements `contestports.CheckerRunner`; this is an infrastructure adapter exception that must be explicitly documented/guarded instead of hidden.
- Plan adjustments after challenge:
  - Scope includes capability application / ports / contracts / host adapter movement.
  - Scope excludes final persistence owner migration unless the code is purely host adapter.
  - Add red architecture tests before moving code.

## Validation

- Commands:
  - Red test:
    - `cd code/backend && timeout 120s go test ./internal/module/container_runtime/... -run 'TestContainerRuntimeOwnsCapabilityImplementationPackages|TestRuntimePackageDoesNotDependOnLegacyRuntimeImplementation' -count=1`
  - Focused green tests:
    - `cd code/backend && timeout 180s go test ./internal/module/container_runtime/... -count=1`
    - `cd code/backend && timeout 180s go test ./internal/module/runtime/... -count=1`
    - `cd code/backend && timeout 180s go test ./internal/app/composition -count=1`
    - `cd code/backend && timeout 180s go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries|TestRuntimeHostExecutorUsageIsRestricted' -count=1`
    - `cd code/backend && timeout 180s go test ./internal/app -run 'TestArchitectureRules.*|TestRouterCompositionStructure|TestRuntime.*|TestBuildContainerRuntimeModule.*' -count=1`
  - Completion:
    - `timeout 900s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
    - `timeout 1200s bash scripts/check-workflow-complete.sh`
- Manual checks:
  - `rg -n "ctf-platform/internal/module/runtime/(application|ports|agentcontracts)" code/backend/internal/module/container_runtime code/backend/internal/app -g '*.go'`
  - `rg -n "RuntimeHostExecutor" code/backend/internal/module code/backend/internal/app -g '*.go'`
- Review focus:
  - Verify `container_runtime` does not absorb instance / contest / practice business owner logic.
  - Verify old `runtime` capability implementation paths are retired instead of containing duplicate behavior.
  - Verify runtime-agent JSON protocol and Docker behavior are unchanged.

## Tasks

### Task 1: Add Container Runtime Ownership Guard

**Files:**
- Modify: `code/backend/internal/module/container_runtime/architecture_test.go`

- [x] **Step 1: Write red architecture tests**

Require `container_runtime` to own `application`, `application/commands`, `contracts`, `domain`, `infrastructure`, and `ports`, and require `container_runtime/runtime` to stop importing legacy `runtime/application` and `runtime/ports`.

- [x] **Step 2: Run red test**

Run:

```bash
cd code/backend && timeout 120s go test ./internal/module/container_runtime/... -run 'TestContainerRuntimeOwnsCapabilityImplementationPackages|TestRuntimePackageDoesNotDependOnLegacyRuntimeImplementation' -count=1
```

Expected: FAIL because `container_runtime/application` does not exist and `runtime/module.go` still imports old implementation packages.

### Task 2: Move Capability Contracts, Ports, Application, And Domain

**Files:**
- Move/create:
  - `runtime/contracts/*.go` capability shapes to `container_runtime/contracts/*.go`
  - `runtime/ports/*.go` capability ports to `container_runtime/ports/*.go`
  - `runtime/application/{container_file_service,container_stats_service,image_runtime_service,context,contracts,dependency_helpers}.go` to `container_runtime/application/`
  - `runtime/application/commands/{context,dependency_helpers,provisioning_service,runtime_cleanup_service}.go` to `container_runtime/application/commands/`
  - `runtime/domain/{resources,topology_acl}.go` to `container_runtime/domain/`
- Modify:
  - app / module imports for touched production code.

- [x] **Step 1: Move files with `git mv`**

Move pure capability files to `container_runtime` package directories.

- [x] **Step 2: Update package names and imports**

Change moved package names to their new package paths and import `container_runtime/contracts|ports|domain`.

- [x] **Step 3: Retire old runtime capability packages**

Move remaining behavior tests to `container_runtime` or the real business owner, and remove old capability wrapper package usage instead of adding aliases.

- [x] **Step 4: Run focused package tests**

Run:

```bash
cd code/backend && timeout 180s go test ./internal/module/container_runtime/... ./internal/module/runtime/... -count=1
```

Expected: PASS.

### Task 3: Move Host Adapter And Runtime-Agent Protocol

**Files:**
- Move/create:
  - `runtime/agentcontracts/*` to `container_runtime/agentcontracts/`
  - pure host adapter files from `runtime/infrastructure` to `container_runtime/infrastructure`
  - `runtime/infrastructure/agentclient/*` to `container_runtime/infrastructure/agentclient/`
  - `runtime/infrastructure/agentserver/*` to `container_runtime/infrastructure/agentserver/`
- Modify:
  - app composition host executor construction.
  - architecture baseline and RuntimeHostExecutor usage guard.

- [x] **Step 1: Move host adapter files**

Move local Docker engine, metrics, file, inventory, provisioning, ACL, engine error, runtime-agent client/server, and agent contract files into `container_runtime`.

- [x] **Step 2: Update production imports**

Switch `app/composition/runtime_module.go`, runtime node client wiring, and tests that exercise host adapter behavior to `container_runtime/infrastructure`.

- [x] **Step 3: Keep runtime persistence owners in place**

Do not move `AllocationRepository`, `ManagedInstanceRepository`, `AWDRepository`, node repository, Redis state stores, or proxy traffic recorder in this task unless a test proves they are pure host adapter code.

- [x] **Step 4: Run host adapter tests**

Run:

```bash
cd code/backend && timeout 180s go test ./internal/module/container_runtime/infrastructure/... ./internal/module/runtime/infrastructure -count=1
cd code/backend && timeout 180s go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter|TestBuildRuntimeHostExecutor|TestAWDDefenseSSHGateway' -count=1
```

Expected: PASS.

### Task 4: Update Guards, Docs, And TODO State

**Files:**
- Modify:
  - `code/backend/internal/module/container_runtime/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/app/architecture_rules_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] **Step 1: Update architecture guardrails**

Remove stale `container_runtime -> runtime/application|ports` exceptions and keep only explicit, reviewed persistence/state exceptions that remain necessary.

- [x] **Step 2: Update architecture docs**

Record `container_runtime` as owner of capability contracts / ports / application / host adapter. Record `runtime` as persistence/state owner only.

- [x] **Step 3: Update TODO**

If only persistence state remains, rewrite the active item to no longer claim capability contracts / ports / infrastructure are still in old `runtime`; keep only the precise remaining persistence owner item.

- [x] **Step 4: Run docs and architecture validation**

Run:

```bash
cd code/backend && timeout 180s go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries|TestRuntimeHostExecutorUsageIsRestricted' -count=1
cd code/backend && timeout 180s go test ./internal/app -run 'TestArchitectureRules.*|TestRouterCompositionStructure' -count=1
timeout 120s git diff --check
```

Expected: PASS.

### Task 5: Completion Validation And Independent Review

**Files:**
- Review all changed files.
- Create: `docs/reviews/architecture/2026-06-10-container-runtime-tail-migration-review.md`

- [x] **Step 1: Run completion validation**

Run:

```bash
timeout 900s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
```

Expected: PASS.

- [x] **Step 2: Run independent architecture review**

Archive the review under:

```text
docs/reviews/architecture/2026-06-10-container-runtime-tail-migration-review.md
```

- [x] **Step 3: Fix material findings and rerun impacted validation**

Result: same-context review found no material findings. Independent reviewer gate is explicitly recorded as unmet because no separate reviewer tool is available in this session.

- [x] **Step 4: Run workflow completion gate**

Run:

```bash
timeout 1200s bash scripts/check-workflow-complete.sh
```

Result: PASS.
