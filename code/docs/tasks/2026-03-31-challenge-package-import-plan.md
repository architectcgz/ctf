# Challenge Package Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace manual challenge creation with a `challenge.yml`-driven package import workflow in the admin challenge management flow.

**Architecture:** Keep `challenge` as the owner of challenge package parsing and import orchestration. Reuse one shared parser for both admin online import and CLI pack import, and keep runtime ownership in the existing runtime module by importing only image/runtime references in v1.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, Vue 3, TypeScript, Element Plus, Vitest

---

### Task 1: Add design-aligned failing tests for challenge package import

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/api/http/handler.go`
- Create: `code/backend/internal/module/challenge/api/http/import_http_integration_test.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/frontend/src/views/admin/__tests__/ChallengeManage.test.ts`

- [ ] **Step 1: Add failing backend parser/import tests**
- [ ] **Step 2: Add failing admin route test for upload preview and commit**
- [ ] **Step 3: Add failing frontend test asserting import is the primary action**
- [ ] **Step 4: Run focused tests to verify RED**

### Task 2: Extract shared `challenge.yml` package parser in challenge module

**Files:**
- Create: `code/backend/internal/module/challenge/domain/package_manifest.go`
- Create: `code/backend/internal/module/challenge/domain/package_parser.go`
- Create: `code/backend/internal/module/challenge/domain/package_parser_test.go`
- Modify: `code/backend/cmd/import-challenge-packs/main.go`

- [ ] **Step 1: Define `challenge.yml` manifest structs and validation rules**
- [ ] **Step 2: Implement zip/directory parser with statement and attachment resolution**
- [ ] **Step 3: Rewire existing CLI importer to use the shared parser**
- [ ] **Step 4: Run focused parser tests**

### Task 3: Implement backend admin challenge-import workflow

**Files:**
- Create: `code/backend/internal/dto/challenge_import.go`
- Create: `code/backend/internal/model/challenge_import.go`
- Create: `code/backend/migrations/000043_create_challenge_imports_table.up.sql`
- Create: `code/backend/migrations/000043_create_challenge_imports_table.down.sql`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/repository.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/api/http/handler.go`
- Modify: `code/backend/internal/app/router_routes.go`

- [ ] **Step 1: Add import DTOs and persistence model for preview/commit state**
- [ ] **Step 2: Implement upload-preview command**
- [ ] **Step 3: Implement commit-import command to create/update challenge draft**
- [ ] **Step 4: Register admin routes for import preview and commit**
- [ ] **Step 5: Run focused backend import tests**

### Task 4: Replace manual create UI with package import UI

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Create: `code/frontend/src/composables/useChallengePackageImport.ts`
- Create: `code/frontend/src/components/admin/challenge/ChallengePackageImportEntry.vue`
- Create: `code/frontend/src/components/admin/challenge/ChallengePackageImportReview.vue`
- Modify: `code/frontend/src/views/admin/ChallengeManage.vue`

- [ ] **Step 1: Add frontend contracts for import preview and commit**
- [ ] **Step 2: Implement import composable**
- [ ] **Step 3: Replace “创建挑战” dialog with upload and review flow**
- [ ] **Step 4: Keep challenge list actions intact**
- [ ] **Step 5: Run focused frontend tests**

### Task 5: Validation and documentation sync

**Files:**
- Modify: `code/docs/tasks/frontend-task-breakdown.md`
- Modify: `code/docs/tasks/2026-03-31-challenge-package-import-design.md`

- [ ] **Step 1: Run backend validation**
  Run: `cd code/backend && go test ./internal/module/challenge/... ./internal/app/...`
- [ ] **Step 2: Run frontend validation**
  Run: `cd code/frontend && npm test -- --run src/views/admin/__tests__/ChallengeManage.test.ts`
- [ ] **Step 3: Run frontend typecheck**
  Run: `cd code/frontend && npm run typecheck`
- [ ] **Step 4: Review doc impact and update smallest relevant docs**
- [ ] **Step 5: Run code review and fix findings**
