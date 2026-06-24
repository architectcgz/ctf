<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# assessment / teaching_query 反向依赖收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the production `assessment -> teaching_query` dependency while preserving class report export behavior and the valid `teaching_query -> assessment` recommendation-provider dependency.

**Architecture:** `assessment` owns report generation and report export DTOs. `teaching_query` keeps the teacher read-model repository. App composition owns the private adapter that maps teaching-query read-model output into assessment-owned class insight types, so module code no longer imports back across the query aggregation layer.

**Tech Stack:** Go backend, modular monolith architecture tests, GORM-backed read models, report PDF/Excel/JSON rendering.

---

## Task Metadata

- Task Slug: `2026-06-10-assessment-teaching-query-reverse-dependency`
- Started At: `2026-06-10T14:48:26Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-assessment-teaching-query-reverse-dependency`
- Branch: `task/2026-06-10-assessment-teaching-query-reverse-dependency`
- Requested Plan Path: `docs/plan/archive/impl-plan/2026-06/2026-06-10-assessment-teaching-query-reverse-dependency-plan.md` was not present in the repository or Git history.
- Active Gate Plan Path: `docs/plan/archive/impl-plan/2026-06/2026-06-10-assessment-teaching-query-reverse-dependency-implementation-plan.md`

## Objective And Non-Goals

- Objective:
  - Add a production architecture guard proving `assessment` no longer imports `internal/module/teaching_query`.
  - Replace `assessment/ports.AssessmentClassInsightRepository` return types with assessment-owned class insight structs.
  - Update class report generation/rendering/tests to use assessment-owned report types.
  - Add a private app-composition adapter from `teaching_query` repository output to `assessment` port output.
  - Remove `assessment -> teaching_query` from `moduleDependencyBaseline` only after the real production import disappears.
  - Update backend architecture fact source to mark the reverse edge as resolved.
- Non-Goals:
  - Do not change report API routes, report file formats, JSON field names, PDF/Excel visible content, database schema, class insight query SQL, or recommendation logic.
  - Do not remove the valid `teaching_query -> assessment` dependency; teacher class review still reads recommendation provider contracts from assessment.
  - Do not move class report export ownership to `teaching_query` in this slice.

## Inputs

- Source docs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/README.md`
  - `docs/plan/README.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/teaching_query/ports/query.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-module-reverse-dependency-convergence-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-module-reverse-dependency-convergence-slice-3-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: Backend structural refactor across `assessment`, `teaching_query`, app composition, architecture baseline, tests, and architecture documentation.

## Files

- Create:
  - `code/backend/internal/app/composition/assessment_class_insight_adapter.go`
- Modify:
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/internal/module/assessment/domain/report.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/commands/report_service_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_class_test.go`
  - `code/backend/internal/app/composition/assessment_module.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Review:
  - `code/backend/internal/module/teaching_query/ports/query.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository.go`
  - `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- Test:
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_class_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_service_test.go`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/module/architecture_baseline_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Assessment domain report types already carry report-export JSON shape.
  - Assessment ports already define consumer-side repositories for report generation.
  - Composition already owns cross-module glue such as `BuildAssessmentModule`.
  - Teaching query repository already returns neutral port structs for class summary/trend and shared `teaching/advice` snapshots.
- Reuse / extend / split / create-new decision:
  - Extend `assessment/domain/report.go` with assessment-owned class insight export structs instead of importing `teaching_query/contracts`.
  - Keep `teaching_query/ports` unchanged for teacher read-model services.
  - Create a private composition adapter instead of making `assessment` depend on `teaching_query` or moving the read-model SQL.
- Owner boundary:
  - `assessment`: report generation, report export payload shape, class report review rendering, class report repository port.
  - `teaching_query`: teacher read-model SQL and teacher-facing query services.
  - `app/composition`: private mapping adapter between provider read model and consumer report port.
- Why this is the narrowest safe surface:
  - The behavior remains the same because the same teaching-query repository still supplies summary/trend/snapshots.
  - Only type ownership and wiring move; SQL, routes, schema, and report rendering semantics stay unchanged.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: The task is a structural owner-boundary cleanup; the main decision is where the neutral class insight type and adapter should live.
