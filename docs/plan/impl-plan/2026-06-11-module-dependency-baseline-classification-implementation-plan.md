<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# moduleDependencyBaseline 分类标注 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or an equivalent step-by-step execution loop. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `moduleDependencyBaseline` describe why each reviewed module edge remains, not only that the edge exists.

**Architecture:** Keep the existing import scanner and baseline owner in `internal/module`; change only the baseline value type from an empty marker to structured review metadata. The guard should still fail on unknown or stale edges, and additionally fail when a baseline entry has no category or rationale.

**Tech Stack:** Go backend architecture tests, modular monolith docs, code-workflow gates.

---

## Task Metadata

- Task Slug: `2026-06-11-module-dependency-baseline-classification`
- Started At: `2026-06-11T01:46:12Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-module-dependency-baseline-classification`
- Branch: `task/2026-06-11-module-dependency-baseline-classification`

## Objective And Non-Goals

- Objective:
  - Replace `moduleDependencyBaseline` empty map values with structured category/rationale metadata.
  - Add guard coverage so new or retained baseline entries must carry non-empty classification.
  - Update the backend modular monolith architecture note to explain the baseline categories.
- Non-Goals:
  - Do not remove additional module edges in this slice.
  - Do not change runtime behavior, public APIs, database schema, or module wiring.
  - Do not split the architecture scanner by package layer in this slice.
  - Do not start the assessment training-fact semantic unification work.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/impl-plan/2026-06-10-ops-assessment-event-readmodel-boundary-plan.md`
- Related architecture/contracts:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/architecture_test.go`
- Related prior work:
  - `container_runtime -> contest`, `instance -> contest`, and `assessment -> teaching_query` have already been removed from `moduleDependencyBaseline`.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Touches backend architecture guardrails and the module-boundary fact source.
  - The code change is small, but it affects how future cross-module dependencies are reviewed.

## Files

- Create:
  - None.
- Modify:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/architecture_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/impl-plan/2026-06-11-module-dependency-baseline-classification-implementation-plan.md`
- Review:
  - Confirm categories do not hide debt by turning every edge into a generic "allowed" bucket.
  - Confirm unknown/stale edge detection remains unchanged.
  - Confirm docs match the code categories.
- Test:
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## 复用与 Owner 决策

- Existing patterns searched:
  - Current `moduleDependencyBaseline` in `architecture_baseline_test.go`.
  - Existing `TestModuleDependencyBaselineIsCurrent` scanner in `architecture_test.go`.
  - Architecture docs table under `docs/architecture/backend/07-modular-monolith-refactor.md`.
- Reuse / extend / split / create-new decision:
  - Extend the existing baseline map in place.
  - Do not introduce a separate YAML/JSON registry; that would create a second fact source.
- Owner boundary:
  - `internal/module` architecture tests remain the owner of backend module dependency guardrails.
  - Architecture docs explain category semantics, but the test map is the mechanical source.
- Why this is the narrowest safe surface:
  - Only the baseline metadata shape changes.
  - Existing import scanning, private-layer checks, and stale-entry behavior stay intact.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task changes how future architecture decisions are represented; the key choice is metadata shape and taxonomy, not runtime code.
- Evidence inspected:
  - Remaining baseline has 27 entries.
  - Current baseline test only stores `map[string]struct{}` and cannot distinguish query aggregation, event consumption, provider contracts, or ports.
  - Architecture docs already state that `teaching_query` query aggregation, `ops` event consumption, and `assessment` analysis read-model dependencies are intentional.
- grill-with-docs findings:
  - The term "module dependency" is too broad. The guard should record why each edge remains so future reviews do not treat all edges as equal debt.
  - The classification must not weaken unknown/stale-edge detection.
  - The taxonomy should stay small and tied to documented collaboration modes.
- Plan adjustments after challenge:
  - Use structured metadata directly in Go tests.
  - Add a guard for missing category/rationale.
  - Update docs with category meanings instead of adding a new registry file.

## Tasks

- [x] **Step 1: Add a RED guard for classified baseline metadata**
  - Update `TestModuleDependencyBaselineIsCurrent` to expect every baseline value to expose categories and rationale.
  - Run: `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - Expected: FAIL to compile or fail because current entries still use empty marker values.

- [x] **Step 2: Add baseline entry metadata types and categories**
  - In `architecture_baseline_test.go`, add a small typed metadata struct.
  - Add category constants for provider contracts, port boundaries, runtime capability, event consumers, query aggregation, and analysis read models.

- [x] **Step 3: Classify all current baseline entries**
  - Replace every `{}` value with category/rationale metadata.
  - Keep all 27 existing keys unchanged.
  - Do not add or delete dependency edges.

- [x] **Step 4: Verify GREEN for module baseline guard**
  - Run: `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - Expected: PASS.

- [x] **Step 5: Update architecture documentation**
  - Update `docs/architecture/backend/07-modular-monolith-refactor.md` to state that `moduleDependencyBaseline` is now classified by dependency purpose.
  - Mention that this does not mean the edge is automatically permanent; it means the reviewed reason is explicit.

- [x] **Step 6: Run focused architecture validation**
  - Run: `cd code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1`
  - Expected: PASS.

- [x] **Step 7: Run completion validation**
  - Run: `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Expected: PASS.

- [x] **Step 8: Review gate handling**
  - Current conversation direction is not to run independent review for this baseline pass.
  - Do a local self-check for category drift and report explicitly that no independent review was run.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1`
  - `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - Baseline still has 27 entries.
  - No runtime code outside architecture tests changes.
  - Documentation category names match Go constants.
- Review focus:
  - Categories should clarify review intent without becoming broad exemptions.
  - Query aggregation and event consumption should remain distinguishable from owner-coupled write-side dependencies.
