<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 前端架构边界修复 Implementation Plan

**Goal:** Fix the frontend route/feature owner drift identified by the architecture review while preserving user-visible UI and runtime behavior.

**Architecture:** Keep `pages/**` as route entrypoints and move page-sized UI ownership into feature slices. Platform-only capabilities should live under `features/platform/*`; top-level `features/*` remains for student/shared/cross-role capabilities.

**Tech Stack:** Vue 3, Vite, TypeScript, Feature-Sliced-style frontend layers, Vitest architecture guards.

---

## Task Metadata

- Task Slug: `2026-06-09-frontend-architecture-boundary-fix`
- Started At: `2026-06-09T04:21:23Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-frontend-architecture-boundary-fix`
- Branch: `task/2026-06-09-frontend-architecture-boundary-fix`

## Objective And Non-Goals

- Objective:
  - Move student challenge-list page shell from the route view into `features/challenge-list/ui` without changing DOM/class semantics.
  - Move confirmed platform-only feature directories from top-level `features/*` into `features/platform/*`.
  - Update active source imports, active frontend tests, and frontend architecture policy references.
- Non-Goals:
  - Do not redesign visible UI, copy, spacing, color, or interaction behavior.
  - Do not rewrite the broad `?raw` style-test system in this task.
  - Do not edit archived historical plans/reviews except where an active test or policy needs current paths.
  - Do not change API contracts, router paths, role permissions, or backend code.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/architecture/frontend/01-architecture-overview.md`
  - `docs/architecture/frontend/06-components.md`
  - `docs/architecture/frontend/07-pages-dataflow.md`
- Related architecture/contracts:
  - `code/frontend/scripts/frontend-architecture-policy.json`
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- Related prior work:
  - Earlier feature UI migrations in `docs/plan/archive/impl-plan/2026-05/*challenge-topology-studio*`
  - Earlier feature UI migrations in `docs/plan/archive/impl-plan/2026-05/*contest-projector*`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Moves feature directories and route-view ownership boundaries.
  - Touches architecture guard paths, active tests, and import paths across several frontend slices.
  - Requires completion validation plus independent review gate.

## Files

- Create:
  - `code/frontend/src/features/challenge-list/ui/ChallengeListPage.vue`
- Modify:
  - `code/frontend/src/pages/challenges/ChallengeListRoutePage.vue`
  - `code/frontend/src/features/challenge-list/index.ts`
  - `code/frontend/src/features/challenge-list/ui/index.ts`
  - `code/frontend/src/features/platform/{image-management,audit-log,contest-projector,contest-awd-config,challenge-topology-studio}/**`
  - Active source/test imports under `code/frontend/src/**`
  - `code/frontend/scripts/frontend-architecture-policy.json`
  - Relevant active frontend architecture docs if path facts become stale.
- Review:
  - `code/frontend/src/pages/**/__tests__/*`
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- Test:
  - `bash scripts/check-frontend-test-guard.sh --files <changed frontend tests>`
  - `bash scripts/check-frontend-architecture.sh`
  - Focused Vitest files for moved slices and challenge list route/page tests.

## 复用与 Owner 决策

- Existing patterns searched:
  - Route pages already import feature page shells through public APIs, e.g. platform challenge detail and overview pages.
  - Platform namespace already contains `features/platform/overview`, `features/platform/challenges`, `features/platform/contest-manage`, `features/platform/instance-management`.
- Reuse / extend / split / create-new decision:
  - Reuse existing feature UI shells and public API pattern.
  - Create only `ChallengeListPage.vue` as a direct extraction of existing route page template/style/script.
  - Move platform-only directories physically instead of leaving wrapper aliases, because wrappers would preserve the owner drift.
- Owner boundary:
  - `pages/challenges/ChallengeListRoutePage.vue`: route entry only.
  - `features/challenge-list/ui/ChallengeListPage.vue`: challenge-list page shell and presentational layout.
  - `features/challenge-list/model/useChallengeListPage.ts`: query/filter/loading/error/list owner remains unchanged.
  - `features/platform/*`: platform-only capabilities and their model/ui/tests.
- Why this is the narrowest safe surface:
  - Keeps runtime behavior and visual DOM/class semantics intact.
  - Does not combine this with test-system redesign or archived-doc cleanup.
  - Updates active imports and guards so future code follows the corrected owner path.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - This is structural frontend cleanup, not a runtime bug. The key decision is owner boundary and migration scope.
- grill-with-docs findings:
  - Existing docs already say route pages are thin composition entries and platform-only features belong under `features/platform/*`.
  - Several top-level feature directories are confirmed platform-only by route callers and API usage.
  - The `?raw` test cluster is a real maintenance risk, but fixing it broadly would exceed this owner-boundary slice.
- Plan adjustments after challenge:
  - Limit platform namespace migration to five confirmed active platform-only slices.
  - Keep archived historical docs untouched unless active checks depend on the path.
  - Treat challenge-list extraction as DOM/class-preserving move, not UI redesign.

## Validation

- Commands:
  - `git diff --check`
  - `bash scripts/check-frontend-test-guard.sh --files <changed frontend tests>`
  - `bash scripts/check-frontend-architecture.sh`
  - Focused `npm run test:run -- ...` for moved-slice tests if needed after path updates.
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - Inspect route page diff for DOM/class/CSS semantic preservation.
  - Inspect imports to confirm no active source still imports moved platform-only slices from top-level `features/*`.
- Review focus:
  - No visual DOM/class drift in challenge list page shell.
  - No wrapper alias hiding platform owner drift.
  - Active architecture guards updated to current paths.
  - No broad archived-doc churn.

## Execution Checklist

- [x] Move challenge-list page shell into `features/challenge-list/ui/ChallengeListPage.vue`.
- [x] Reduce `pages/challenges/ChallengeListRoutePage.vue` to a route entry importing the feature page shell.
- [x] Move platform-only feature directories under `features/platform/*`.
- [x] Update active import paths and architecture policy references.
- [x] Run focused frontend test guard and architecture checks.
- [x] Run completion-full.
- [x] Complete independent review gate and fix findings if needed.
