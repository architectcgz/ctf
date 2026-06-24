<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# teaching-read-model-module-rename Implementation Plan

**Goal:** Rename the backend module `teaching_query` to `teaching_analysis` so the module name matches the repository's business-owner naming style instead of an implementation-style query name.

**Architecture:** This is a structural rename across the backend module directory, the app composition entrypoint, architecture guardrails, and current architecture docs. The external HTTP contracts stay unchanged; only internal module identifiers, import paths, wiring symbols, and current-fact documentation are renamed to the new module owner term.

**Tech Stack:** Go backend, Gin, GORM, architecture guard tests, code-workflow plan/review gates.

---

## Task Metadata

- Task Slug: `2026-06-11-teaching-read-model-module-rename`
- Started At: `2026-06-11T15:29:18Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-teaching-read-model-module-rename`
- Branch: `task/2026-06-11-teaching-read-model-module-rename`

## Objective And Non-Goals

- Objective:
  - Rename `code/backend/internal/module/teaching_query` to `code/backend/internal/module/teaching_analysis`.
  - Rename app composition entrypoints and route wiring symbols from `TeachingQuery*` / `teachingQuery*` to `TeachingAnalysis*` / `teachingAnalysis*`.
  - Update active architecture facts and project pattern references so they describe the new module name consistently.
- Non-Goals:
  - Do not change any external HTTP route, response DTO, permission rule, or read-model behavior.
  - Do not redesign the split between directory query, overview, class insight, and student review services.
  - Do not rewrite archived implementation plans or history-only docs under `docs/plan/archive/`.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/features/教师教学概览聚合架构.md`
  - `docs/architecture/features/教学复盘优化设计.md`
  - `docs/architecture/features/攻击证据链与教学复盘架构.md`
  - `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
  - `docs/architecture/features/教学复盘建议生成架构.md`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-05/2026-05-24-admin-teaching-query-owner-decoupling-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-assessment-teaching-query-boundary-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - The rename touches a protected backend module boundary, app composition wiring, architecture guard tests, and current-fact docs.
  - The work includes key file and directory moves, broad import-path updates, and consistency changes across multiple packages.

## Files

- Create:
  - `code/backend/internal/app/composition/teaching_analysis_module.go`
- Modify:
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_user_teacher_routes.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `code/backend/internal/app/router_route_wiring_test.go`
  - `code/backend/internal/app/full_router_integration_test.go`
  - `code/backend/internal/app/full_router_module_wiring_test.go`
  - `code/backend/internal/app/composition/assessment_module.go`
  - `code/backend/internal/app/composition/assessment_class_insight_adapter.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/mapper_policy_test.go`
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/cmd/seed-teaching-review-data/main.go`
  - `code/backend/cmd/seed-teaching-review-data/main_test.go`
  - `code/backend/internal/module/teaching_analysis/**`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/02-database-design.md`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/features/教师教学概览聚合架构.md`
  - `docs/architecture/features/教学复盘优化设计.md`
  - `docs/architecture/features/攻击证据链与教学复盘架构.md`
  - `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
  - `docs/architecture/features/教学复盘建议生成架构.md`
  - `docs/architecture/features/校园级CTF-AWD模式完整设计.md`
  - `docs/architecture/features/赛事导出与复盘归档架构.md`
  - `harness/policies/project-patterns.yaml`
- Review:
  - Renamed module path and composition symbol coverage
  - Assessment anti-dependency guard updated to the new prefix
  - Active docs only; archived plans intentionally unchanged
- Test:
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `code/backend/internal/module/teaching_analysis/architecture_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/app/router_route_wiring_test.go`
  - `code/backend/internal/module/assessment/architecture_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Existing module naming in `code/backend/internal/module/*`
  - Current module map in `docs/architecture/backend/07-modular-monolith-refactor.md`
  - Teaching-side feature docs under `docs/architecture/features/*`
- Reuse / extend / split / create-new decision:
  - Reuse the existing module structure and service split; rename only the module owner term and its wiring/documentation references.
  - Do not create a second compatibility alias module or duplicate composition wrapper.
- Owner boundary:
  - The module continues to own teacher-facing cross-owner analysis aggregation.
  - `internal/teaching/*` remains the shared rule/evidence kernel and is not merged into the module rename.
- Why this is the narrowest safe surface:
  - It fixes the naming mismatch without changing any API or behavior, but still updates every current-fact owner boundary that would otherwise drift.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task is a structural rename where the main risk is choosing the wrong owner term, not implementation mechanics.
- grill-with-docs findings:
  - `teaching_read_model` is not acceptable because it describes implementation shape, not business owner.
  - Current docs consistently describe the module as teacher-facing teaching analysis and review aggregation; `teaching_analysis` is the narrowest business-owner name that still covers overview, class insight, student review, evidence, and attack sessions.
  - `teaching_review` and `teaching_insight` are too narrow for the existing owner surface.
- Plan adjustments after challenge:
  - Rename both the backend module directory and the app composition entrypoint to `teaching_analysis`.
  - Update current architecture docs and project pattern references in the same slice instead of leaving naming drift behind.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/app -run 'TestTeachingAnalysisModuleContractsCompile|TestTeachingAnalysisModuleUsesTypedDeps|TestTeacherRoutesAreServedByTeachingAnalysis' -count=1`
  - `cd code/backend && go test ./internal/module/teaching_analysis -count=1`
  - `cd code/backend && go test ./internal/module/assessment -run 'TestAssessmentRuntimeCodeDoesNotDependOnTeachingAnalysis|TestProductionAssessmentDoesNotImportTeachingAnalysisModule' -count=1`
- Manual checks:
  - Confirm no active code/docs/harness reference `teaching_query` outside archived plans or historical records.
- Review focus:
  - Missing import-path replacements
  - Composition symbol drift (`TeachingQuery*` leftover)
  - Current-fact docs still naming the old module

## Execution Checklist

- [x] Confirm `teaching_analysis` as the business-owner module name and keep external HTTP contracts unchanged.
- [x] Rename backend module directory and app composition entrypoint from `teaching_query` to `teaching_analysis`.
- [x] Update imports, route wiring, architecture guards, system test helpers, and seed command references to the new module path.
- [x] Sync active architecture docs and project pattern references to the new module name without rewriting archived plans.
- [x] Run `gofmt` on affected Go files and execute the planned targeted `go test` commands.