- Evidence inspected:
  - `moduleDependencyBaseline` contains both `assessment -> teaching_query` and `teaching_query -> assessment`.
  - `docs/architecture/backend/07-modular-monolith-refactor.md` says `assessment -> teaching_query` is the only remaining reverse edge.
  - `assessment/ports/ports.go` imports `teaching_query/ports`.
  - `assessment/application/commands/report_service.go` imports `teaching_query/contracts` and `teaching_query/ports`.
  - `app/composition/assessment_module.go` injects `teaching_query/infrastructure.NewRepository(root.DB())` directly as the assessment class insight repo.
- grill-with-docs findings:
  - No user question blocks execution. The missing requested plan filename is recoverable because the active architecture fact source and baseline name the exact edge and solution direction.
  - Class report export remains `assessment` owner; moving the whole export workflow to `teaching_query` would be a larger user-visible/reporting ownership change.
  - The adapter belongs in app composition, not `assessment`, because `assessment` must not import `teaching_query`.
- Plan adjustments after challenge:
  - Add a production import guard before implementation.
  - Keep `teaching_query` output untouched and map at the composition edge.
  - Update architecture fact source after baseline is green.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/assessment -run TestProductionAssessmentDoesNotImportTeachingQueryModule -count=1` (red before implementation, green after)
  - `cd code/backend && go test ./internal/module/assessment/application/commands -run 'TestBuildClassReportDataUsesSharedWindowedClassInsight|TestReport' -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildAssessmentModule|Test.*Assessment' -count=1`
  - `cd code/backend && go test ./internal/module/assessment/... -count=1`
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`
  - `python3 scripts/check-docs-consistency.py`
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
- Manual checks:
  - `rg 'ctf-platform/internal/module/teaching_query' code/backend/internal/module/assessment -g '*.go' -g '!**/*_test.go'` returns no production matches.
  - `rg --fixed-strings '"assessment -> teaching_query"' code/backend/internal/module/architecture_baseline_test.go` returns no matches after implementation.
- Review focus:
  - `assessment` production code must not import `teaching_query`.
  - Composition adapter must not leak `teaching_query` types into assessment packages.
  - Report JSON field names and PDF/Excel rendering helpers must stay behavior-compatible.
  - Baseline edge must be removed only after real imports disappear.

## Validation Evidence

- `cd code/backend && go test ./internal/module/assessment -run TestProductionAssessmentDoesNotImportTeachingQueryModule -count=1`
  - Red before implementation: failed on `application/commands/report_service.go` importing `teaching_query/contracts`.
  - Green after implementation: passed.
- `cd code/backend && go test ./internal/module/assessment/application/commands -run 'TestBuildClassReportDataUsesSharedWindowedClassInsight|TestReport' -count=1`: passed.
- `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - Failed once with stale `assessment -> teaching_query` baseline after the production import edge was removed.
  - Passed after removing the stale baseline entry.
- `cd code/backend && go test ./internal/app/composition -run 'TestBuildAssessmentModule|Test.*Assessment' -count=1`: passed package compile; no tests matched the run filter.
- `cd code/backend && go test ./internal/module/assessment/... -count=1`: passed.
- `python3 scripts/check-docs-consistency.py`: passed.
- `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`: passed.
- `rg 'ctf-platform/internal/module/teaching_query' code/backend/internal/module/assessment -g '*.go' -g '!**/*_test.go'`: no production matches.
- `rg --fixed-strings '"assessment -> teaching_query"' code/backend/internal/module/architecture_baseline_test.go`: no matches.
- `bash scripts/run-workflow-stage.sh completion-full`: passed.
- `bash scripts/run-workflow-stage.sh pre-commit-quick`: passed.
- `git diff --check`: passed.
- `/home/azhi/.codex/skills/development-pipeline/scripts/check_impl_plan_done.sh docs/plan/archive/impl-plan/2026-06/2026-06-10-assessment-teaching-query-reverse-dependency-implementation-plan.md`: passed; all 7 checklist items complete.

Independent review gate: not satisfied in this session because no independent reviewer/subagent tool is available. Same-context self-check found no material blocker, but it does not count as the workflow independent gate.

## Checklist

- [x] Add red architecture guard for production `assessment` not importing `teaching_query`.
- [x] Add assessment-owned class insight report types.
- [x] Change assessment ports, report service, and command tests to use assessment-owned types.
- [x] Add app composition adapter from teaching-query read model to assessment class insight port.
- [x] Remove the real production import edge and update module dependency baseline.
- [x] Update backend architecture fact source.
- [x] Run validation commands and record evidence.
