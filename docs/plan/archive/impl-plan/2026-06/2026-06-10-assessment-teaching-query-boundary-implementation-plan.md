<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# 2026-06-10-assessment-teaching-query-boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or an equivalent step-by-step execution loop. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the production `assessment -> teaching_query` module dependency while preserving class report content and teacher query behavior.

**Architecture:** `assessment` owns report export commands and should not depend inward on the teacher read-model module. `teaching_query -> assessment` remains an intentional query aggregation edge through `assessment/contracts`; `assessment` consumes class insight through its own port types, with app composition mapping the `teaching_query` repository adapter at the edge.

**Tech Stack:** Go backend, modular monolith architecture tests, GORM-backed teaching query repository, code-workflow gates.

---

## Task Metadata

- Task Slug: `2026-06-10-assessment-teaching-query-boundary`
- Started At: `2026-06-10T15:46:41Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-assessment-teaching-query-boundary`
- Branch: `task/2026-06-10-assessment-teaching-query-boundary`

## Objective And Non-Goals

- Objective:
  - Make `assessment` production code free of `ctf-platform/internal/module/teaching_query/...` imports.
  - Keep class report JSON/PDF/XLSX output shape unchanged.
  - Remove `assessment -> teaching_query` from `moduleDependencyBaseline` after imports are gone.
  - Update the backend architecture fact source to say the reverse edge has been closed.
- Non-Goals:
  - Do not remove `teaching_query -> assessment`; it is still the target direction for recommendation read aggregation.
  - Do not redesign `teaching_query` read models, recommendation semantics, or class review advice logic.
  - Do not move `teaching_query/infrastructure.Repository`; it can still be wired by app composition.
  - Do not broaden this slice into practice/contest/assessment semantic unification.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-module-dependency-baseline-implementation-plan.md`
- Related architecture/contracts:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/app/composition/assessment_module.go`
- Related prior work:
  - Runtime module retirement and reverse dependency convergence removed the earlier `runtime`, `container_runtime -> contest`, and `instance -> contest` edges.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Cross-module backend boundary refactor touching application commands, ports, composition wiring, architecture guards, tests, and architecture docs.
  - This closes a documented structural debt edge in `moduleDependencyBaseline`.
  - Requires TDD red/green and independent review gate before completion.

## Files

- Create:
  - `code/backend/internal/app/composition/assessment_class_insight_adapter.go`
- Modify:
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/commands/report_service_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_class_test.go`
  - `code/backend/internal/app/composition/assessment_module.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Review:
  - Confirm only app composition imports `teaching_query` on behalf of assessment wiring.
  - Confirm `teaching_query -> assessment` remains and is still only through contracts for recommendation provider usage.
  - Confirm class report output fields and JSON tags stay equivalent.
- Test:
  - `cd code/backend && go test ./internal/module/assessment -run TestAssessmentRuntimeCodeDoesNotDependOnTeachingQuery -count=1`
  - `cd code/backend && go test ./internal/module/assessment/application/commands -run 'TestBuildClassReportDataUsesSharedWindowedClassInsight|TestCreateClassReportRejectsCrossClassTeacherRequest' -count=1`
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && go test ./internal/module/assessment/... ./internal/app -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - Consumer-side ports in `assessment/ports`.
  - App composition adapters such as `teachingQueryUserLookupAdapter`.
  - Existing teaching class insight data shape in `teaching_query/ports`.
  - Class review neutral package in `internal/teaching/classreview`.
- Reuse / extend / split / create-new decision:
  - Reuse `teaching_query/infrastructure.Repository` as the concrete data reader.
  - Add assessment-owned `ClassInsightSummary`, `ClassInsightTrend`, and `ClassInsightTrendPoint` port data shapes.
  - Add a composition adapter that maps `teaching_query/ports` results into assessment port shapes.
  - Keep report output DTOs local to `report_service.go` because they are export payload shape, not public teaching query contracts.
- Owner boundary:
  - `assessment` owns report generation and its report payload shape.
  - `teaching_query` owns teacher-facing HTTP/read-model contract shape.
  - `app/composition` owns cross-module adapter wiring.
- Why this is the narrowest safe surface:
  - No database query, API route, permission, or report behavior changes are needed.
  - The change removes only type ownership coupling and baseline entry, preserving the existing repository implementation and report workflow.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The key choice is ownership direction, not algorithm design. The code and architecture docs answer the ambiguity without asking the user.
- Evidence inspected:
  - `moduleDependencyBaseline` currently includes both `assessment -> teaching_query` and `teaching_query -> assessment`.
  - `assessment/ports/ports.go` imports `teaching_query/ports`.
  - `assessment/application/commands/report_service.go` imports `teaching_query/contracts` and `teaching_query/ports`.
  - `app/composition/assessment_module.go` wires `teaching_query/infrastructure.Repository` into assessment.
  - `docs/architecture/backend/07-modular-monolith-refactor.md` identifies `assessment -> teaching_query` as the remaining reverse edge.
