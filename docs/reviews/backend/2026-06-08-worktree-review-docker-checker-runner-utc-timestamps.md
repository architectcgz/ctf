# Docker Checker Runner UTC Timestamp Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Task slug: `2026-06-08-backend-architecture-guard-quality`
- Plan input read before review: `docs/plan/archive/impl-plan/2026-06/2026-06-08-backend-architecture-guard-quality-implementation-plan.md`
- Diff source: current uncommitted diff against `HEAD`
- Files reviewed:
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`

## Classification Check

- Agree with `非琐碎任务`.

## Gate Verdict

- `pass`

## Findings

- No findings.

## Material Findings

- None.

## Non-blocking Suggestions

- None.

## Missing Validation

- No blocker-level validation gap was found.
- Independent reviewer reran:
  - `timeout 180s bash -lc 'cd code/backend && go test ./internal/module/contest/infrastructure ./internal/module -run "TestTimeNowUsageExceptionsAreCurrent|TestModuleArchitectureBoundaries" -count=1'`
- There is still no behavior-specific unit test that asserts `CheckerRunResult.StartedAt` / `FinishedAt` are UTC while `Duration` continues to come from the original monotonic `time.Now()` pair across early-return branches. Current confidence comes from direct code inspection plus the architecture guard, not from a branch-by-branch runtime assertion.

## Open Questions Or Assumptions

- This review assumes the project-wide backend time contract in `AGENTS.md` remains authoritative: business timestamps that can enter JSON / audit / cross-request state must be normalized to UTC, while pure runtime duration measurement may keep plain `time.Now()`.
- I did not find a current serialization path that emits `CheckerRunResult.StartedAt` / `FinishedAt` directly; the visible downstream consumer in this repo is the checker audit builder's duration fallback. The UTC normalization is still consistent with the stated contract and does not regress that fallback.

## Senior Implementation Assessment

- The introduced `finish()` closure is the simplest safe shape for this change. It removes repeated completion bookkeeping without changing the existing error/control-flow structure.
- The monotonic duration requirement is preserved:
  - `startedAt` is still captured from plain `time.Now()`.
  - `result.Duration` is computed from `finishedAt.Sub(startedAt)`, so the monotonic component remains available.
  - The UTC-normalized values are only written to `result.StartedAt` and `result.FinishedAt`, not reused for duration math.
- All current exit paths in `RunChecker` still stamp completion time and duration:
  - nil runner / docker client
  - container spec build failure
  - container create failure
  - file copy failure
  - container start failure
  - wait channel error
  - timeout return
  - normal completion before output-limit / exit-code / JSON parsing result classification
- Keeping `contest/infrastructure/docker_checker_runner.go` inside `reviewedTimeNowFiles` is correct. The module guard only auto-allows `time.Now()` when the same AST call chain ends in `UTC()`. This file still needs a reviewed exception because `startedAt := time.Now()` is intentionally retained for monotonic duration measurement and normalized later via `startedAt.UTC()`.

## Required Re-validation

- None.

## Residual Risk

- The main remaining risk is test coverage, not code shape: a future refactor could accidentally compute duration from UTC-stripped values or miss a new early return without being caught by a behavior-level test in `docker_checker_runner_test.go`.
- `checker_audit.go` still contains a fallback `result.FinishedAt.Sub(result.StartedAt)` path. That fallback now operates on UTC wall-clock timestamps if `Duration` is absent, which is correct for cross-process audit math but naturally loses monotonic semantics. This diff does not worsen that behavior because the primary path still supplies `Duration`.

## Touched Known-debt Status

- No active current-fact structural debt was found on the touched surface.
- The reviewed `time.Now` exception remains current and justified for this file because the monotonic-duration exception is intentional and still protected by the architecture guard.
