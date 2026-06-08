# Time Now Guard Runtime-only Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Task slug: `2026-06-08-backend-architecture-guard-quality`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-08-backend-architecture-guard-quality-implementation-plan.md`
- Diff source: current uncommitted worktree diff
- Files reviewed:
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`

## Classification Check

- Agree with `非琐碎任务`.

## Gate Verdict

- `pass`

## Findings

- No findings.

## Material Findings

- None.

## Non-blocking Suggestions

- The stale baseline cleanup itself looks consistent with the current implementation. The removed entries I spot-checked now match direct runtime-only patterns (`Unix/UnixNano`, `time.Since`, `Sub`, `SetDeadline`, deadline comparisons), while `runtime/application/commands/provisioning_service.go` still remains blocked because `stageStartedAt` is passed into `logTopologyStage`.

## Missing Validation

- No additional blocker-level validation gap.
- Independent re-review validation:
  - `timeout 180s go test ./internal/module -run 'TestTimeNowUsage' -count=1`

## Open Questions Or Assumptions

- Assumption: this review follows the intent you gave in the handoff, namely that the guard should only allow AST-provable local runtime-only patterns and should not start doing helper-level or cross-function inference.

## Senior Implementation Assessment

- The blocker is closed.
- `directTimeNowCallInAssignedExpr` now matches the intended boundary: it accepts direct local call chains such as `time.Now()` and `time.Now().Add(timeout)`, but does not infer through helper calls like `wrap(...)`.
- The new tests cover both missing edges from the prior review:
  - wrapped runtime-only values are rejected
  - raw assigned business time still fails on direct field write

## Required Re-validation

- No further re-validation required for this review pass.

## Residual Risk

- I did not find evidence that the 12 removed baseline entries are currently stale for the wrong reason; their visible `time.Now()` usages still match the intended direct runtime-only cases.
- Remaining risk is the expected one already documented by the task: cross-function flows like `runtime/application/commands/provisioning_service.go` still need reviewed exceptions because this guard intentionally does not do inter-procedural analysis.

## Touched Known-debt Status

- No separate active structural-debt record was found for these two files.
- The touched surface includes the explicit design boundary from this task handoff ("no cross-function inference"), and the current implementation now respects that boundary.
