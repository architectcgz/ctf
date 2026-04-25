# Community Writeup And Recommended Solutions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the student writeup review workflow with a LeetCode-style challenge solutions area that supports recommended solutions, community writeups, and a single per-student writeup draft/publish flow gated by solve status.

**Architecture:** Keep `challenge_writeups` as the official-solution source and reuse `submission_writeups` as the community-writeup store, but migrate its fields from review semantics to publishing and moderation semantics. Expose a unified challenge-solution read path for the student challenge page, add recommendation and moderation actions for teacher/admin operators, and remove review-only UI and APIs from teacher analysis surfaces.

**Tech Stack:** Go, Gin, GORM, PostgreSQL migrations, Vue 3, Vite, Vitest, Vue Test Utils

---

### Task 1: Migrate writeup models from review semantics to community semantics

**Files:**
- Create: `code/backend/migrations/000049_convert_submission_writeups_to_community_mode.up.sql`
- Create: `code/backend/migrations/000049_convert_submission_writeups_to_community_mode.down.sql`
- Modify: `code/backend/internal/model/submission_writeup.go`
- Modify: `code/backend/internal/model/topology.go`
- Test: `code/backend/internal/app/full_router_state_matrix_integration_test.go`

- [ ] **Step 1: Write the failing integration assertions for the new writeup schema behavior**

Add assertions in `code/backend/internal/app/full_router_state_matrix_integration_test.go` that expect:
- student writeups to use `draft` / `published`
- no `review_comment` / `reviewed_at` fields in responses
- official writeups to expose recommendation metadata when present

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/app -run FullRouterStateMatrix -count=1
```

Expected: FAIL on missing fields, status values, or route behavior.

- [ ] **Step 2: Add the migration files**

Implement migration `000049_*` to:
- convert `submission_status='submitted'` to `published`
- drop `review_status`, `reviewed_by`, `reviewed_at`, `review_comment`
- add `visibility_status`, `is_recommended`, `recommended_at`, `recommended_by`, `published_at`
- add official writeup recommendation fields to `challenge_writeups`
- backfill community writeups to `visibility_status='visible'` and `is_recommended=false`

- [ ] **Step 3: Update backend models to match the migrated schema**

In `code/backend/internal/model/submission_writeup.go`:
- rename constants to community semantics
- remove review constants and review-only fields
- add visibility and recommendation fields

In `code/backend/internal/model/topology.go`:
- add recommendation fields to `ChallengeWriteup`

- [ ] **Step 4: Re-run the router state matrix test**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/app -run FullRouterStateMatrix -count=1
```

Expected: still failing, but now on DTOs, repositories, or handlers rather than schema mismatch.

### Task 2: Replace DTOs, mappings, and repository queries with community/recommendation contracts

**Files:**
- Modify: `code/backend/internal/dto/writeup.go`
- Modify: `code/backend/internal/module/challenge/domain/mappers.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- Test: `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`
- Test: `code/backend/internal/module/challenge/application/commands/writeup_topology_service_test.go`
- Test: `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`

- [ ] **Step 1: Write failing backend unit tests for community-writeup and recommended-writeup DTOs**

Add or update tests to expect:
- student writeup responses with publishing/moderation fields
- teacher/admin list items without review-only fields
- official writeups with recommendation metadata

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/challenge/application/commands -run Writeup -count=1
```

Expected: FAIL because DTOs and mappers still expose review semantics.

- [ ] **Step 2: Update DTOs and mappers**

In `code/backend/internal/dto/writeup.go`:
- rename request/response semantics from review to community moderation
- add DTOs for:
  - recommended solution list items
  - community solution list items
  - teacher/admin moderation actions

In `code/backend/internal/module/challenge/domain/mappers.go`:
- remove review-only mapping logic
- add mapping for recommendation metadata and visibility/moderation state

- [ ] **Step 3: Update repository contracts and SQL projections**

In `code/backend/internal/module/challenge/ports/ports.go` and `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`:
- replace review-based list/query methods with community-writeup list/detail methods
- add official/community recommendation lookups
- add visibility filtering for public community listings
- preserve `user_id + challenge_id` uniqueness

- [ ] **Step 4: Re-run the challenge writeup unit tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/challenge/application/commands -run Writeup -count=1
go test ./internal/module/challenge/application/queries -run Writeup -count=1
```

Expected: FAIL only on service behavior that has not yet been updated.

### Task 3: Rewrite challenge writeup services and routes around publishing, moderation, and recommendations

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- Modify: `code/backend/internal/module/challenge/api/http/writeup_handler.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`
- Test: `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- Test: `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`

