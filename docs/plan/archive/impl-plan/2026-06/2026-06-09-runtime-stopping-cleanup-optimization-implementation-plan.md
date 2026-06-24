<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# runtime-stopping-cleanup-optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent step-by-step execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce stopping-instance cleanup database pressure and duplicate multi-node cleanup work while preserving reliable eventual runtime cleanup.

**Architecture:** Keep the database `status = stopping` row as the cleanup source of truth. Add a bounded cleanup batch, a Redis-backed cleanup lock, a local event wake-up signal, and a matching database index; the event only accelerates cleanup and the polling loop remains the recovery fallback.

**Tech Stack:** Go, Gin, GORM, PostgreSQL migrations, Redis lock, in-process `platform/events` bus, Go tests.

---

## Task Metadata

- Task Slug: `2026-06-09-runtime-stopping-cleanup-optimization`
- Started At: `2026-06-09T02:01:15Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-runtime-stopping-cleanup-optimization`
- Branch: `task/2026-06-09-runtime-stopping-cleanup-optimization`

## Objective And Non-Goals

- Objective:
  - Bound `ListStoppingInstances` reads by configured cleanup concurrency.
  - Add a PostgreSQL index that matches `status = ? ORDER BY updated_at, id` for stopping cleanup.
  - Prevent duplicate stopping cleanup runs across API replicas with a Redis lock.
  - Publish a best-effort local wake-up signal after a manual destroy transitions an instance to `stopping`.
  - Record the good-taste rule in `feedback/`: events can wake workers, but durable DB state remains the owner and polling stays as fallback.
