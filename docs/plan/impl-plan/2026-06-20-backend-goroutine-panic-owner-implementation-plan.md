<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# 2026-06-20-backend-goroutine-panic-owner Implementation Plan

**Goal:** 撤回共享 `safego` 试点，把 goroutine panic 语义收回到各自 owner，并把可复用代码范例沉淀到 CTF 后端 harness 规则。

**Architecture:** 本任务不新增替代性共享 goroutine helper。root background job、practice async task、runtime cleaner 继续显式持有 goroutine、cancel、wait 和 panic 语义；项目规则通过 `ctf-backend-patterns` 路由，完整背景保留在 `feedback/`。

**Tech Stack:** Go, context.Context, sync.WaitGroup, zap, Go architecture tests, project harness feedback/skills

---

## Task Metadata

- Task Slug: `2026-06-20-backend-goroutine-panic-owner`
- Started At: `2026-06-20T13:09:13Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-20-backend-goroutine-panic-owner`
- Branch: `task/2026-06-20-backend-goroutine-panic-owner`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `implemented` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 删除 `internal/shared/safego` 共享 helper 及其 helper 导入 guard。
  - 将三个试点文件恢复为显式 goroutine 写法，避免共享 helper 默认吞 panic。
  - 用架构测试约束“不得重新引入共享 SafeGo 默认包装”。
  - 在 CTF 后端 harness 规则中加入代码范例，说明推荐与禁止写法。
- Non-Goals:
  - 不治理仓库内全部裸 goroutine。
  - 不新增 worker supervisor、panic reporter、errgroup 封装或重启机制。
  - 不改变 outbox/stream/practice/cleaner 的业务调度语义。

## Problem Statement

- Current behavior / structure:
  - `internal/shared/safego.Go` 同时负责 goroutine 启动、panic recover、日志和 WaitGroup 同步。
  - `root.go`、`service_lifecycle.go`、`cleaner.go` 三处试点依赖 `safego`。
  - 架构测试锁定三处试点必须导入 `safego`，这与新的 owner 边界规则相反。
- Target behavior / structure:
  - goroutine 启动保持显式；panic 后果由 owner 决定，不由共享 helper 默认吞掉。
  - wait-only goroutine 使用最小裸 `go func(){ wg.Wait(); close(done) }()`。
  - harness 规则包含具体代码范例，后续实现能直接对照。
- Why this task is needed now:
  - 用户已经明确要求试点也不使用 `safego`，并要求先写规则、再改代码、最后补代码范例。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `feedback/AGENTS.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `docs/plan/impl-plan/2026-06-13-backend-async-safego-implementation-plan.md`
  - `docs/reviews/backend/2026-06-13-backend-review-async-safego.md`
- Related prior work:
  - SafeGo 试点 slice：`2026-06-13-backend-async-safego`
  - 本任务前已在主工作区形成的规则结论：goroutine panic owner boundary

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达后端并发语义、共享包删除、架构测试、项目 harness skill 和 feedback。

## Files

- Create:
  - `feedback/2026-06-20-goroutine-panic-owner-boundary.md`
- Modify:
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/app/composition/root_safego_test.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle_safego_test.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner_safego_test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
- Review:
  - `code/backend/internal/app/composition/background_job_loop.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
