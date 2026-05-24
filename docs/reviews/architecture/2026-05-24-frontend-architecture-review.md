# 2026-05-24 CTF Frontend Architecture Review

## Review Target

- Repository: `ctf`
- Scope: `code/frontend/src`
- Focus: route view ownership, feature boundaries, admin/teacher coupling, request/error policy, residual legacy structure
- Review prompt asset: `harness/prompts/frontend-architecture-review.md`
- Files reviewed:
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/api/request.ts`
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts`
  - `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts`
  - `code/frontend/src/features/platform-users/model/usePlatformUsers.ts`
  - `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
  - representative views, widgets, and shared composables under `code/frontend/src/views`, `features`, `components`, `composables`

## Classification Check

- Classification: architecture review
- Decision: agree with non-trivial review classification

## Gate Verdict

- Verdict: blocked

## Findings

### P1-1 Allowlist-driven architecture means the target structure is still not the actual structure

- Location:
  - `code/frontend/src/__tests__/architectureAllowlist.ts:3-180`
  - duplicate examples:
    - `code/frontend/src/composables/useImageManagePage.ts`
    - `code/frontend/src/features/image-management/model/useImageManagePage.ts`
    - `code/frontend/src/composables/usePlatformUsers.ts`
    - `code/frontend/src/features/platform-users/model/usePlatformUsers.ts`
- Evidence:
  - `architectureAllowlist.ts` currently hard-codes at least:
    - 49 component -> feature exceptions
    - 49 feature -> router exceptions
    - 18 legacy `*Page.vue` component exceptions
    - 10 widget -> legacy component exceptions
    - 6 oversized view exceptions
  - Root-level `src/composables/` and `src/features/*/model/` still keep 54 same-named hooks in parallel.
- Risk:
  - Boundary tests are mostly freezing historical exceptions instead of enforcing the intended end state.
  - New contributors have to guess whether the real owner is `views`, `components/*Page.vue`, `composables`, or `features/*/model`.
  - Search cost and refactor blast radius stay high because legacy and new entry points coexist on disk.
- Recommendation:
  - Convert the migration target from directory convention into a real convergence plan.
  - Pick one canonical owner per workflow:
    - route state and page workflow: `features/*/model`
    - presentational containers: `widgets` or `components`
    - generic behavior only: `composables`
  - For each feature family, delete or deprecate the legacy twin after the new model is adopted.
  - Reduce allowlists in slices, not by one global cleanup pass.

### P1-2 Admin workflows are still coupled to teacher views and teacher APIs

- Location:
  - `code/frontend/src/router/routes/platformRoutes.ts:41-138`
  - `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts:4-22`
  - `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts:4-18`
  - `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts:4-6`
- Evidence:
  - Multiple `/platform/*` routes render teacher views directly, for example:
    - `PlatformClassStudents -> @/views/teacher/TeacherClassStudents.vue`
    - `PlatformClassTrend -> @/views/teacher/TeacherClassWorkspaceSection.vue`
    - `PlatformStudentAnalysis -> @/views/teacher/TeacherStudentAnalysis.vue`
    - `PlatformAwdReviewDetail -> @/views/teacher/TeacherAWDReviewDetail.vue`
  - Admin page models call `@/api/teacher` endpoints directly.
- Risk:
  - Admin surface behavior now depends on teacher-side DTO shape, route assumptions, and page composition decisions.
  - A teacher feature change can silently regress admin behavior even when admin requirements diverge.
  - Permission boundaries become structural afterthoughts instead of first-class module boundaries.
- Recommendation:
  - Keep shared widgets or feature modules where the workflow is genuinely identical, but stop routing admin pages directly to teacher route views.
  - Extract role-neutral workspace features, then let `/academy/*` and `/platform/*` compose them with separate page owners.
  - Rename the current `teacher`-owned query helpers or move them under neutral feature names before more admin cases reuse them.

### P1-3 `platform-users` is an over-broad bucket that mixes unrelated admin capabilities

- Location:
  - `code/frontend/src/features/platform-users/model/index.ts:1-6`
  - `code/frontend/src/features/platform-users/model/usePlatformUsers.ts:1-177`
  - `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts:1-164`
  - `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts:1-238`
  - `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts:1-88`
- Evidence:
  - One feature namespace currently owns:
    - user CRUD and import
    - class list management
    - student directory management
    - instance management
  - These flows do not share the same API, permissions, state model, or future change cadence.
- Risk:
  - Feature ownership is ambiguous: a maintainer touching “platform users” may unintentionally widen coupling across users, classes, students, and instances.
  - The module name stops being useful as an architectural boundary and becomes a dumping area for admin pages.
- Recommendation:
  - Split by use case rather than by an umbrella admin noun.
  - A low-risk path is:
    - keep `platform-users` only for user governance
    - move class/student directory to a dedicated admin teaching workspace feature
    - move instance management to its own platform instance feature

### P1-4 The request layer owns navigation-level error policy, so page owners cannot fully control failure UX

- Location:
  - `code/frontend/src/api/request.ts:105-180`
  - representative page owners expecting local recovery:
    - `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts:10-34`
    - `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts:28-73`
- Evidence:
  - The axios response interceptor directly redirects on `429`, `401`, and generic HTTP failures.
  - Page models also catch errors locally and try to show inline fallback state.
- Risk:
  - Ownership is split between transport and page layer.
  - A request that should stay recoverable inside a page can still force a global route jump.
  - This makes retries, drafts, and page-specific error UX hard to reason about and hard to test.
- Recommendation:
  - Keep transport concerns in `request.ts`: error normalization, auth/session handling, cancellation, request IDs.
  - Move navigation decisions to the page or feature owner, except for a very small set of truly global auth/session failures.
  - If some statuses must stay global, declare them explicitly and keep the rest local.

### P1-5 Instance management does client-side filtering and pagination over a whole dataset

- Location:
  - `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts:35-100`
- Evidence:
  - `loadInstances()` loads one full list from `getTeacherInstances(...)`.
  - Keyword filtering, status grouping, counts, and pagination all happen in computed state on the client.
- Risk:
  - The page will degrade as instance count grows.
  - Admin-side list semantics become inconsistent with server truth if backend later adds paging, sorting, or partial loading.
  - Destroying one item mutates a local snapshot instead of working against a stable query contract.
- Recommendation:
  - Introduce a server-owned query contract for platform instance directory: `keyword`, `status`, `page`, `page_size`, and summary counters.
  - Keep the current UI, but move filtering/paging ownership down to the API query layer.

### P2-1 Runtime `console.log` and broad `console.error` usage still leak through production code

- Location:
  - `code/frontend/src/features/auth/model/useLoginPage.ts:18-55,72`
  - representative files:
    - `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts:19`
    - `code/frontend/src/features/platform-users/model/usePlatformUsers.ts:69-75`
    - `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts:101-107,126-131`
- Risk:
  - Login page prints themed console output on every mount.
  - Error handling across the app relies heavily on ad hoc console logging, which is noisy locally and not useful as structured observability.
- Recommendation:
  - Remove decorative runtime logs from the login flow.
  - Replace raw `console.error` scatter with a small reporting utility or keep it strictly in development-only helpers.

## Material Findings

- P1-1 Allowlist-driven architecture debt is still the effective architecture.
- P1-2 Admin and teacher surfaces are structurally coupled.
- P1-3 `platform-users` is over-broad and should not keep accumulating responsibilities.
- P1-4 Request interception owns page navigation for recoverable failures.
- P1-5 Platform instance management will not scale with current client-side query ownership.

## Senior Implementation Assessment

- The repo is already moving in the right direction: route views are thin, most new page workflows live under `features/*/model`, and there are dedicated boundary tests.
- The remaining risk is not “lack of architecture effort”, but incomplete convergence. The project currently has four overlapping ownership layers:
  - route views
  - legacy `components/*Page.vue`
  - root `composables/use*.ts`
  - feature-scoped `model/use*.ts`
- A lower-risk path is not a wholesale FSD rewrite. It is a convergence plan:
  - freeze new legacy entry points
  - split the known over-broad buckets
  - move admin/teacher shared logic under neutral feature owners
  - shrink allowlists by feature slice
  - keep route views as composition-only surfaces

## Required Re-Validation

- After fixes, re-check:
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - targeted tests for any split admin/teacher page workflows
  - a code search proving removed duplicate hook entry points are no longer referenced

## Residual Risk

- This review focused on frontend architecture and runtime ownership. It did not fully re-audit every visual component or every API wrapper.
- Some duplicate `src/composables/use*.ts` files may still be referenced by tests or historical tooling even when runtime code no longer imports them.
- The backend contract ownership behind `@/api/teacher` vs admin-facing routes was inferred from frontend call sites and route structure, not from backend module review in this pass.

## Touched Known-Debt Status

- Known debt touched by review only, not fixed in this pass.
- Current status: blocked until the touched architecture debt is converged slice by slice instead of remaining permanently allowlisted.