- Non-Goals:
  - Do not replace polling with events.
  - Do not introduce a cross-process message broker.
  - Do not change the public HTTP API response contract.
  - Do not move instance ownership back into `runtime/application` compatibility wrappers.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
  - `feedback/AGENTS.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Related architecture/contracts:
  - `instance` owns instance command/query/maintenance behavior.
  - `runtime` owns low-level container cleanup capability.
  - `platform/events` is an in-process, non-critical-path bus and does not replace strong state ownership.
- Related prior work:
  - Current `DestroyInstance` / `DestroyTeacherInstance` only mark `stopping`.
  - Current `RunStoppingCleanupLoop` scans every `delete_poll_interval`, default `1s`.
  - Current `runtime_cleaner` already uses Redis lock for multi-node cron work.

## Task Classification

- Classification: `非琐碎任务`
- Why: touches runtime cleanup behavior, repository query shape, migration/indexing, background worker concurrency, event wiring, and harness feedback.

## Files

- Create:
  - `code/backend/migrations/000014_instances_stopping_cleanup_index.up.sql`
  - `code/backend/migrations/000014_instances_stopping_cleanup_index.down.sql`
  - `feedback/2026-06-09-event-wakeup-keeps-durable-state-owner.md`
- Modify:
  - `code/backend/internal/module/instance/contracts/events.go`
  - `code/backend/internal/module/instance/application/commands/instance_service.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/configs/config.yaml` only if a new lock TTL config is required; prefer reusing existing cleanup lock TTL first.
- Review:
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - Migration ordering and rollback.
- Test:
  - `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/runtime/service_repository_test.go`
  - `code/backend/internal/module/runtime/infrastructure/cleaner_test.go` if Redis lock behavior is extracted there.
  - `code/backend/internal/module/runtime/application/instance_service_test.go` or `code/backend/internal/module/instance/application/commands/*_test.go` for wake-up publish.

## 复用与 Owner 决策

- Existing patterns searched:
  - Redis lock: `runtime/infrastructure/cleaner.go`, `infrastructure/redislock`, `shared/lockkeepalive`.
  - In-process events: `platform/events`, practice/contest event publishing.
  - Instance owner: `instance/application/commands`.
- Reuse / extend / split / create-new decision:
  - Reuse `platform/events.Bus` for local wake-up.
  - Reuse `redislock.Acquire` and existing cleanup lock TTL for multi-node stopping cleanup.
  - Extend repository method signature to accept a `limit`.
  - Add a dedicated cache key for stopping cleanup lock, separate from cron cleanup lock.
- Owner boundary:
  - `InstanceService` owns the manual destroy transition and publishes the wake-up after successful `MarkStopping`.
  - `InstanceMaintenanceService` owns selecting stopping rows, locking, cleanup, and finalization.
  - `RuntimeCleanupService` remains the owner of low-level runtime resource deletion.
  - `runtime/infrastructure.Repository` owns persistence query shape and finalization.
- Why this is the narrowest safe surface:
  - It changes the cleanup trigger and query shape without changing HTTP API, status model, runtime resource extraction, or provisioning.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: the work adds a new asynchronous trigger and changes runtime maintenance behavior, so the owner, reliability, and fallback boundaries need explicit choice before implementation.
- grill-with-docs findings:
  - Existing architecture docs say `platform/events` is in-process and not a strong consistency or cross-process broker.
  - `instance` is the production owner for instance command/query/maintenance; runtime compatibility wrappers are not the owner.
  - Therefore the event must be a wake-up hint only; durable state remains `instances.status`.
- Plan adjustments after challenge:
  - Keep polling fallback.
  - Add Redis lock around each cleanup dispatch to avoid multi-node duplicate cleanup.
  - Add batch limit and matching index before reducing/altering polling cadence.

## Validation

- Commands:
  - `go test ./internal/module/runtime/... ./internal/module/instance/...`
  - `go test ./internal/app/composition/...` if composition tests compile/run as a package.
  - `bash scripts/check-startup-gate.sh`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash scripts/check-workflow-governance.sh`
- Manual checks:
  - Confirm generated migration names are ordered after existing migrations.
  - Confirm event wake-up does not make cleanup correctness depend on event delivery.
- Review focus:
  - No request-context goroutine leaks.
  - Redis lock release and nil-Redis behavior.
  - Batch limit does not starve old stopping rows.
  - Event publish is best-effort and cannot break `DestroyInstance`.
  - Migration rollback is credible.

## Task Steps

### Task 1: Repository Batch Limit And Index

**Files:**
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Modify tests: `code/backend/internal/module/runtime/service_repository_test.go`
- Create: `code/backend/migrations/000014_instances_stopping_cleanup_index.up.sql`
- Create: `code/backend/migrations/000014_instances_stopping_cleanup_index.down.sql`

- [x] Write a failing repository test proving `ListStoppingInstances(ctx, cutoff, limit)` returns at most `limit` rows in `updated_at ASC, id ASC` order.
- [x] Run the targeted repository test and verify it fails before implementation.
- [x] Implement the repository limit, treating `limit <= 0` as the existing unbounded behavior for compatibility.
- [x] Add the composite index migration and rollback.
- [x] Run the targeted repository test and verify it passes.

### Task 2: Multi-Node Lock Around Stopping Cleanup Dispatch

**Files:**
- Modify: `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify tests: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`

- [x] Write failing tests proving stopping cleanup only queries/dispatches when the lock is acquired and skips when another node holds it.
- [x] Run the targeted maintenance test and verify it fails before implementation.
- [x] Add a small lock port to `InstanceMaintenanceService` and wire it to Redis in composition.
- [x] Keep nil lock store behavior permissive for tests/local no-Redis paths.
- [x] Run the targeted maintenance test and verify it passes.

### Task 3: Event Wake-Up

**Files:**
- Create/modify: `code/backend/internal/module/instance/contracts/events.go`
- Modify: `code/backend/internal/module/instance/application/commands/instance_service.go`
- Modify: `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify tests: `code/backend/internal/module/runtime/application/instance_service_test.go`
- Modify tests: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`

- [x] Write failing tests proving `DestroyInstance` publishes a best-effort wake-up only after `MarkStopping` succeeds.
- [x] Write a failing test proving the maintenance service wake-up handler triggers one cleanup dispatch without waiting for the ticker.
- [x] Run targeted tests and verify they fail before implementation.
- [x] Implement event constants, best-effort publishing, and wake-up subscription.
- [x] Ensure event handler uses lifecycle context, not request context, for actual cleanup work.
- [x] Run targeted tests and verify they pass.

### Task 4: Harness Feedback And Validation

**Files:**
- Create: `feedback/2026-06-09-event-wakeup-keeps-durable-state-owner.md`

- [x] Record the good-taste rule in `feedback/` with required `## 沉淀状态`.
- [x] Run narrow Go package tests.
- [x] Run startup gate and workflow completion checks.
- [x] Run independent backend review gate and archive review evidence.
- [x] Fix material findings and re-run impacted validation.
