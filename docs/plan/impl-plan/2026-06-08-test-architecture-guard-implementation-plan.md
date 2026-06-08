<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# Test Architecture Guard Implementation Plan

**Goal:** Add a project harness guard that runs test-architecture checks whenever test code changes.

**Architecture:** Stable script wrapper in `scripts/`, deterministic guard body in `harness/checks/`, workflow-stage wiring under `harness/workflow-plugins/code-workflow/`.

**Tech Stack:** Bash, Python unittest, existing Go/Vitest architecture guard entrypoints.

---

## Task Metadata

- Task Slug: `2026-06-08-test-architecture-guard`
- Started At: `2026-06-08T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-test-architecture-guard`
- Branch: `task/2026-06-08-test-architecture-guard`

## Objective And Non-Goals

- Objective:
  - Add a harness-owned test architecture guard.
  - Ensure workflow stages run it after test files are changed.
  - Keep the guard scoped to test architecture, not full test execution.
- Non-Goals:
  - Do not migrate or rewrite existing frontend/backend tests.
  - Do not change production code.
  - Do not make every ordinary code change run heavy test architecture checks.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `works/backend-test-architecture-rewrite-blueprint.md`
  - `feedback/2026-06-03-backend-test-layering-and-pdf-assertion-ownership.md`
  - `feedback/2026-06-04-frontend-state-surface-owner-and-ui-test-layering.md`
- Related architecture/contracts:
  - `scripts/check-architecture.sh`
  - `scripts/check-backend-architecture.sh`
  - `scripts/check-frontend-test-guard.sh`
  - `harness/workflow-plugins/code-workflow/`
- Related prior work:
  - `harness/checks/check_frontend_test_guard.sh`
  - `code/backend/tests/architecture/test_architecture_test.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Touches harness checks and workflow-stage enforcement.
  - Changes completion/pre-commit behavior for future test work.
  - Needs mechanical validation and review of trigger scope.

## Files

- Create:
  - `scripts/check-test-architecture.sh`
  - `harness/checks/check_test_architecture_guard.py`
  - `harness/checks/test_test_architecture_guard.py`
  - workflow-stage plugin files if needed
- Modify:
  - `harness/workflow-plugins/code-workflow/pre-commit-quick.d/`
  - `harness/workflow-plugins/code-workflow/completion-full.d/`
  - `scripts/lib/check-consistency/architecture.sh`
  - local executable manifest if needed
- Review:
  - Existing architecture and frontend test guard scripts.
- Test:
  - Python unittest for trigger classification.
  - Red/green script behavior checks.
  - Relevant workflow stages.

## 复用与 Owner 决策

- Existing patterns searched:
  - Stable wrapper scripts delegate into `harness/checks/`.
  - Workflow plugins consume `WORKFLOW_CHANGED_FILES`.
  - Frontend test guard already supports working-tree/staged/files modes.
- Reuse / extend / split / create-new decision:
  - Create a new test architecture guard because this is neither generic architecture nor frontend-only fine-grained assertion checking.
  - Reuse existing backend architecture package and frontend test guard instead of duplicating those checks.
- Owner boundary:
  - `harness/checks/check_test_architecture_guard.py` owns test-change detection and dispatch.
  - Backend architecture tests remain in `code/backend/tests/architecture`.
  - Frontend assertion guard remains in `harness/checks/check_frontend_test_guard.sh`.
- Why this is the narrowest safe surface:
  - It adds a routing guard and workflow wiring without changing existing test semantics.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task is workflow behavior design with trigger-scope tradeoffs.
- grill-with-docs findings:
  - Existing docs already define backend test owner layering and frontend raw-test limits.
  - No new glossary or ADR is needed; this is mechanical enforcement of existing direction.
- Plan adjustments after challenge:
  - Keep guard scoped to test-related paths and reuse existing backend/frontend checks.

## Validation

- Commands:
  - `python3 -m unittest harness.checks.test_test_architecture_guard`
  - Red/green command proving missing guard wiring is caught.
  - `bash scripts/check-test-architecture.sh --working-tree`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/check-workflow-complete.sh`
- Manual checks:
  - Confirm unrelated dirty work in main worktree is untouched.
- Review focus:
  - Trigger scope: avoids false negatives for test code, avoids excessive runtime for ordinary implementation.
