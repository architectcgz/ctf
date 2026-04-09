# Instance Sharing Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add challenge-level instance sharing strategies so the platform can reuse launchable challenge instances as `per_user`, `per_team`, or `shared` instead of always scaling container count with user count.

**Architecture:** Keep challenge metadata as the source of truth for allocation strategy, persist the resolved scope on each instance record, and reuse the existing practice/runtime orchestration by extending scope resolution, visibility queries, and management rules. Shared instances are restricted to challenge types that do not rely on per-instance dynamic flag injection.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, Vue 3, TypeScript, Vitest

---

### Task 1: Lock the behavior with failing backend tests

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- Modify: `code/backend/internal/module/runtime/service_test.go`

- [ ] **Step 1: Add failing challenge tests for invalid `shared + dynamic` combinations**
- [ ] **Step 2: Add failing practice tests for shared practice reuse and non-AWD team reuse**
- [ ] **Step 3: Add failing runtime tests for shared instance visibility and forbidden destroy/extend**
- [ ] **Step 4: Run focused tests and confirm RED**

### Task 2: Add instance sharing metadata to challenge and instance models

**Files:**
- Modify: `code/backend/internal/model/challenge.go`
- Modify: `code/backend/internal/model/instance.go`
- Create: `code/backend/migrations/000043_add_instance_sharing_to_challenges_and_instances.up.sql`
- Create: `code/backend/migrations/000043_add_instance_sharing_to_challenges_and_instances.down.sql`

- [ ] **Step 1: Define model constants for `per_user`, `per_team`, and `shared`**
- [ ] **Step 2: Add `instance_sharing` to challenges and `share_scope` to instances**
- [ ] **Step 3: Replace active-instance unique indexes with scope-aware unique indexes**
- [ ] **Step 4: Run migration-sensitive tests**

### Task 3: Wire challenge CRUD and query responses

**Files:**
- Modify: `code/backend/internal/dto/challenge.go`
- Modify: `code/backend/internal/module/challenge/domain/mappers.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service.go`

- [ ] **Step 1: Add DTO fields for `instance_sharing`**
- [ ] **Step 2: Validate unsupported shared configurations in challenge create/update**
- [ ] **Step 3: Expose `instance_sharing` in admin and student-facing challenge responses**
- [ ] **Step 4: Run focused challenge module tests**

### Task 4: Extend practice scope resolution and startup reuse rules

**Files:**
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/infrastructure/repository.go`

- [ ] **Step 1: Extend `InstanceScope` with resolved share mode**
- [ ] **Step 2: Resolve challenge launch scope from challenge config + contest context**
- [ ] **Step 3: Reuse shared instances and refresh TTL on repeated shared starts**
- [ ] **Step 4: Persist resolved `share_scope` on created instances**
- [ ] **Step 5: Run focused practice module tests**

### Task 5: Update runtime visibility, access control, and management restrictions

**Files:**
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Modify: `code/backend/internal/module/runtime/application/queries/instance_service.go`
- Modify: `code/backend/internal/module/runtime/application/commands/instance_service.go`
- Modify: `code/backend/internal/module/runtime/ports/http.go`
- Modify: `code/backend/internal/dto/instance.go`

- [ ] **Step 1: Make shared practice instances visible to users and shared contest instances visible to registered contestants**
- [ ] **Step 2: Return `share_scope` in instance responses**
- [ ] **Step 3: Reject user destroy/extend for `shared` instances**
- [ ] **Step 4: Run focused runtime module tests**

### Task 6: Sync minimal frontend behavior for shared instances

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/challenge.ts`
- Modify: `code/frontend/src/api/instance.ts`
- Modify: `code/frontend/src/composables/useChallengeInstance.ts`
- Modify: `code/frontend/src/components/challenge/ChallengeInstanceCard.vue`
- Modify: `code/frontend/src/components/common/InstancePanel.vue`
- Modify: `code/frontend/src/api/__tests__/instance.test.ts`
- Modify: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- Modify: `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`

- [ ] **Step 1: Add contract fields for `instance_sharing` and `share_scope`**
- [ ] **Step 2: Hide extend/destroy actions for shared instances**
- [ ] **Step 3: Keep start/open flow intact for shared instances**
- [ ] **Step 4: Run focused frontend tests**

### Task 7: Validation and doc sync

**Files:**
- Modify: `code/docs/tasks/2026-04-09-instance-sharing-design.md`
- Modify: `docs/contracts/openapi-v1.yaml`

- [ ] **Step 1: Run backend verification**
  Run: `cd code/backend && go test ./internal/module/challenge/... ./internal/module/practice/... ./internal/module/runtime/...`
- [ ] **Step 2: Run focused frontend verification**
  Run: `cd code/frontend && npm test -- --run src/api/__tests__/instance.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts src/views/instances/__tests__/InstanceList.test.ts`
- [ ] **Step 3: Review API contract impact and update smallest necessary docs**
- [ ] **Step 4: Re-read design doc and confirm scope did not drift into container-pool implementation**
