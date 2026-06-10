<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# container_runtime 底层模块边界收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish `container_runtime` as the physical bottom-module entry for container execution capabilities, without moving instance or contest business ownership into it.

**Architecture:** First introduce a narrow `internal/module/container_runtime/runtime` builder that owns the container capability module output used by app composition. Keep existing `internal/module/runtime/{application,contracts,ports,infrastructure}` as the legacy implementation substrate for this slice, and add guardrails so the new module does not import `contest`, `practice`, or `instance` business ports.

**Tech Stack:** Go, Gin app composition, GORM-backed runtime repositories, Docker/runtime-agent adapters, project architecture tests.

---

## Task Metadata

- Task Slug: `2026-06-09-container-runtime-module-boundary`
- Started At: `2026-06-09T15:32:03Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-container-runtime-module-boundary`
- Branch: `task/2026-06-09-container-runtime-module-boundary`

## Objective And Non-Goals

- Objective:
  - Create a physical `container_runtime` bottom module entry for container runtime capability wiring.
  - Switch `internal/app/composition` from `runtime/runtime.Build` to `container_runtime/runtime.Build`.
  - Keep `container_runtime` free of business owner dependencies such as `contest/ports`, `practice/ports`, and `instance/ports`.
  - Update architecture docs / TODO facts for this first boundary slice.
- Non-Goals:
  - Do not move every `runtime/contracts`, `runtime/ports`, `runtime/infrastructure`, or `runtime/application` file in this slice.
  - Do not change external HTTP routes, database schema, API DTOs, Docker behavior, runtime-agent protocol, or instance lifecycle behavior.
  - Do not delete the legacy `internal/module/runtime/runtime` package until callers and tests no longer need the compatibility surface.

## Inputs

- Source docs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Related architecture/contracts:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/runtime/architecture_test.go`
  - `code/backend/internal/app/composition/architecture_test.go`
- Related prior work:
  - Existing `ContainerRuntimeModule` app composition view.
  - Existing split where `instance` owns instance command/query/proxy/maintenance and `runtime` keeps container capability implementation.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - This is a protected backend architecture boundary change.
  - It touches module ownership, app composition wiring, architecture tests, and architecture docs.

## Files

- Create:
  - `code/backend/internal/module/container_runtime/runtime/module.go`
  - `code/backend/internal/module/container_runtime/runtime/module_test.go`
  - `code/backend/internal/module/container_runtime/architecture_test.go`
- Modify:
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/module/runtime/runtime/module.go`
  - `code/backend/internal/module/runtime/runtime/module_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Review:
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/module/runtime/ports/*.go`
- Test:
  - `go test ./internal/module/container_runtime/... -count=1`
  - `go test ./internal/module/runtime/runtime ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestModuleArchitectureBoundaries' -count=1`
  - `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1`
  - `go test ./internal/app -run TestRouterCompositionStructure -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - Existing modules expose a `runtime.Module` output from module-local `runtime` packages.
  - Existing app composition already names the view `ContainerRuntimeModule`.
  - Existing guardrails already ban cross-module private imports and reviewed application concrete imports.
- Reuse / extend / split / create-new decision:
  - Create `internal/module/container_runtime/runtime` as a new physical module entry.
  - Reuse existing `runtime/application`, `runtime/commands`, `runtime/contracts`, and `runtime/ports` implementation in this first slice instead of moving all implementation packages.
  - Keep legacy `internal/module/runtime/runtime` as a compatibility alias/wrapper only.
- Owner boundary:
  - `container_runtime` owns container capability module assembly.
  - `instance` owns instance business lifecycle and access semantics.
  - `contest`, `practice`, and `challenge` own their business flows and consume container runtime through app composition or narrow ports.
- Why this is the narrowest safe surface:
  - It changes the physical module entry and architecture direction without mixing a large mechanical package move with behavior-preserving refactor.
  - It creates guardrails for the new boundary before further capability migration.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The user clarified a module-boundary design choice: `container_runtime` should be a bottom module.
  - Local evidence shows the current app view exists, while the physical implementation still lives in `internal/module/runtime`.
- grill-with-docs findings:
  - The term `container_runtime` should mean platform/container execution capability, not instance business owner.
  - The first slice must not simply rename the old broad `runtime` module; it must prevent new business owner dependencies from entering the bottom module.
  - A full package move is too broad for one reviewable slice because many existing app, runtime-agent, and repository tests import `runtime/contracts|ports|infrastructure`.
- Plan adjustments after challenge:
  - Add `container_runtime/architecture_test.go` to block `contest/ports`, `practice/ports`, `instance/ports`, and HTTP/API dependencies from the new module runtime package.
  - Keep legacy runtime implementation as substrate for now and document the remaining migration item precisely.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/container_runtime/... -count=1`
  - `cd code/backend && go test ./internal/module/runtime/runtime ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestModuleArchitectureBoundaries' -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1`
  - `cd code/backend && go test ./internal/app -run TestRouterCompositionStructure -count=1`
  - `bash scripts/check-startup-gate.sh`
