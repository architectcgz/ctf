<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# module-dependency-baseline Implementation Plan

**Goal:** Reduce `moduleDependencyBaseline` by migrating obvious cross-module leaks to owned contracts or consumer ports, without weakening backend module guardrails.

**Architecture:** CTF backend modular monolith / Onion-style module boundaries. The current guard tracks module-owner edges as `owner -> dependency`; this task removes edges by changing real import ownership, not by renaming the baseline.

**Tech Stack:** Go backend, AST/module architecture tests, existing `code-workflow` gates.

---

## Task Metadata

- Task Slug: `2026-06-08-module-dependency-baseline`
- Started At: `2026-06-08T15:40:34Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Branch: `task/2026-06-08-module-dependency-baseline`

## Objective And Non-Goals

- Objective:
  - Turn `moduleDependencyBaseline` from a broad historical edge list into a smaller, reviewable set of intentional module relationships.
  - Start with edges that are mechanically local and have clear owner direction: type/DTO leaks, runtime adapter leaks, and provider-owned constants that should live in `contracts`.
  - Preserve `teaching_query` and other intentional query aggregation semantics unless a narrow migration is obvious.
- Non-Goals:
  - Do not collapse modules or redesign the whole backend dependency graph in one batch.
  - Do not move persistence or runtime implementation into `internal/shared`.
  - Do not relax `TestModuleDependencyBaselineIsCurrent`; baseline entries should disappear only after the imports disappear or the guard becomes more precise.
  - Do not mix module-dependency cleanup with transaction/time guard cleanup already completed in `task/2026-06-08-backend-architecture-guard-quality`.

## Inputs

- Source docs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `AGENTS.md`
- Related architecture/contracts:
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/testutil/archtest/archtest.go`
- Related prior work:
  - `f2794c5d0 Merge branch 'task/2026-06-08-backend-architecture-guard-quality'`
  - Previous guard cleanup reduced time/transaction baselines and left `moduleDependencyBaseline` as the largest remaining broad baseline.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - This changes backend module ownership and architecture guard behavior.
  - Each batch can affect imports, generated mappers, ports/contracts, runtime wiring, and tests across several bounded contexts.
  - Independent review is required before claiming a batch complete.

## Files

- Create:
  - Possibly new narrow `contracts` files when a provider needs to expose stable data shapes or constants.
  - Possibly new consumer-side `ports` interfaces when application code should depend on a capability instead of a provider module.
