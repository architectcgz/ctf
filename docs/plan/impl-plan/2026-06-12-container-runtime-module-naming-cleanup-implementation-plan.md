<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# container-runtime-module-naming-cleanup Implementation Plan

**Goal:** Remove the leftover `RuntimeModule` / `BuildRuntimeModule` compatibility naming from app composition, then rename the composition source file to align with the container-runtime module owner naming.

**Architecture:** This is a narrow backend composition naming cleanup. The runtime capability owner stays `internal/module/container_runtime`; the work only touches app composition entrypoints, active callers, architecture guard tests, and current-fact backend docs, without changing runtime behavior or external HTTP contracts.

**Tech Stack:** Go backend, app composition guard tests, code-workflow startup gate.

---

## Task Metadata

- Task Slug: `2026-06-12-container-runtime-module-naming-cleanup`
- Started At: `2026-06-11T16:31:39Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-container-runtime-module-naming-cleanup`
- Branch: `task/2026-06-12-container-runtime-module-naming-cleanup`

## Objective And Non-Goals

- Objective:
  - Remove `RuntimeModule` and `BuildRuntimeModule` compatibility aliases from app composition.
  - Rename the composition source file from `runtime_module.go` to a container-runtime-aligned file name and update active references.
  - Keep current backend docs and guardrails aligned with the active naming.
- Non-Goals:
  - Do not rename the `container_runtime` module root or any runtime capability contracts.
  - Do not change runtime behavior, lifecycle wiring, or external HTTP routes.

## Inputs

- Source docs:
  - `AGENTS.md`
- Related architecture/contracts:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/app/composition/runtime_module.go`
- Related prior work:
  - `dec1303b6 refactor(backend): 重命名 teaching_analysis 模块`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - The cleanup touches protected app composition wiring, startup-gated workflow state, and current architecture facts across code and docs.

## Files

- Create:
  - None for the first slice; the second slice may create a renamed composition file path when the source file is moved.
- Modify:
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `code/backend/internal/testutil/systemapp/practice_flow.go`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Review:
  - Active references to `BuildRuntimeModule`, `RuntimeModule`, and `runtime_module.go` outside archived plans/reviews.
- Test:
  - `go test ./internal/app -run 'TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestRuntimeStatePersistenceWiredFromCompositionRoot|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule' -count=1`
  - `go test ./internal/app -run 'TestPracticeFlow' -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - Current module roots under `code/backend/internal/module/*`
  - Existing composition entrypoints already aligned as `*_module.go`
- Reuse / extend / split / create-new decision:
  - Reuse the existing `ContainerRuntimeModule` owner surface and extend current guardrails instead of introducing a second compatibility alias.
- Owner boundary:
  - `container_runtime` remains the runtime capability owner; app composition only exposes its container-facing view to `challenge` / `contest` / `ops` / `instance`.
- Why this is the narrowest safe surface:
  - It removes naming drift without moving business owners, changing imports across unrelated modules, or altering runtime semantics.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `architect-agent`
- Why this pass fits:
  - The task is a boundary naming cleanup where the main risk is leaving active composition references half-migrated.
- grill-with-docs findings:
  - No second active module-root naming problem remains after `teaching_analysis`; the remaining drift is the composition compatibility layer around `ContainerRuntimeModule`.
- Plan adjustments after challenge:
  - Split the work into two commits: first remove active compatibility aliases, then rename the physical composition file after explicit user authorization.

## Validation

- Commands:
  - `go test ./internal/app -run 'TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestRuntimeStatePersistenceWiredFromCompositionRoot|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule' -count=1`
  - `go test ./internal/app -run 'TestPracticeFlow' -count=1`
- Manual checks:
  - Confirm active implementation and current-fact backend docs no longer claim that `RuntimeModule` is kept as a compatibility alias.
- Review focus:
  - Leftover active references to `BuildRuntimeModule` / `RuntimeModule`
  - Path references that still assume `runtime_module.go` after the second slice
