<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 拆分 practice submission 大测试文件 Implementation Plan

**Goal:** Split the oversized practice submission test file into smaller behavior-owned files without changing production behavior.

**Architecture:** Keep all tests in the existing `commands` package and reuse existing stubs/helpers. Move tests by behavior owner: submission flow, manual review, submission history, and dynamic flag / solve-grace behavior.

**Tech Stack:** Go test, existing practice command test helpers, code-workflow gates.

---

## Task Metadata

- Task Slug: `2026-06-08-split-practice-submission-tests`
- Started At: `2026-06-08T08:28:39Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-08-multi-instance-flag-secret-contract/2026-06-08-split-practice-submission-tests`
- Branch: `task/2026-06-08-split-practice-submission-tests`

## Objective And Non-Goals

- Objective:
  - Reduce `submission_manual_review_test.go` from a mixed 1900+ line file to a focused manual-review test file.
  - Preserve every existing test function and assertion.
  - Move only test code; no production code behavior changes.
- Non-Goals:
  - Do not refactor `Service` production code.
  - Do not redesign repository stubs or broaden package-wide test execution.
  - Do not delete or rename the original test file.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
  - `code/backend/internal/module/practice/application/commands/service_test.go`
  - `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- Related prior work:
  - Previous dynamic flag and solve-grace contract work pushed more submit-path tests into `submission_manual_review_test.go`, making the mixed file harder to review and maintain.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Test-only, but it touches many behavior tests in a protected backend module and needs workflow validation/review to avoid silently dropping coverage.

## Files

- Create:
  - `code/backend/internal/module/practice/application/commands/submission_test_helpers_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_submit_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_history_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_dynamic_flag_test.go`
- Modify:
  - `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Review:
  - No test function lost during move.
  - Imports stay minimal and `gofmt` clean.
  - Existing helper visibility remains package-local.
- Test:
  - Focused `go test` over moved test names.
  - Existing workflow `pre-commit-quick`.

## 复用与 Owner 决策

- Existing patterns searched:
  - Existing package already keeps broad shared stubs in `service_test.go` and `repository_stub_test.go`.
  - Other practice command tests are split by behavior, for example instance provisioning and contest instance service tests.
- Reuse / extend / split / create-new decision:
  - Reuse existing stubs; split test functions into new files by behavior owner.
- Owner boundary:
  - `submission_manual_review_test.go` remains owner for teacher/manual review flows.
  - `submission_submit_test.go` owns normal submit/score/update/event tests.
  - `submission_history_test.go` owns student submission history query tests.
  - `submission_dynamic_flag_test.go` owns dynamic flag validation and solve-grace instance interactions.
- Why this is the narrowest safe surface:
  - It reduces the oversized file without changing helper semantics or production code.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task is a structural test organization change; the main risk is choosing the wrong split boundary and dropping coverage.
- grill-with-docs findings:
  - No domain glossary change is needed; the split follows existing test behavior names.
  - The work is current implementation organization, not architecture fact-source content.
- Plan adjustments after challenge:
  - Keep the original file instead of deleting/renaming it.
  - Avoid extracting new abstraction unless duplicated setup clearly blocks compilation.

## Validation

- Commands:
  - `go test ./internal/module/practice/application/commands -run 'TestSubmitFlag|TestPracticePublishesFlagAcceptedEvent|TestListMyChallengeSubmissions|TestListTeacherManualReviewSubmissions|TestGetTeacherManualReviewSubmission|TestReviewManualReviewSubmission|TestBuildInstanceFlagUsesGlobalSecret' -count=1`
  - `git diff --check`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - `bash scripts/run-workflow-stage.sh completion-full`
- Manual checks:
  - `rg '^func Test'` before/after to ensure all test names remain present exactly once.
  - duplicate `Test*` name check stays empty.
  - `wc -l` confirms the former oversized file is materially smaller.
- Environment note:
  - `go test ./internal/module/practice/application/commands -count=1` failed because existing Docker container `ctf-instance-iot-c8-t16-s23` binds host port `30000`, causing unrelated AWD readiness tests to fail with `address already in use`.
- Review focus:
  - Test coverage preservation and clean ownership boundaries.
  - Round 2 independent review archived at `docs/reviews/backend/2026-06-08-worktree-review-split-practice-submission-tests-round2.md`.