- Modify:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/architecture_test.go` if the guard needs a more precise category than raw `owner -> dependency`.
  - Backend module files under `application`, `domain`, `ports`, `contracts`, `runtime`, and `infrastructure` touched by each migration batch.
- Review:
  - All current `moduleDependencyBaseline` entries.
  - Provider/consumer owner direction for every removed edge.
  - Generated mapper wrappers if DTO/contract fields move.
- Test:
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - Affected package tests for each migrated module.
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Independent review gate, then `workflow-governance`.

## 复用与 Owner 决策

- Existing patterns searched:
  - Provider contracts: `identity/contracts`, `instance/contracts`, `runtime/contracts`, `contest/contracts`, `challenge/contracts`.
  - Consumer ports: module-local `ports` packages already model capabilities needed by application code.
  - Query aggregation owner: `teaching_query` is explicitly documented as the cross-owner read aggregation module.
- Reuse / extend / split / create-new decision:
  - Reuse existing provider `contracts` for stable cross-module data shapes and constants.
  - Use consumer `ports` when the consumer needs behavior/capability rather than a data contract.
  - Avoid new shared packages unless the abstraction is truly stable across bounded contexts.
- Owner boundary:
  - Provider modules own their public `contracts`.
  - Consumer modules own the `ports` they require.
  - `runtime` owns container capability implementation, but `instance` owns instance business state and access views.
  - `teaching_query` owns teacher-facing cross-owner read models; it should not be treated the same as write-side coupling.
- Why this is the narrowest safe surface:
  - The first batches can remove dependency edges by moving only type/constant imports or adapter boundaries, without schema or API behavior changes.
  - Larger write-side flows such as `practice -> contest/instance/runtime` should be split after the cheap leaks are gone, because they may need capability ports and runtime wiring changes.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The request is architecture cleanup with multiple possible owner directions. The first decision is batch order and boundary classification, not code edits.
- Evidence inspected:
  - Current `moduleDependencyBaseline` has 37 edges.
  - `TestModuleDependencyBaselineIsCurrent` currently records raw module owner edges, regardless of whether imports target `contracts`, `ports`, `application`, or `runtime`.
  - Edge scan across non-test runtime Go files grouped current imports by `owner -> dependency` and file.
  - Architecture docs explicitly allow `teaching_query` as cross-owner read aggregation, and document `runtime`/`instance` split as still partially transitional.
- grill-with-docs findings:
  - The term "module dependency" is too broad for implementation decisions. This task should distinguish write-side coupling, read aggregation, provider contract imports, consumer port imports, and transitional runtime/instance coupling.
  - `teaching_query -> assessment/challenge/contest/identity` is likely an intentional query-aggregation category, not the first migration target.
  - `runtime -> instance` and `instance -> runtime` are the riskiest pair because docs say the runtime/instance physical split is still partially transitional.
  - First implementation slice should target obviously misplaced imports or guard precision, not deep cross-module workflow redesign.
- Plan adjustments after challenge:
  - Do not start by deleting all edges.
  - Batch 1 targets cheap, owner-obvious edges and/or guard precision.
  - Defer high-coupling runtime/instance/practice/contest flows until each has a concrete owner migration plan.

## Current Edge Groups

- Query aggregation / likely intentional:
  - `teaching_query -> assessment`
  - `teaching_query -> challenge`
  - `teaching_query -> contest`
  - `teaching_query -> identity`
- Authentication / identity flow:
  - `auth -> identity`
  - `ops -> auth`
- Challenge and contest coupling:
  - `challenge -> contest`
  - `contest -> challenge`
  - `contest -> instance`
  - `contest -> runtime`
  - `contest -> identity`
  - `contest -> auth`
- Practice orchestration coupling:
  - `practice -> challenge`
  - `practice -> contest`
  - `practice -> identity`
  - `practice -> instance`
  - `practice -> runtime`
- Runtime / instance transitional coupling:
  - `runtime -> instance`
  - `instance -> runtime`
  - `runtime -> challenge`
  - `runtime -> contest`
  - `runtime -> identity`
  - `runtime -> ops`
- Assessment coupling:
  - `assessment -> contest`
  - `assessment -> challenge`
  - `assessment -> identity`
  - `assessment -> practice`
  - `assessment -> teaching_query`
- Ops notifications / dashboard coupling:
  - `ops -> challenge`
  - `ops -> contest`
  - `ops -> identity`
  - `ops -> practice`
- Challenge / instance coupling:
  - `challenge -> identity`
  - `challenge -> instance`
  - `challenge -> runtime`
- Instance / contest / identity coupling:
  - `instance -> contest`
  - `instance -> identity`

## Initial Batch Strategy

1. Guard precision batch:
   - Consider splitting `moduleDependencyBaseline` into categories, e.g. reviewed query-aggregation edges vs reviewed transitional owner edges.
   - This is useful only if it makes future migrations stricter, not if it hides debt.
2. Cheap contract leak batch:
   - Move stable constants/types to provider `contracts` where current imports pull a module edge only for data shape.
   - Candidate areas from scan: `challenge -> contest`, `runtime -> identity`, `runtime -> ops`, `assessment -> teaching_query`.
3. Consumer port batch:
   - For application services importing provider modules for behavior, define consumer ports and wire adapters in `runtime` / app composition.
   - Candidate areas: `assessment -> practice`, `ops -> challenge/contest/practice`, some `instance -> contest`.
4. Runtime / instance split batch:
   - Treat separately because docs already identify this as transitional.
   - Any change here must preserve `ContainerRuntimeModule` vs `InstanceModule` views.
5. Query aggregation classification:
   - Preserve `teaching_query` read aggregation unless evidence shows a write-side dependency or provider contract leak.

## Validation

- Commands:
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `go test ./internal/module/...` for affected modules as needed.
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
- Manual checks:
  - Compare removed baseline entries with actual imports.
  - Verify no application/domain layer gains concrete provider imports.
  - Verify generated mapper changes are regenerated when DTO shape changes.
- Review focus:
  - Whether each removed edge is backed by real owner migration.
  - Whether any guard precision change makes the baseline weaker.
  - Whether query aggregation and runtime/instance transitional edges are documented honestly instead of hidden.