- Chosen direction:
  - Keep report export in `assessment`.
  - Keep teaching query repository reuse at app composition edge.
  - Convert all assessment-internal types to assessment-owned or neutral types.
- grill-with-docs findings:
  - Moving report export into `teaching_query` would mix report lifecycle/storage/export ownership into the teacher read model and widen scope.
  - Moving the repository implementation into shared code would be larger than needed and risks creating a generic shared query bucket.
  - A composition adapter is the smallest DIP-compliant fix because the concrete query source already exists and only its return types leak across the boundary.
- Plan adjustments after challenge:
  - Add a module-local architecture guard so the same reverse edge cannot return through another assessment file.
  - Keep baseline update as the final code step, after production imports are gone.

## Execution Steps

- [x] **Step 1: Write the failing assessment boundary test**
  - Add `TestAssessmentRuntimeCodeDoesNotDependOnTeachingQuery` in `code/backend/internal/module/assessment/architecture_test.go`.
  - The test should walk non-test `.go` files under assessment and fail on any import path with prefix `ctf-platform/internal/module/teaching_query`.

- [x] **Step 2: Verify RED**
  - Run: `cd code/backend && go test ./internal/module/assessment -run TestAssessmentRuntimeCodeDoesNotDependOnTeachingQuery -count=1`
  - Expected: FAIL, pointing at current `assessment/ports/ports.go` or `report_service.go` imports.

- [x] **Step 3: Move class insight port shapes into assessment**
  - In `assessment/ports/ports.go`, remove `teaching_query/ports` import.
  - Add assessment-owned summary/trend structs.
  - Change `AssessmentClassInsightRepository` to return those structs.
  - Update assessment command tests to use assessment port types.

- [x] **Step 4: Replace report output DTO dependency**
  - In `report_service.go`, remove `teaching_query/contracts` and `teaching_query/ports` imports.
  - Add local class report summary/trend/review DTO structs with the same JSON tags.
  - Update mapper and helper signatures to use assessment-owned/local types.

- [x] **Step 5: Add the composition adapter**
  - Create `app/composition/assessment_class_insight_adapter.go`.
  - Wrap `teaching_query/ports.TeachingClassInsightRepository`.
  - Map summary/trend values into `assessmentports` shapes and pass through teaching fact snapshots.
  - Wire `BuildAssessmentModule` through the adapter.

- [x] **Step 6: Verify GREEN for focused tests**
  - Run: `cd code/backend && go test ./internal/module/assessment -run TestAssessmentRuntimeCodeDoesNotDependOnTeachingQuery -count=1`
  - Run: `cd code/backend && go test ./internal/module/assessment/application/commands -run 'TestBuildClassReportDataUsesSharedWindowedClassInsight|TestCreateClassReportRejectsCrossClassTeacherRequest' -count=1`
  - Expected: PASS.

- [x] **Step 7: Remove stale module baseline edge**
  - Delete `assessment -> teaching_query` from `moduleDependencyBaseline`.
  - Run: `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - Expected: PASS.

- [x] **Step 8: Update architecture fact source**
  - Update `docs/architecture/backend/07-modular-monolith-refactor.md` to state that `assessment -> teaching_query` is closed.
  - Keep any remaining caveat about `teaching_query -> assessment` being intentional.

- [x] **Step 9: Run completion validation**
  - Run: `cd code/backend && go test ./internal/module/assessment/... -count=1`
  - Run: `cd code/backend && go test ./internal/app -run 'TestAssessmentModuleUsesTypedPortsDeps|TestAssessmentModuleUsesTypedCrossModuleDeps' -count=1`
  - Run: `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Note: a wider `go test ./internal/module/assessment/... ./internal/app -count=1` run exposed existing app fixture / source marker failures outside this slice, so the final gate used the focused app composition tests plus `completion-full`.

- [x] **Step 10: Review gate skipped by user instruction**
  - User explicitly changed direction to `不review了`, so no independent review artifact was produced for this slice.
  - Keep the implemented guard and workflow validation as the merge evidence.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/assessment -run TestAssessmentRuntimeCodeDoesNotDependOnTeachingQuery -count=1`
  - `cd code/backend && go test ./internal/module/assessment/application/commands -run 'TestBuildClassReportDataUsesSharedWindowedClassInsight|TestCreateClassReportRejectsCrossClassTeacherRequest' -count=1`
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && go test ./internal/module/assessment/... -count=1`
  - `cd code/backend && go test ./internal/app -run 'TestAssessmentModuleUsesTypedPortsDeps|TestAssessmentModuleUsesTypedCrossModuleDeps' -count=1`
  - `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - `rg -n "ctf-platform/internal/module/teaching_query" code/backend/internal/module/assessment -g '*.go' -g '!*_test.go'` returns no production matches.
  - `moduleDependencyBaseline` still includes `teaching_query -> assessment`.
- Review focus:
  - No class report output shape drift.
  - No new shared package or provider-owned interface.
  - No broad `teaching_query` implementation move.
  - Adapter remains in app composition, not assessment application or ports.