- [ ] **Step 1: Write failing route and command tests for the new API surface**

Add coverage for:
- `GET /api/v1/challenges/:id/solutions/recommended`
- `GET /api/v1/challenges/:id/solutions/community`
- updated `POST /api/v1/challenges/:id/writeup-submissions`
- moderation routes for teacher/admin recommendation and hide/restore actions
- removal of `PUT /api/v1/teacher/writeup-submissions/:id/review`

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/app -run 'Router|FullRouterStateMatrix' -count=1
```

Expected: FAIL because routes and handlers are still review-based.

- [ ] **Step 2: Implement the new command behavior**

In `code/backend/internal/module/challenge/application/commands/writeup_service.go`:
- treat student writeups as `draft` or `published`
- allow draft save before solve
- allow publish only after solve
- support teacher/admin hide/restore
- support teacher/admin recommend/unrecommend for community and official solutions

- [ ] **Step 3: Implement the new query behavior**

In `code/backend/internal/module/challenge/application/queries/writeup_service.go`:
- return official writeup for admin as before
- return recommended/community lists for solved users only
- return locked/forbidden behavior for unsolved users
- return the current user's writeup regardless of solve state

- [ ] **Step 4: Update handler methods and route registration**

In `code/backend/internal/module/challenge/api/http/writeup_handler.go` and `code/backend/internal/app/router_routes.go`:
- remove review handler methods
- add list and moderation handlers
- keep official writeup CRUD for admin routes

- [ ] **Step 5: Re-run the router and writeup test suites**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/challenge/application/commands -run Writeup -count=1
go test ./internal/module/challenge/application/queries -run Writeup -count=1
go test ./internal/app -run 'Router|FullRouterStateMatrix' -count=1
```

Expected: PASS for the updated challenge writeup service and route behavior.

