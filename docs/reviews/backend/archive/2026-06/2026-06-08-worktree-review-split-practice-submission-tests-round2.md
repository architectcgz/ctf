# Split Practice Submission Tests Gate Review Round 2

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-08-multi-instance-flag-secret-contract/2026-06-08-split-practice-submission-tests`
- Branch: `task/2026-06-08-split-practice-submission-tests`
- Task: `2026-06-08-split-practice-submission-tests`
- Plan: `docs/plan/impl-plan/2026-06-08-split-practice-submission-tests-implementation-plan.md`
- Files reviewed:
  - `code/backend/internal/module/practice/application/commands/service_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_submit_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_history_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_dynamic_flag_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_manual_review_teacher_test.go`
  - `code/backend/internal/module/practice/application/commands/submission_test_helpers_test.go`

## Classification Check

Agree with the non-trivial `code-workflow` classification. The change is test-only, but it restructures an oversized backend test surface and needs evidence that no behavior coverage was lost.

## Gate Verdict

Pass.

## Findings

- No blocker or major findings.
- The previous owner-mixing blocker is closed. `submission_manual_review_test.go` now only keeps manual-review-owned tests.
- Student submission history tests are owned by `submission_history_test.go` and use `wirePracticeSubmissionHistoryAdapters`.
- Shared submit rate-limit helper ownership is neutral in `submission_test_helpers_test.go`.
- Normal submit, event, repeat-submit, solve-grace, and submit error-mapping tests are owned by `submission_submit_test.go`.

## Material Findings

None.

## Senior Implementation Assessment

The current split is the simpler maintainable shape for this task. It keeps existing package-level stubs and assertions, avoids production-code changes, and changes only test ownership and helper placement.

## Required Re-Validation

Already executed by implementation context:

- `gofmt` on touched Go test files.
- Old/new `Test*` function set diff against `HEAD`: empty.
- Duplicate `Test*` name check: empty.
- `go test ./internal/module/practice/application/commands -run 'TestSubmitFlag|TestPracticePublishesFlagAcceptedEvent|TestListMyChallengeSubmissions|TestListTeacherManualReviewSubmissions|TestGetTeacherManualReviewSubmission|TestReviewManualReviewSubmission|TestBuildInstanceFlagUsesGlobalSecret' -count=1`: pass.
- `git diff --check`: pass.
- `bash scripts/run-workflow-stage.sh pre-commit-quick`: pass.
- `bash scripts/run-workflow-stage.sh completion-full`: pass.

Independently rerun by reviewer:

- `go test ./internal/module/practice/application/commands -run 'TestSubmitFlag|TestPracticePublishesFlagAcceptedEvent|TestListMyChallengeSubmissions|TestListTeacherManualReviewSubmissions|TestGetTeacherManualReviewSubmission|TestReviewManualReviewSubmission|TestBuildInstanceFlagUsesGlobalSecret' -count=1`: pass.

## Residual Risk

- Full `go test ./internal/module/practice/application/commands -count=1` did not produce a clean pass in this environment because an existing Docker container `ctf-instance-iot-c8-t16-s23` binds host port `30000`, causing unrelated AWD readiness tests to fail with `address already in use`.
- The diff is test-only reorganization. The focused submission package tests compile the full package and run the moved behavior set, so the port conflict is treated as environment residue rather than evidence of a regression from this change.

## Touched Known-Debt Status

The touched oversized test surface was the known debt. This change closes the mixed-owner portion of `submission_manual_review_test.go` by splitting behavior-owned tests into separate files.
