# Backend Review: Runtime Mapper Island Removal

Date: 2026-06-09
Reviewer: independent code-reviewer subagent
Repository: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-09-runtime-container-runtime-port-file-split/2026-06-09-runtime-instance-port-alias-cleanup`
Task slug: `2026-06-09-runtime-instance-port-alias-cleanup`
Scope: current uncommitted diff against `HEAD`

## Review Target

- Files reviewed:
  - `code/backend/internal/module/runtime/application/commands/response_mapper.go`
  - `code/backend/internal/module/runtime/application/commands/response_mapper_assign.go`
  - `code/backend/internal/module/runtime/application/commands/response_mapper_gen.go`
  - `code/backend/internal/module/runtime/application/queries/response_mapper.go`
  - `code/backend/internal/module/runtime/application/queries/response_mapper_assign.go`
  - `code/backend/internal/module/runtime/application/queries/response_mapper_gen.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/runtime/service_test.go`
  - `code/backend/internal/module/runtime/service_topology_test.go`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-09-runtime-instance-port-alias-cleanup-implementation-plan.md`

## Classification Check

- Agree with `非琐碎任务`.

## Gate Verdict

- `pass`

## Findings

- Blocker: none.

## Material Findings

- None.

## Non-blocking Suggestions

- Low: `code/backend/internal/module/runtime/service_test.go:118-193` now contains a test-only maintenance repository adapter that duplicates the production composition adapter shape. It is correct in this slice, but if the maintenance repository contract grows again, this test adapter and the production adapter can drift independently. A shared test helper or a focused adapter test would reduce that sync cost.
- Low: deleting the six runtime mapper files removes dead `runtime -> instance` noise cleanly, but it also removes one more place where `instance` response shapes were mechanically mirrored. If a future runtime slice needs similar mapping again, it should either live in the actual owning package or come back with a real caller plus goverter source, not another orphan island.

## Missing Validation

- I did not find a focused test that directly exercises `runtimeTestMaintenanceRepository.CreateAWDServiceOperation` in isolation. Current confidence comes from package tests passing through the maintenance service path, plus code inspection that all fields and `operation.ID` backfill are preserved.
- I did not rerun `go generate` because the deleted runtime mapper island had no surviving `go:generate` or `goverter:converter` source file in `runtime/application/{commands,queries}` after this diff, and the policy test already passed. If the team wants belt-and-suspenders evidence, `go generate ./internal/module/runtime/application/...` would still be a harmless extra check.

## Open Questions Or Assumptions

- Assumed the six deleted mapper files were intentionally hand-maintained leftovers rather than required generated artifacts, because:
  - none of them contained `go:generate` directives
  - none of the surviving runtime application files referenced `runtimeResponseMapper`, `instanceResponseMapperImpl`, `ToInstanceResp`, `ToInstanceInfo`, `response_mapper_assign`, or `CopyTime`
  - `TestMapperWrappersFollowGlobalDelegationPolicy` still passed after deletion
- Assumed baseline should remain unchanged in this round because runtime production code still legitimately imports `instance` through `runtime/api/http`, `runtime/domain`, `runtime/application/commands/runtime_cleanup_service.go`, and `runtime/infrastructure/repository.go`.

## Senior Implementation Assessment

- The mapper deletion is justified. The removed files were a dead island:
  - `code/backend/internal/module/runtime/application/commands/response_mapper.go:1-20`
  - `code/backend/internal/module/runtime/application/commands/response_mapper_assign.go:1-7`
  - `code/backend/internal/module/runtime/application/commands/response_mapper_gen.go:1-27`
  - `code/backend/internal/module/runtime/application/queries/response_mapper.go:1-24`
  - `code/backend/internal/module/runtime/application/queries/response_mapper_assign.go:1-7`
  - `code/backend/internal/module/runtime/application/queries/response_mapper_gen.go:1-33`
  These files only referenced instance contracts/ports, had no runtime business caller, and had no remaining goverter entrypoint to justify their existence.
- The runtime maintenance test changes are test-surface-only. `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go:20-203` swaps stub types from runtime-owned aliases to `instanceports` types, matching the current production interface owned by `instance/application/commands/maintenance_service.go`. I did not find any production behavior change hidden in those test edits.
- `code/backend/internal/module/runtime/service_test.go:118-193` maps the maintenance repository contract completely for the current consumer surface:
  - `FindRunningAWDDefenseWorkspaceByInstanceID` reduces the runtime row to `ContainerID`, which is the only field consumed by maintenance recovery logic.
  - `CreateAWDServiceOperation` copies every current field into `runtimeentity.AWDServiceOperation` and preserves `operation.ID = row.ID` backfill semantics.
  - `FinishAWDServiceOperation`, `FinalizeStoppedRuntime`, `RequeueLostRuntime`, and `ListActiveContainerIDs` delegate directly without semantic changes.
- No new unreasonable `runtime -> instance` dependency was introduced. This diff actually removes one runtime application import island while leaving the still-reviewed production dependency surface unchanged. Because runtime still owns repository and HTTP surfaces that speak instance contracts, the existing `runtime -> instance` baseline entry should remain.

## Required Re-validation

- Independent reviewer reran:
  - `go test ./internal/module/runtime/application/... -count=1`
  - `go test ./internal/module/runtime/... -count=1`
  - `go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- Additional local evidence inspected:
  - `rg -n "go:generate|goverter|runtimeResponseMapper|instanceResponseMapperImpl|ToInstanceResp\\(|ToInstanceInfo\\(|response_mapper_assign|CopyTime\\(" code/backend/internal/module/runtime code/backend/internal/module -g '*.go'`
  - `rg -n 'ctf-platform/internal/module/instance/(contracts|ports|entity|infrastructure)|instancecontracts\\.|instanceports\\.|instanceentity\\.' code/backend/internal/module/runtime -g '*.go' --glob '!**/*_test.go'`

## Residual Risk

- The runtime module still carries legitimate production `runtime -> instance` dependencies in API, repository, and cleanup surfaces. This review only confirms that the deleted mapper island was dead noise, not that runtime is now boundary-clean overall.
- The test-only maintenance adapter in `runtime/service_test.go` and the production adapter in app composition now need to evolve in parallel when the instance maintenance repository contract changes.

## Touched Known-debt Status

- This slice reduces the known `runtime -> instance` noise by removing dead runtime application mapper wrappers.
- It does not close the broader reviewed debt that `runtime/infrastructure` and `runtime/api/http` still depend on instance-owned contracts and ports, so the module baseline should remain unchanged in this round.