- Test:
  - `code/backend/tests/architecture/test_architecture_test.go`
  - `code/backend/internal/app/composition/root_safego_test.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner_safego_test.go`
  - `code/backend/internal/shared/safego/safego_test.go`
  - Delete: `code/backend/internal/shared/safego/safego.go`
  - Delete: `code/backend/internal/shared/safego/safego_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `NewLoopBackgroundJob` already owns explicit goroutine lifecycle with `sync.WaitGroup`.
  - `ReportService.runAsyncReport` shows owner-local panic-to-business-failure behavior.
  - Middleware recovery remains the HTTP request panic boundary.
- Reuse / extend / split / create-new decision:
  - Reuse explicit owner-local goroutine patterns; delete shared helper instead of creating a replacement abstraction.
- Owner boundary:
  - root composition owns root background job lifecycle.
  - practice service owns practice async task lifecycle.
  - runtime cleaner owns cleaner run/stop lifecycle.
  - harness skill owns reusable guidance; feedback keeps archived context.
- Why this is the narrowest safe surface:
  - Only the existing SafeGo package, three试点 callers, and their guardrails/tests are changed.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task changes concurrency strategy and failure semantics rather than fixing a single failing line.
- grill-with-docs findings:
  - No user question blocks implementation: user explicitly chose to stop using `safego` even for试点.
  - Existing docs/review identify the SafeGo slice as a试点 and not full-repo policy.
  - The new harness rule should be active in `.agents/skills/ctf-backend-patterns`, with feedback archived after absorption.
- Plan adjustments after challenge:
  - Do not introduce `safe go v2`.
  - Keep code examples in harness rules after implementation so they match actual code shape.
  - Use architecture guard to prevent reintroducing `internal/shared/safego`, not to require a helper.

## Execution Slices

### Slice 1: Red tests for no shared SafeGo default

- Goal:
  - Replace old SafeGo helper tests/guard with tests that fail while `internal/shared/safego` still exists or is imported.
- Dependencies:
  - Plan ready.
- Files:
  - Create:
    - 无
  - Modify:
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `code/backend/internal/app/composition/root_safego_test.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle_safego_test.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner_safego_test.go`
  - Review:
    - `code/backend/internal/shared/safego/safego_test.go`
  - Test:
    - `go test ./tests/architecture -run 'TestNoSharedSafeGoDefault' -count=1`
    - impacted package tests as needed
- Steps:
  - [x] Step 1: Write failing architecture test rejecting `internal/shared/safego` package/imports.
  - [x] Step 2: Rename/remove old safego behavior tests so they describe owner behavior instead of helper usage.
  - [x] Step 3: Run the architecture test and confirm it fails because safego still exists/imports remain.
- Validation:
  - Red test fails for the intended reason.
- Review focus:
  - Tests prove owner boundary, not helper import preference.
- Done criteria:
  - Red state captured before production code changes.

### Slice 2: Remove safego helper and restore owner-local goroutine code

- Goal:
  - Delete `internal/shared/safego`; update three callers to explicit goroutine lifecycle.
- Dependencies:
  - Slice 1 red test.
- Files:
  - Create:
    - 无
  - Modify:
    - `code/backend/internal/app/composition/root.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - Delete:
    - `code/backend/internal/shared/safego/safego.go`
    - `code/backend/internal/shared/safego/safego_test.go`
  - Test:
    - `go test ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
    - `go test ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
    - `go test ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
- Steps:
  - [x] Step 1: Remove safego imports and replace `safego.Go` calls with explicit `wg.Add` / `go func` / `defer wg.Done`.
  - [x] Step 2: Keep wait-only goroutines naked and minimal.
  - [x] Step 3: Delete shared safego package and tests.
  - [x] Step 4: Run impacted package tests and architecture test.
- Validation:
  - Package tests and architecture test pass.
- Review focus:
  - No helper fallback hides nil ctx/logger.
  - No business scheduling semantics changed.
- Done criteria:
  - No production code imports `internal/shared/safego`; package removed.

### Slice 3: Add harness code examples

- Goal:
  - Add concrete good/bad code examples to the CTF backend pattern rule after code shape is settled.
- Dependencies:
  - Slice 2 green.
- Files:
  - Create:
    - `feedback/2026-06-20-goroutine-panic-owner-boundary.md`
  - Modify:
    - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - Review:
    - `feedback/AGENTS.md`
  - Test:
    - `bash scripts/check-shared-skills.sh`
    - `bash scripts/check-workflow-governance.sh`
- Steps:
  - [x] Step 1: Add rule summary and code examples to `ctf-backend-patterns`.
  - [x] Step 2: Add archived feedback context with owner link.
  - [x] Step 3: Run harness validation.
- Validation:
  - Shared skill and workflow governance checks pass.
- Review focus:
  - Examples match actual code and do not promote a new generic helper.
- Done criteria:
  - Code examples are present in the active harness rule.

## Impact And Compatibility

- API / DTO:
  - None.
- Data / migration:
  - None.
- State / cache / queue / event:
  - No state schema changes; background task panic semantics become owner-local rather than shared helper recover.
- Runtime / config:
  - Root background job panic is no longer swallowed by `safego`.
- Frontend route / state / UX:
  - None.
- Docs / contracts:
  - Updates project harness skill and archived feedback.

## Plan Review / Architecture Fit

- Target owner boundary:
  - Clear: root composition, practice service, cleaner, and harness skill each own their part.
- Reuse points / landing zones:
  - Reuse explicit goroutine lifecycle from `NewLoopBackgroundJob`; reuse report-service pattern as documented example for business-state async tasks.
- Known structural debt touched:
  - The existing `internal/shared/safego`试点 itself is the debt being removed.
- How this plan avoids behavior-only convergence:
  - It removes the shared abstraction and its guardrail, not only changing log text or tests.
- Hidden second-redesign risk:
  - Low for this slice because no replacement abstraction is introduced.
- Decision after review:
  - Approved for implementation in this task worktree.

## Documentation Owner

- Current fact sources to read:
  - `AGENTS.md`
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `feedback/AGENTS.md`
  - `docs/文档规范.md`
- Fact sources to update after implementation:
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `feedback/2026-06-20-goroutine-panic-owner-boundary.md`
- Plan-only notes that must not become architecture source:
  - Historical SafeGo plan remains historical; do not edit it into current fact source.
- Archive condition:
  - Archive after implementation, validation, and review are complete.

## Validation

- Required gate summary:
  - Red architecture guard failed before implementation for the existing shared `safego` package/imports.
  - Affected package tests, focused race checks, production `safego` residual scan, `completion-full`, and workflow governance passed after implementation.
  - Draft review on the current diff found no blocker or material finding; formal review still needs a commit/range after commit.
- Detailed commands and evidence:
  - See `## Validation Plan` and `## Validation Evidence` below.

