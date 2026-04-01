# Attack Evidence Review Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a teacher-facing attack evidence and replay view based on existing audit and timeline data.

**Architecture:** Reuse `audit_logs`, `instances`, `submissions`, and timeline-related read models as the evidence source. Add a teaching-readmodel query layer that aggregates attack sessions, then expose teacher-only APIs and a focused replay panel inside the existing student analysis screen. Standardize event metadata in runtime/practice audit writes instead of introducing a new event store.

**Tech Stack:** Go, Gin, GORM, Vue 3, Vite, Vitest

---

### Task 1: Define DTOs and API contracts

**Files:**
- Modify: `code/backend/internal/dto`
- Modify: `code/frontend/src/api/contracts.ts`
- Test: `code/frontend/src/api/__tests__` or relevant view tests

- [ ] Step 1: Add backend DTOs for attack session list/detail responses.
- [ ] Step 2: Add frontend contract types matching the backend DTOs.
- [ ] Step 3: Verify naming and JSON fields align with existing teacher APIs.

### Task 2: Add teaching readmodel query support

**Files:**
- Modify: `code/backend/internal/module/teaching_readmodel/ports`
- Modify: `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- Test: `code/backend/internal/module/teaching_readmodel/...`

- [ ] Step 1: Write failing repository/query tests for teacher attack sessions.
- [ ] Step 2: Implement evidence aggregation from audit logs, submissions, hints, and instances.
- [ ] Step 3: Add service-level access control and response mapping.
- [ ] Step 4: Run targeted backend tests for the new query flow.

### Task 3: Expose teacher evidence API

**Files:**
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Test: `code/backend/internal/app/full_router_state_matrix_integration_test.go` or targeted router tests

- [ ] Step 1: Add failing route/handler tests for attack session list/detail.
- [ ] Step 2: Implement handler methods and route registration.
- [ ] Step 3: Run targeted backend route tests.

### Task 4: Build teacher replay panel

**Files:**
- Modify: `code/frontend/src/api/teacher.ts`
- Modify: `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- Modify: `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- Create: `code/frontend/src/components/teacher/attack-review/*`
- Test: `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

- [ ] Step 1: Add failing frontend tests covering replay panel rendering and session switching.
- [ ] Step 2: Implement teacher attack replay API calls.
- [ ] Step 3: Implement compact session list and session detail panel.
- [ ] Step 4: Run targeted frontend tests.

### Task 5: Standardize evidence metadata in audit writes

**Files:**
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/module/practice` write paths if needed
- Test: `code/backend/internal/app/practice_flow_integration_test.go`

- [ ] Step 1: Add failing integration assertions for normalized evidence metadata.
- [ ] Step 2: Normalize audit detail keys for instance access, proxy requests, hint unlocks, and submissions.
- [ ] Step 3: Re-run the practice flow integration test.

### Task 6: Sync docs

**Files:**
- Modify: `docs/contracts/api-contract-v1.md`
- Modify: `docs/requirements/platform-gap-must-do-2026-04-01.md`
- Modify: `docs/superpowers/specs/2026-04-01-attack-evidence-review-design.md`

- [ ] Step 1: Update docs if implementation diverges from spec.
- [ ] Step 2: Run targeted backend and frontend checks together.
- [ ] Step 3: Prepare scoped commit for the attack evidence feature.
- Modify: `docs/architecture/backend/04-api-design.md`
- Modify: `docs/architecture/frontend/04-api-layer.md`

- [ ] Step 1: Document the new teacher evidence review API and response structure
- [ ] Step 2: Document frontend consumption path
- [ ] Step 3: Re-run any affected documentation examples mentally against implementation