- Manual checks:
  - Confirm app composition imports `internal/module/container_runtime/runtime` for module build.
  - Confirm `container_runtime/runtime` does not import `contest`, `practice`, or `instance`.
  - Confirm docs distinguish “physical bottom module entry now exists” from “contracts/ports/infrastructure fully migrated”.
- Review focus:
  - Dependency direction and whether the new module accidentally reintroduces business owner coupling.
  - Compatibility risk around old `runtime/runtime.Module` tests and existing app test fixtures.
  - Whether docs overclaim completion beyond the first slice.

## Tasks

### Task 1: Add Container Runtime Physical Module Entry

**Files:**
- Create: `code/backend/internal/module/container_runtime/runtime/module.go`
- Create: `code/backend/internal/module/container_runtime/runtime/module_test.go`
- Create: `code/backend/internal/module/container_runtime/architecture_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/runtime/module_test.go`

- [x] **Step 1: Add the new module package test first**

Run target after adding the test:

```bash
cd code/backend && go test ./internal/module/container_runtime/... -count=1
```

Expected before implementation: package missing or build failure.

- [x] **Step 2: Implement `container_runtime/runtime.Module`**

Move the module assembly logic into the new package, but keep service implementation imports pointed at existing `runtime/application`, `runtime/application/commands`, `runtime/contracts`, and `runtime/ports`.

- [x] **Step 3: Add compatibility wrapper for old `runtime/runtime`**

Keep old package API compiling by aliasing `Module`, `Deps`, and `BackgroundJob`, and forwarding `Build`.

- [x] **Step 4: Add architecture guardrail for new module boundary**

Block the new package from importing `contest/ports`, `practice/ports`, `instance/ports`, Gin, GORM, Redis, Docker SDK, or module API packages.

- [x] **Step 5: Run focused module tests**

```bash
cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/runtime/runtime -count=1
```

Expected: PASS.

### Task 2: Switch App Composition To Container Runtime Module

**Files:**
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
- Modify: `code/backend/internal/app/router_composition_structure_test.go`
- Modify: `code/backend/internal/module/architecture_baseline_test.go`

- [x] **Step 1: Update composition imports and module field types**

Replace `ctf-platform/internal/module/runtime/runtime` usage in app composition with `ctf-platform/internal/module/container_runtime/runtime`.

- [x] **Step 2: Update structure tests to expect the new builder**

Adjust source-string guard tests so they verify `containerruntime.Build` and `containerruntime.Deps`.

- [x] **Step 3: Update module dependency baseline**

Record the new reviewed module dependency edge and remove stale app expectations if needed.

- [x] **Step 4: Run focused app and architecture tests**

```bash
cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1
cd code/backend && go test ./internal/app -run TestRouterCompositionStructure -count=1
cd code/backend && go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestModuleArchitectureBoundaries' -count=1
```

Expected: PASS.

### Task 3: Update Architecture Facts

**Files:**
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] **Step 1: Update current architecture doc**

Record that `container_runtime` now has a physical module entry for builder/wiring, while lower-level contracts/ports/infrastructure migration remains in progress.

- [x] **Step 2: Narrow TODO wording**

Change the active debt from “physical owner not defined” to “lower-level capability packages and remaining runtime persistence are not fully migrated”.

- [x] **Step 3: Run workflow/startup gate check**

```bash
bash scripts/check-startup-gate.sh
```

Expected: PASS.

### Task 4: Completion Validation And Review

**Files:**
- Review changed files from Tasks 1-3.

- [x] **Step 1: Run narrow validation**

```bash
cd code/backend && go test ./internal/module/container_runtime/... ./internal/module/runtime/runtime -count=1
cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1
cd code/backend && go test ./internal/app -run TestRouterCompositionStructure -count=1
cd code/backend && go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestModuleArchitectureBoundaries' -count=1
```

Expected: PASS.

- [x] **Step 2: Run completion-full if narrow validation passes**

```bash
bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
```

Expected: PASS or only documented unrelated failures.

- [x] **Step 3: Run independent backend architecture review**

Archive review under `docs/reviews/architecture/2026-06-10-container-runtime-module-boundary-review.md`.

- [x] **Step 4: Fix material findings and rerun impacted validation**

Expected: all material review findings resolved or explicitly rejected with technical rationale.

Resolution note:

- Fixed old `runtime/runtime` compatibility surface by keeping an explicit wrapper `Module` with the legacy `ContestContainerFiles` exported field and delegating construction to `container_runtime/runtime`.
- Fixed `container_runtime` boundary guard coverage by blocking concrete transport / persistence imports with exact-or-prefix matching, so Docker SDK subpackages such as `github.com/docker/docker/client` are covered.
- Reran impacted validation:

```bash
cd code/backend && timeout 180s go test ./internal/module/runtime/runtime -count=1
cd code/backend && timeout 180s go test ./internal/module/container_runtime/... -count=1
cd code/backend && timeout 180s go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1
cd code/backend && timeout 180s go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries' -count=1
cd code/backend && timeout 180s go test ./internal/app -run 'TestArchitectureRulesConcreteCrossModuleImportExceptionsAreCurrent|TestArchitectureRulesRejectConcreteCrossModuleImports|TestRouterCompositionStructure' -count=1
timeout 900s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
```

Result: all commands passed.