## Validation Plan

- Per-slice commands:
  - `go test ./tests/architecture -run 'TestNoSharedSafeGoDefault' -count=1`
  - `go test ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
  - `go test ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
  - `go test ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
  - `bash scripts/check-shared-skills.sh`
  - `bash scripts/check-workflow-governance.sh`
- Integration commands:
  - `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
- Manual checks:
  - `rg -n '"ctf-platform/internal/shared/safego"|safego\\.Go' code/backend/internal code/backend/tests/architecture`
  - Confirm `.agents/skills/ctf-backend-patterns/SKILL.md` and `feedback/2026-06-20-goroutine-panic-owner-boundary.md` contain code examples for allowed owner-local patterns and the forbidden shared-helper anti-pattern.
- Commands intentionally skipped and why:
  - Full backend test suite unless impacted package tests reveal broader risk.

## Validation Evidence

- Command: `go test ./tests/architecture -run TestNoSharedSafeGoDefault -count=1`
  - Result: FAIL before implementation
  - Notes: Red test failed because `internal/shared/safego` still existed and production callers still imported / called it.
- Command: `go test ./tests/architecture -run TestNoSharedSafeGoDefault -count=1`
  - Result: PASS after implementation
  - Notes: Guard now permits only the architecture test's own forbidden snippets.
- Command: `go test ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
  - Result: PASS
  - Notes: Root background job context behavior and nil context rejection preserved.
- Command: `go test ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
  - Result: PASS
  - Notes: Practice async task shutdown behavior preserved after removing safego.
- Command: `go test ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
  - Result: PASS
  - Notes: Cleaner stop / cancellation behavior preserved after removing safego.
- Command: `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
  - Result: PASS
  - Notes: Impacted packages and architecture guard pass together.
- Command: `go test -race ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
  - Result: PASS
  - Notes: Focused race check for root background job tests.
- Command: `go test -race ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
  - Result: PASS
  - Notes: Focused race check for practice async task lifecycle tests.
- Command: `go test -race ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
  - Result: PASS
  - Notes: Focused race check for cleaner lifecycle tests.
- Command: `bash scripts/check-shared-skills.sh`
  - Result: PASS
  - Notes: Project skill files remain discoverable and valid after adding code examples.
- Command: `rg -n '"ctf-platform/internal/shared/safego"|safego\.Go' code/backend/internal`
  - Result: PASS
  - Notes: Command returned no matches, so production/internal backend code no longer imports or calls `safego`.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS
  - Notes: Startup gate check did not report blocking gated changes.
- Command: `bash scripts/run-workflow-stage.sh completion-full`
  - Result: PASS
  - Notes: Code change contracts, backend architecture, frontend architecture no-op check, and backend test architecture guard passed.
- Command: `bash scripts/check-workflow-governance.sh`
  - Result: PASS
  - Notes: Workflow governance accepted the updated skill, archived feedback status, and review/harness structure.
- Review: `docs/reviews/backend/2026-06-20-backend-review-goroutine-panic-owner-draft.md`
  - Result: PASS for current diff
  - Notes: Draft review found no blocker or material finding; formal gate review still needs a commit/range if this task is committed.

## Independent Review Handoff

- Review target:
  - Current worktree diff on branch `task/2026-06-20-backend-goroutine-panic-owner`.
  - Draft review file: `docs/reviews/backend/2026-06-20-backend-review-goroutine-panic-owner-draft.md`.
- Validation evidence summary:
  - Red architecture test failed before implementation for existing `internal/shared/safego`.
  - Affected package tests, aggregate package test, focused race checks, `check-shared-skills`, `completion-full`, and `workflow-governance` passed after implementation.
  - Production/internal backend scan found no `safego` import or `safego.Go` call.
- Architecture / contract inputs:
  - `AGENTS.md`
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `feedback/2026-06-20-goroutine-panic-owner-boundary.md`
  - `code/backend/tests/README.md`
- Known risks / review focus:
  - Root background job panic now re-panics after logging by owner.
  - Practice async task and runtime cleaner no longer share a default recover helper.
  - Architecture guard must forbid shared SafeGo default without banning owner-local goroutines.
- Project-local checks to consider:
  - `go test ./tests/architecture -run TestNoSharedSafeGoDefault -count=1`
  - `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
  - focused `go test -race` commands for the three touched package areas
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/check-workflow-governance.sh`

## Rollback / Recovery

- Safe revert boundary:
  - Revert this task branch to restore prior safego package, callers, tests, and harness text.
- Data / config / runtime recovery notes:
  - No data/config changes.
- Irreversible operations:
  - None.

## Residual Risks

- Risk:
  - Root background job panic may now crash the process instead of being logged and swallowed.
- Why acceptable:
  - This matches the new owner rule: critical root jobs must not silently die while API keeps serving.
- Follow-up owner, if any:
  - If restart/supervision policy is desired, create a separate root lifecycle/supervisor task.