### Task 4: Update frontend contracts and student challenge API consumption

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/challenge.ts`
- Test: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- Test: `code/frontend/src/api/__tests__/*.test.ts`

- [ ] **Step 1: Write failing frontend tests for the new solution payloads**

Update challenge detail tests to expect:
- a three-tab solutions module for solved users
- locked state for unsolved users
- my writeup using `draft` / `published` plus hidden messaging
- recommendation and community list rendering

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/challenges/__tests__/ChallengeDetail.test.ts
```

Expected: FAIL because the page still assumes official writeup plus review-based student writeup data.

- [ ] **Step 2: Update frontend contracts**

In `code/frontend/src/api/contracts.ts`:
- replace review-oriented writeup types with community/moderation types
- add:
  - recommended solution item type
  - community solution item/detail type
  - moderation status types

- [ ] **Step 3: Update challenge API helpers**

In `code/frontend/src/api/challenge.ts`:
- add recommended/community solution fetchers
- update student writeup upsert payloads to `draft` / `published`
- remove review-only normalization

- [ ] **Step 4: Re-run the student challenge tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/challenges/__tests__/ChallengeDetail.test.ts
```

Expected: still failing, but now on actual UI rendering rather than API type mismatch.

### Task 5: Rebuild the challenge detail page into a LeetCode-style solutions area

**Files:**
- Modify: `code/frontend/src/views/challenges/ChallengeDetail.vue`
- Modify: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- Possibly create: `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
- Possibly create: `code/frontend/src/components/challenge/ChallengeSolutionList.vue`
- Possibly create: `code/frontend/src/components/challenge/MyChallengeWriteupEditor.vue`

- [ ] **Step 1: Extend the failing UI tests before implementation**

Cover:
- unsolved users see counts and lock messaging only
- solved users see `推荐题解 / 社区题解 / 我的题解`
- recommended solution cards render official and community sources
- community list search/sort basics
- my writeup editor supports save draft, publish, and hidden-state notice

- [ ] **Step 2: Extract or implement the solutions panel UI**

Prefer extracting the solution area into focused components if `ChallengeDetail.vue` becomes too large. Keep the existing challenge description, flag submission, and hint sections intact.

- [ ] **Step 3: Implement the solve-gated display logic**

Use `challenge.is_solved` to:
- show the lock state when unsolved
- fetch and render recommended/community lists only when solved
- continue allowing draft editing before solve if the backend permits it

- [ ] **Step 4: Replace the old official-writeup button flow**

Remove the standalone `查看题解` toggle and fold official content into the recommended solutions tab.

- [ ] **Step 5: Re-run the challenge detail test suite**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/challenges/__tests__/ChallengeDetail.test.ts
```

Expected: PASS.

### Task 6: Extend admin writeup management with recommendation and community moderation tools

**Files:**
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/components/admin/writeup/ChallengeWriteupEditorPage.vue`
- Modify: `code/frontend/src/composables/useChallengeWriteupEditorPage.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ChallengeWriteup.test.ts`

- [ ] **Step 1: Write failing admin UI tests**

Cover:
- official writeup recommendation toggle
- community writeup management list rendering
- hide/restore/recommend actions dispatching correctly

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/admin/__tests__/ChallengeWriteup.test.ts
```

Expected: FAIL because the current page only edits the official writeup.

- [ ] **Step 2: Add admin API methods**

In `code/frontend/src/api/admin.ts`:
- add official writeup recommend/unrecommend calls
- add community writeup list and moderation calls for the admin screen

- [ ] **Step 3: Extend the admin writeup page and composable**

In `code/frontend/src/components/admin/writeup/ChallengeWriteupEditorPage.vue` and `code/frontend/src/composables/useChallengeWriteupEditorPage.ts`:
- keep the existing official editor
- add a community writeup management section
- surface recommendation and moderation actions

- [ ] **Step 4: Re-run the admin writeup tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/admin/__tests__/ChallengeWriteup.test.ts
```

Expected: PASS.

### Task 7: Remove teacher review UI and replace it with read-only community writeup status

**Files:**
- Modify: `code/frontend/src/api/teacher.ts`
- Modify: `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- Modify: `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`

- [ ] **Step 1: Write failing teacher UI tests**

Cover:
- no review actions or review-comment editor for student writeups
- student analysis shows publish/hidden/recommended summaries instead
- review archive no longer expects writeup review fields

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
```

Expected: FAIL because teacher views still expect review semantics.

- [ ] **Step 2: Update teacher API types and fetchers**

In `code/frontend/src/api/teacher.ts`:
- remove review request methods for writeups
- update teacher writeup list payloads to community-state fields only

- [ ] **Step 3: Simplify teacher analysis surfaces**

In `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue` and `code/frontend/src/components/teacher/StudentInsightPanel.vue`:
- remove review actions
- keep writeup visibility and recommendation summaries
- preserve manual-review answer flow for flag submissions, which is unrelated

- [ ] **Step 4: Re-run teacher tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
```

Expected: PASS.

### Task 8: Final verification and doc sync

**Files:**
- Modify: `docs/architecture/features/2026-04-04-community-writeup-and-recommended-solutions-design.md`
- Modify: `docs/contracts/api-contract-v1.md`
- Modify: `docs/architecture/backend/04-api-design.md`
- Modify: `docs/architecture/frontend/04-api-layer.md`

- [ ] **Step 1: Update docs if implementation diverges from the approved spec**

Keep the spec and contract docs aligned with the implemented route names and payload shapes.

- [ ] **Step 2: Run targeted backend verification**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/challenge/application/commands -run Writeup -count=1
go test ./internal/module/challenge/application/queries -run Writeup -count=1
go test ./internal/app -run 'Router|FullRouterStateMatrix' -count=1
```

Expected: PASS.

- [ ] **Step 3: Run targeted frontend verification**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm exec vitest run \
  src/views/challenges/__tests__/ChallengeDetail.test.ts \
  src/views/admin/__tests__/ChallengeWriteup.test.ts \
  src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts \
  src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts
npm exec eslint \
  src/views/challenges/ChallengeDetail.vue \
  src/components/admin/writeup/ChallengeWriteupEditorPage.vue \
  src/composables/useChallengeWriteupEditorPage.ts \
  src/views/teacher/TeacherStudentAnalysis.vue \
  src/components/teacher/StudentInsightPanel.vue \
  src/api/challenge.ts \
  src/api/admin.ts \
  src/api/teacher.ts \
  src/api/contracts.ts
```

Expected: PASS, except for any pre-existing sanitized `v-html` warnings already accepted by the repo.

- [ ] **Step 4: Prepare the final scoped commits**

Suggested commit breakdown:

```bash
git add code/backend/migrations code/backend/internal/model code/backend/internal/dto \
  code/backend/internal/module/challenge code/backend/internal/app
git commit -m "feat(writeup): convert writeup review flow to community solutions"

git add code/frontend/src/api code/frontend/src/views/challenges \
  code/frontend/src/components/challenge code/frontend/src/views/admin \
  code/frontend/src/components/admin/writeup code/frontend/src/composables \
  code/frontend/src/views/teacher code/frontend/src/components/teacher
git commit -m "feat(frontend): add community and recommended challenge solutions"

git add docs/contracts docs/architecture docs/superpowers/specs
git commit -m "docs: sync community writeup and solution APIs"
```
