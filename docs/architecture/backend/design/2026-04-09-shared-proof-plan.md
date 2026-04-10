# Shared Proof Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `shared_proof` submission mode so shared challenge instances can keep one runtime environment while issuing one-time, user-bound submission proofs for both practice and contest flows.

**Architecture:** Reuse the existing runtime proxy ticket chain as the authenticated context carrier, extend it with challenge and contest scope, persist hashed proofs in a dedicated table, and validate/consume proofs inside the existing practice and contest submission services instead of bypassing the normal submit pipeline.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, Redis, Vue 3, TypeScript, Vitest

---

### Task 1: Lock the new behavior with failing backend tests

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/flag_service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- Modify: `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`

- [ ] **Step 1: Add failing flag configuration tests for `shared_proof`**
- [ ] **Step 2: Add failing practice submission tests for proof success, replay rejection, and cross-user rejection**
- [ ] **Step 3: Add failing contest submission tests for contest-bound proof validation**
- [ ] **Step 4: Add failing proxy ticket tests for extended claims**
- [ ] **Step 5: Run focused tests and confirm RED**

### Task 2: Add shared proof persistence and model constants

**Files:**
- Modify: `code/backend/internal/model/challenge.go`
- Create: `code/backend/internal/model/shared_proof.go`
- Create: `code/backend/migrations/000052_create_shared_proofs_table.up.sql`
- Create: `code/backend/migrations/000052_create_shared_proofs_table.down.sql`

- [ ] **Step 1: Define `FlagTypeSharedProof` model constant**
- [ ] **Step 2: Add shared proof model with status constants**
- [ ] **Step 3: Write migration for `shared_proofs` table and indexes**
- [ ] **Step 4: Run migration-sensitive tests**

### Task 3: Enforce challenge configuration rules

**Files:**
- Modify: `code/backend/internal/dto/challenge.go`
- Modify: `code/backend/internal/module/challenge/application/commands/flag_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/flag_service.go`

- [ ] **Step 1: Extend DTO enum bindings with `shared_proof`**
- [ ] **Step 2: Add flag service command for configuring shared proof mode**
- [ ] **Step 3: Reject `shared_proof` on non-shared challenges and reject `dynamic` on shared challenges**
- [ ] **Step 4: Expose `shared_proof` as configured in flag query responses**
- [ ] **Step 5: Run focused challenge module tests**

### Task 4: Extend runtime proxy ticket context and proof issuing service

**Files:**
- Modify: `code/backend/internal/module/runtime/ports/http.go`
- Modify: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Create: `code/backend/internal/module/runtime/application/queries/shared_proof_service.go`
- Create: `code/backend/internal/module/runtime/application/queries/shared_proof_service_test.go`

- [ ] **Step 1: Extend proxy ticket claims with `challenge_id`, `contest_id`, and `share_scope`**
- [ ] **Step 2: Resolve instance metadata before issuing proxy tickets**
- [ ] **Step 3: Add shared proof issue service that validates shared instance context and stores proof hashes**
- [ ] **Step 4: Add internal runtime endpoint for shared proof issuance**
- [ ] **Step 5: Run focused runtime query tests**

### Task 5: Wire shared proof validation into practice submissions

**Files:**
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- Modify: `code/backend/internal/module/practice/infrastructure/repository.go`

- [ ] **Step 1: Add repository port methods for finding and consuming shared proofs**
- [ ] **Step 2: Branch practice submit validation on `flag_type = shared_proof`**
- [ ] **Step 3: Consume proof inside the same logical success path as normal correct submissions**
- [ ] **Step 4: Run focused practice submission tests**

### Task 6: Wire shared proof validation into contest submissions

**Files:**
- Modify: `code/backend/internal/module/challenge/contracts/contracts.go`
- Modify: `code/backend/internal/module/contest/ports/submission.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- Modify: `code/backend/internal/module/contest/infrastructure/submission_lookup_repository.go`

- [ ] **Step 1: Stop relying on challenge flag validator for `shared_proof` contests**
- [ ] **Step 2: Add contest repository methods needed for proof lookup and consumption**
- [ ] **Step 3: Validate proof against contest, user, challenge, and instance scope**
- [ ] **Step 4: Run focused contest submission tests**

### Task 7: Minimal contract and frontend sync

**Files:**
- Modify: `docs/contracts/openapi-v1.yaml`
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/views/challenges/ChallengeDetail.vue`
- Modify: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

- [ ] **Step 1: Add `shared_proof` to API and frontend flag type enums**
- [ ] **Step 2: Adjust challenge detail copy so shared proof challenges do not imply static flag semantics**
- [ ] **Step 3: Run focused frontend tests and typecheck**

### Task 8: Validation and documentation sync

**Files:**
- Modify: `docs/architecture/backend/design/2026-04-09-shared-proof-design.md`

- [ ] **Step 1: Run backend verification**
  Run: `cd code/backend && go test ./internal/module/challenge/... ./internal/module/practice/... ./internal/module/runtime/... ./internal/module/contest/...`
- [ ] **Step 2: Run focused frontend verification**
  Run: `cd code/frontend && npm test -- --run src/views/challenges/__tests__/ChallengeDetail.test.ts && npm run typecheck`
- [ ] **Step 3: Re-read the design doc and confirm the implementation still uses proxy ticket reuse instead of a parallel session system**
- [ ] **Step 4: Update any drift between final code and contract/doc wording**
