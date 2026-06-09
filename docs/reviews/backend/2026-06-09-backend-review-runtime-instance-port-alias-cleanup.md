# Backend Review: Runtime Instance Port Alias Cleanup

Date: 2026-06-09
Reviewer: independent code-reviewer subagent
Repository: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-09-runtime-container-runtime-port-file-split/2026-06-09-runtime-instance-port-alias-cleanup`
Task slug: `2026-06-09-runtime-instance-port-alias-cleanup`
Scope: current uncommitted diff against `HEAD`

## Review Inputs

- Plan read before review:
  - `docs/plan/impl-plan/2026-06-09-runtime-instance-port-alias-cleanup-implementation-plan.md`
- Architecture references:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/design/backend-module-boundary-target.md`
- Review references:
  - `ctf-current-review-status-checks.md`
  - `technical-risk-checks.md`
  - `test-strategy-review.md`

## Classification Check

- Agree with `非琐碎任务`.

## Gate Verdict

- `pass`

## Findings

- Blocker: none.

## Material Findings

- None.

## Non-blocking Suggestions

- Low: `code/backend/internal/module/instance/contracts/access_url.go:10-161` is now a local copy of runtime-side URL rewrite and runtime-details parsing logic. This is a reasonable owner move for this slice, but it introduces duplicated behavior. If the alias host rule, entry-network selection rule, or runtime-details JSON shape changes later, both copies need to move together or a subtle divergence can appear.
- Low: `code/backend/internal/module/instance/ports/ports.go:202-235` intentionally replaces runtime aliases with instance-owned structs/constants, which is the right boundary direction. The remaining cost is string-value duplication for `AWDServiceOperation*` constants and shape duplication for `AWDServiceOperation`; future enum expansion on the persistence side should be mirrored here deliberately rather than drifting silently.
- Low: `code/backend/internal/app/composition/instance_module.go:259-295` currently preserves `CreateAWDServiceOperation` ID backfill semantics by copying `row.ID` back to `operation.ID`. This is correct for the current flow, but it is an adapter-only contract now; a focused composition/maintenance test would make that behavior less review-dependent.

## Missing Validation

- No direct test hit was found for the new instance-owned helpers in `code/backend/internal/module/instance/contracts/access_url.go:10-161`. Current safety comes from:
  - `code/backend/internal/module/instance/infrastructure/awd_target_proxy_repository.go:96` using `ResolveInstanceAliasAccessURL`, with existing repository tests indirectly covering alias rewrite behavior.
  - `code/backend/internal/module/instance/application/commands/instance_service.go:196-203` and `code/backend/internal/module/instance/application/queries/instance_service.go:149-203` passing package tests after switching to `ResolveInstancePublicAccessURL`.
  A small helper-level test set for alias rewrite, public-host rewrite, and container-ID extraction would close this gap.
- No focused adapter test was found for `code/backend/internal/app/composition/instance_module.go:171-312`. Package-level composition tests passed, but there is still no narrow regression test that explicitly proves:
  - `runtime.nodeRouter` path still supports `ListManagedContainers` + `InspectManagedContainer` + `StartContainer`
  - `CreateAWDServiceOperation` continues to backfill `operation.ID`

## Open Questions Or Assumptions

- Assumed `instanceMaintenanceService` only needs `ContainerID` from `FindRunningAWDDefenseWorkspaceByInstanceID`, so reducing `instanceports.AWDDefenseWorkspace` to that single field is sufficient for the current owner boundary.
- Assumed the runtime-details JSON consumed by `ExtractInstanceRuntimeContainerIDs` and alias-IP resolution remains the same schema currently produced by runtime provisioning. This review did not find any touched producer-side schema changes in the current diff.

## Senior Implementation Assessment

- The `instance -> runtime` production import cleanup appears real, not baseline-only:
  - `code/backend/internal/module/instance/ports/ports.go:189-235` no longer aliases runtime-owned types.
  - `code/backend/internal/module/instance/contracts/access_url.go:10-161` absorbs the URL/access/runtime-details helpers previously reached through runtime contracts.
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go:18-28,171-176,369-415,509-522`
  - `code/backend/internal/module/instance/application/queries/instance_service.go:149-203`
  - `code/backend/internal/module/instance/application/commands/instance_service.go:192-209`
  - `code/backend/internal/module/instance/infrastructure/awd_target_proxy_repository.go:93-97`
  Together these remove the previous runtime contract/port/entity imports from production instance code.
- The new instance-owned port types are scoped to instance’s maintenance and access concerns, and I did not find remaining alias-owner leakage in the touched instance package.
- `code/backend/internal/app/composition/instance_module.go:73-89,171-312` keeps the outer-layer dependency in composition, where it belongs. The adapter mapping is complete for the current `InstanceMaintenanceService` surface:
  - container inventory fields are copied one-for-one
  - inspect state fields are copied one-for-one
  - `runtime.nodeRouter` path still provides both inventory and start-container behavior via `newInstanceMaintenanceRuntime(runtime.nodeRouter, runtime.nodeRouter)`
  - repository adapter preserves AWD operation create/finish flows, including operation ID backfill
- `code/backend/internal/module/architecture_baseline_test.go:18-50` deleting `"instance -> runtime"` is supported by both the code diff and independent grep/test evidence; I did not find a remaining production import leak under `internal/module/instance`.

## Required Re-validation

- Independent reviewer reran:
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `go test ./internal/app/composition -count=1`
  - `go test ./internal/module/instance/... -count=1`
  - `rg -n 'ctf-platform/internal/module/runtime|runtimecontracts\\.|runtimeports\\.|runtimeentity\\.' code/backend/internal/module/instance -g '*.go' --glob '!**/*_test.go'`
- User-provided additional passing evidence reviewed as supporting context:
  - `go test ./internal/module -count=1`
  - `python3 scripts/check-docs-consistency.py`
  - `git diff --check`
  - `bash scripts/check-code-changes.sh`
  - `bash scripts/check-workflow-complete.sh`

## Residual Risk

- This slice correctly removes the instance module’s direct runtime imports, but the maintenance repository behavior still physically lives in `runtime/infrastructure` and is only boundary-isolated by composition adapters. That remaining placement is already known debt, not a regression introduced here.
- Because the moved access/runtime-details helpers are duplicated rather than shared through a neutral contract package, future producer/consumer drift is the main thing not mechanically guarded yet.

## Touched Known-debt Status

- The touched surface closes the targeted baseline debt: the stale `instance -> runtime` allowlist entry can now be removed with evidence.
- No additional blocker-level debt was found on the touched surface beyond the already-known repository implementation placement behind composition adapters.
