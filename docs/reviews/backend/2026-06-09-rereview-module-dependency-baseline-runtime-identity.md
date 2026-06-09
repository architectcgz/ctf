# Module Dependency Baseline Runtime Identity Re-review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Task slug: `2026-06-08-module-dependency-baseline`
- Previous blocked review: `docs/reviews/backend/2026-06-09-gate-review-module-dependency-baseline-runtime-identity.md`
- Diff source: current uncommitted diff against `HEAD`, with focus on the blocker fix
- Files re-reviewed:
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/service_instance_teacher_access_test.go`

## Classification Check

- Agree with `非琐碎任务`.

## Gate Verdict

- `pass`

## Findings

- No blocker findings.

## Material Findings

- None.

## Non-blocking Suggestions

- None.

## Missing Validation

- The blocker fix is now directly covered by a focused repository-level behavior test:
  - `code/backend/internal/module/runtime/service_instance_teacher_access_test.go:78`
- Provided implementation-context validation is sufficient for the touched surface.
- As independent reviewer I additionally reran:
  - `cd code/backend && timeout 180s go test ./internal/module/runtime -run TestRepositoryFindUserByIDIgnoresSoftDeletedUsers -count=1`

## Open Questions Or Assumptions

- This re-review only judges the prior blocker and current `runtime -> identity` slice. It does not reopen unrelated remaining baseline edges.

## Senior Implementation Assessment

- The fix is the correct minimal repair.
- `FindUserByID` now keeps the instance-owned projection while explicitly restoring the old soft-delete visibility contract via `Where("deleted_at IS NULL")`.
- The new focused test exercises the exact regression path that previously escaped package-level coverage: soft-delete a user row, then assert `FindUserByID` returns `nil`.
- I did not find a remaining path in the current diff where deleted users are still visible through this repository method.

## Required Re-validation

- None beyond the already executed targeted commands.

## Residual Risk

- The duplicated `InstanceUserRoleStudent` semantic noted in the prior review remains a non-blocking tradeoff, not a correctness issue for this slice.

## Touched Known-debt Status

- The prior touched-surface blocker on deleted-user authorization visibility is resolved in this diff.
- The baseline deletion for `runtime -> identity` remains supported by the real production import removal.
