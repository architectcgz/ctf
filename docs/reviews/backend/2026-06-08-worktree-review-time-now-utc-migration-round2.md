# Backend UTC Migration Worktree Review Round 2

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Task slug: `2026-06-08-backend-architecture-guard-quality`
- Plan input read before review: `docs/plan/archive/impl-plan/2026-06/2026-06-08-backend-architecture-guard-quality-implementation-plan.md`
- Diff source: current uncommitted diff against `HEAD`
- Files reviewed:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/auth/application/commands/cas_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_service.go`
  - `code/backend/internal/module/challenge/application/commands/image_build_service.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`

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
  - `timeout 180s go test ./internal/module -run 'TestTimeNowUsageExceptionsAreCurrent|TestModuleArchitectureBoundaries' -count=1`
  - `timeout 240s go test ./internal/module/auth/application/commands ./internal/module/challenge/application/commands ./internal/module/instance/application/commands ./internal/module/runtime/infrastructure`
  - `timeout 180s bash scripts/check-backend-architecture.sh --full`

## Open Questions Or Assumptions

- The task slug is still bound to the archived backend-architecture-guard plan, so this review used the current uncommitted diff, project `AGENTS.md` UTC contract, and the user-provided focus points as the effective scope source.
- This review assumes the repository-level database/session timezone contract remains UTC as required by `AGENTS.md`; under that contract, the GORM comparisons on `expires_at` and the repository `updated_at` writes remain semantically unchanged after switching the Go-side reference time to UTC.

## Senior Implementation Assessment

- The diff stays inside the intended business-time surface. Timestamps that are written to DB rows, serialized into publish/self-check responses, or used for cross-request runtime visibility/expiry comparisons are now sourced from `time.Now().UTC()`.
- The allowed runtime-local uses remain intact:
  - `auth/application/commands/cas_service.go` still keeps `time.Now().UnixNano()` only for the random password fallback suffix.
  - `instance/application/commands/maintenance_service.go` still keeps plain `time.Now()` for orphan-container grace-period age calculation.
- `runtime/infrastructure/repository.go` does not show a GORM semantic regression from the UTC normalization:
  - `expires_at` filters still compare absolute instants, now with an explicit UTC reference.
  - `UpdateStatusAndReleasePort` now reuses one UTC `now` for both `updated_at` and `destroyed_at`, which is more consistent than taking multiple separate wall-clock reads.
  - AWD defense workspace upsert/revision timestamps, teacher summary reference `now`, and port allocation `updated_at` all align with the repo-wide UTC business-time rule.
- `challenge/application/commands/challenge_service.go` self-check `StartedAt` / `EndedAt` now match the API contract better because those fields are response payload timestamps rather than monotonic duration measurements.
- The removed `reviewedTimeNowFiles` entries are correctly stale. The guard now proves that:
  - `challenge_service.go`
  - `image_build_service.go`
  - `runtime/infrastructure/repository.go`
  no longer contain reviewed-exception-worthy `time.Now()` usage.
- The remaining `reviewedTimeNowFiles` entries for `auth/application/commands/cas_service.go` and `instance/application/commands/maintenance_service.go` are still justified by the two explicitly allowed non-UTC cases above.

## Required Re-validation

- None.

## Residual Risk

- There is still no behavior-specific API assertion that inspects the serialized timezone form of `ChallengeSelfCheckResp` or `ChallengePublishCheckJobResp`. The current risk is low because the source values are now explicit UTC `time.Time` values and the package tests plus architecture guard passed.
- `runtime/infrastructure/repository.go` still depends on callers to pass already-normalized business timestamps into parameters such as `expiresAt` or `finishedAt` where those values are not created inside the repository. This diff does not worsen that surface; it only normalizes the repository-owned reference timestamps.

## Touched Known-debt Status

- No active current-fact structural debt was found on the touched surfaces.
- The touched `time.Now` exception debt is closed for the removed baseline entries; the remaining auth and maintenance exceptions are still current and justified.
