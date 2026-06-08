# Backend UTC Migration Worktree Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Task slug: `2026-06-08-backend-architecture-guard-quality`
- Plan input read before review: `docs/plan/archive/impl-plan/2026-06/2026-06-08-backend-architecture-guard-quality-implementation-plan.md`
- Diff source: current uncommitted diff against `HEAD`
- Files reviewed:
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/assessment/application/commands/profile_service.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/infrastructure/report_repository.go`
  - `code/backend/internal/module/assessment/infrastructure/repository.go`
  - `code/backend/internal/module/challenge/application/commands/topology_service.go`
  - `code/backend/internal/module/challenge/application/commands/writeup_service.go`
  - `code/backend/internal/module/challenge/application/queries/writeup_service.go`
  - `code/backend/internal/module/instance/application/commands/instance_service.go`
  - `code/backend/internal/module/instance/application/queries/instance_service.go`
  - `code/backend/internal/module/ops/application/queries/risk_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/overview_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/service.go`

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

- Current package coverage proves compile/runtime correctness for the touched packages, and the architecture guard proves the deleted `reviewedTimeNowFiles` entries are stale.
- There is still no focused unit test that explicitly captures the downstream `time.Time` argument location for:
  - `challenge/application/queries/writeup_service.go`
  - `instance/application/commands/instance_service.go`
  - `instance/application/queries/instance_service.go`
  - `ops/application/queries/risk_service.go`
  These paths are low risk because the diff is a direct `time.Now().UTC()` normalization and package tests passed, but the UTC contract is currently protected more by architecture guard plus code review than by behavior-specific assertions.

## Open Questions Or Assumptions

- The task slug currently points at an archived implementation plan about backend architecture guard helper extraction, not this UTC migration batch. This review therefore used the current uncommitted diff, project `AGENTS.md` backend UTC contract, and the user-provided review focus as the effective scope source.
- Assumed repository/database timestamps are already normalized to UTC per project contract, so converting the reference `now`/`since` timestamps to UTC does not introduce mixed-zone comparisons downstream.

## Senior Implementation Assessment

- The diff stays within the intended surface: business timestamps written to entities, persistence `updated_at`, cross-request comparison reference times, report TTL/generated timestamps, and instance visibility checks now all normalize to UTC.
- The diff does not touch the allowed local/monotonic cases. Existing runtime duration measurement in `profile_service.go` still uses `time.Now()`/`time.Since`, and the temporary filename suffix in `report_service.go` still uses `time.Now().Unix()`.
- `report_service.go` remains semantically consistent after the change:
  - contest export `GeneratedAt` now comes from `reportNow()`
  - report expiry still comes from `reportNow().Add(FileTTL)`
  - they are separate timestamps taken from the same UTC time source, which matches the existing behavior where generation time and ready/expiry time are not required to be identical.
- The assessment/teaching query day-window pattern does not regress timezone semantics. After `time.Now().UTC().AddDate(...)`, `time.Date(..., since.Location())` now deterministically reconstructs midnight in `time.UTC`, which is aligned with the repository-wide UTC business-time rule.
- The removed `reviewedTimeNowFiles` entries are correctly stale. Each removed file now either has same-call-chain `time.Now().UTC()` usage that the guard accepts automatically, or no longer contains reviewed-exception-worthy `time.Now()` usage.

## Required Re-validation

- Independent reviewer reran:
  - `timeout 180s go test ./internal/module -run 'TestTimeNowUsageExceptionsAreCurrent|TestModuleArchitectureBoundaries' -count=1`
  - `timeout 240s go test ./internal/module/assessment/application/commands ./internal/module/assessment/infrastructure ./internal/module/challenge/application/commands ./internal/module/challenge/application/queries ./internal/module/ops/application/queries ./internal/module/instance/application/commands ./internal/module/instance/application/queries ./internal/module/teaching_query/application/queries`
  - `timeout 180s bash scripts/check-backend-architecture.sh --full`
- All passed.

## Residual Risk

- Some touched query services do not have behavior-specific UTC assertions, so a future refactor could reintroduce local-time `time.Now()` without immediately failing a unit test unless it also trips the architecture guard.
- The stale task-plan binding should be corrected in workflow metadata if this slug is expected to remain the canonical record for this UTC migration batch.

## Touched Known-debt Status

- No active current-fact structural debt was found on the touched surfaces.
- The touched `time.Now()` exception debt is fully closed for the removed baseline entries: the guard now proves those exceptions are stale.
