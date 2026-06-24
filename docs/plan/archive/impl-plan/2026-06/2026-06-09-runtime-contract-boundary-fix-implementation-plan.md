<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 修复 runtime contracts 边界反向依赖 Implementation Plan

**Goal:** Remove the `runtime/contracts/portreservation` factory that lets a contracts package create runtime infrastructure, and add a guardrail that prevents contracts/ports/domain packages from importing infrastructure or concrete persistence clients.

**Architecture:** Keep `contracts` and `ports` as dependency-boundary packages only. `practice/infrastructure` may adapt runtime port reservation, but the concrete runtime repository must be provided explicitly by wiring code or an injected constructor dependency instead of being hidden behind `runtime/contracts`.

**Tech Stack:** Go, GORM, existing `internal/testutil/archtest` architecture tests, CTF modular monolith layout.

---

## Task Metadata

- Task Slug: `2026-06-09-runtime-contract-boundary-fix`
- Started At: `2026-06-09T03:21:06Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-runtime-contract-boundary-fix`
- Branch: `task/2026-06-09-runtime-contract-boundary-fix`

## Objective And Non-Goals

- Objective:
  - Remove `code/backend/internal/module/runtime/contracts/portreservation`.
  - Make `practice/infrastructure.Repository` receive a `runtimeports.PortReservationOwner` explicitly while preserving the existing `NewRepository(db)` public constructor behavior.
  - Add an architecture test proving `contracts`, `ports`, and `domain` packages cannot import `infrastructure`, GORM, Redis, Docker SDKs, Gin, or other concrete outer-layer packages.
- Non-Goals:
  - Do not redesign the larger `runtime/ports` to `instance/ports` alias surface in this slice.
  - Do not change instance start/restart behavior or database schema.
  - Do not touch the unrelated runtime behavior test failure observed during the earlier architecture audit.

## Inputs

- Source docs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Related architecture/contracts:
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/practice/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/ports/port_reservation.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
- Related prior work:
  - Existing module boundary guardrails already reject cross-module private imports and application -> concrete outer-layer imports.

## Task Classification

- Classification: `非琐碎任务`
- Why: This changes backend module wiring and tightens architecture guardrails across `internal/module`.

## Files

- Create:
  - `code/backend/internal/module/practice/application/commands/runtime_port_owner_test.go`
  - `code/backend/internal/module/practice/application/commands/runtime_port_owner_external_test.go`
- Modify:
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `code/backend/internal/app/composition/practice_module.go`
  - `code/backend/internal/module/practice/runtime/module.go`
  - `code/backend/internal/module/practice/infrastructure/repository.go`
- Remove:
  - `code/backend/internal/module/runtime/contracts/portreservation/owner.go`
- Review:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/runtime/ports/port_reservation.go`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Test:
  - `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`
  - `go test ./internal/module -count=1`
  - `go test ./internal/module/practice/infrastructure -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - `runtime` and `challenge` module runtime wiring use concrete infrastructure only from runtime/composition layers.
  - `architecture_test.go` already parses imports through `archtest.RuntimeGoFiles`.
- Reuse / extend / split / create-new decision:
  - Extend `architecture_test.go` instead of adding a new test helper.
  - Keep `practice/infrastructure.NewRepository(db)` for non-port repository use, but require port-owning paths to call `NewRepositoryWithRuntimePortOwner(db, ownerFor)`.
  - Pass `runtimeinfra.NewRepository` from app composition through `practice/runtime.Deps.RuntimePortOwnerFor` so production wiring stays explicit.
- Owner boundary:
  - `runtime/ports.PortReservationOwner` remains the capability contract.
  - `runtime/infrastructure.NewRepository(db)` remains the concrete owner implementation.
  - `practice/infrastructure.Repository` only receives and delegates to that owner.
- Why this is the narrowest safe surface:
  - It removes the concrete dependency from `contracts` without changing callers above repository construction or changing runtime port reservation behavior.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The implementation direction depends on architecture ownership, not user-visible behavior.
- grill-with-docs findings:
  - Existing docs already say composition/runtime wiring owns concrete implementation assembly; contracts/ports/domain should not create infrastructure.
- Plan adjustments after challenge:
  - Keep the larger `container_runtime` capability landing-zone debt out of scope; this slice only closes the concrete contracts leak.

## Validation

- Commands:
  - Red: `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`
  - Green: `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`
  - Package: `go test ./internal/module -count=1`
  - Focused practice repository: `go test ./internal/module/practice/infrastructure -count=1`
- Manual checks:
  - `rg 'runtime/contracts/portreservation|NewRepositoryWithRuntimePorts' code/backend/internal/module code/backend/internal/app -g'*.go'`
- Review focus:
  - Verify no contracts/ports/domain package can import outer-layer concrete dependencies.
  - Verify `practice/infrastructure` still uses runtime port reservation on the same DB/transaction handle.
  - Verify the deleted factory cannot reappear through a new contracts subpackage.

## Task 1: Guardrail And Wiring Fix

**Files:**
- Modify: `code/backend/internal/module/architecture_test.go`
- Modify: `code/backend/internal/module/practice/infrastructure/repository.go`
- Remove: `code/backend/internal/module/runtime/contracts/portreservation/owner.go`

- [x] **Step 1: Write the failing architecture test**

Add a test that scans runtime Go files under `internal/module` and fails when a `contracts`, `ports`, or `domain` package imports an outer-layer package such as `/infrastructure`, `gorm.io/gorm`, Redis, Docker SDKs, Gin, or `database/sql`.

- [x] **Step 2: Verify RED**

Run: `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`

Expected: FAIL because `runtime/contracts/portreservation/owner.go` imports `runtime/infrastructure` and `gorm.io/gorm`.

- [x] **Step 3: Move concrete construction out of contracts**

Update `practice/infrastructure.Repository` so port-owning paths use a new constructor that accepts a `func(*gorm.DB) runtimeports.PortReservationOwner`. Pass the concrete runtime owner from app composition through `practice/runtime.Deps`, keeping practice infrastructure free of runtime infrastructure imports.

- [x] **Step 4: Remove the invalid contracts factory**

Remove `code/backend/internal/module/runtime/contracts/portreservation/owner.go` after its only production caller is gone.

- [x] **Step 5: Verify GREEN and focused packages**

Run:

```bash
go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1
go test ./internal/module -count=1
go test ./internal/module/practice/infrastructure -count=1
```

Expected: PASS for all listed commands.

Additional executed checks:

```bash
go test ./internal/module/practice/application/commands -count=1
go test ./internal/module/practice/... -count=1
go test ./internal/app -run TestPracticeModuleWiresRuntimePortOwnerFromCompositionRoot -count=1
go test ./internal/app -run 'TestArchitectureRulesRejectConcreteCrossModuleImports|TestCompositionDoesNotReintroduceRuntimeCompatFacade|TestCompositionAndRuntimeDoNotImportLegacyContainerModule|TestInstanceModuleDoesNotInjectRetiredDefenseWorkbenchService|TestPracticeModuleUsesTypedPortsDeps|TestPracticeModuleUsesTypedCrossModuleDeps|TestPracticeFlow_PublishedChallengeLifecycleAndAccess|TestPracticeFlow_PublishedChallengeSubmissionsAndProgress|TestPracticeFlow_PublishedChallengeGeneratesTeacherEvidenceAndAuditTrail' -count=1
```

Review gate:

- Independent backend review verdict: `pass`
- Archive: `docs/reviews/backend/2026-06-09-backend-review-runtime-contract-boundary-fix.md`
- Follow-up from review addressed in this slice:
  - Added a focused source-level guard that locks `RuntimePortOwnerFor: runtimePortOwnerFor` in app composition.
  - Removed the empty `runtime/contracts/portreservation` compatibility shell after explicit user confirmation.
