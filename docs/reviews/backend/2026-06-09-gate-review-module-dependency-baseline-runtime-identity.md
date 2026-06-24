# Module Dependency Baseline Runtime Identity Gate Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Task slug: `2026-06-08-module-dependency-baseline`
- Plan input read before review: `docs/plan/archive/impl-plan/2026-06/2026-06-08-module-dependency-baseline-implementation-plan.md`
- Diff source: current uncommitted diff against `HEAD`
- Files reviewed:
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/runtime/ports/http.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/application/instance_service_test.go`
  - `code/backend/internal/module/runtime/ports/http_context_contract_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`

## Classification Check

- Agree with `非琐碎任务`.

## Gate Verdict

- `blocked`

## Findings

1. Blocking: `FindUserByID` lost the previous soft-delete semantics, so deleted users now look like valid requesters/owners.
   - Location:
     - `code/backend/internal/module/runtime/infrastructure/repository.go:99`
     - Affected callers:
       - `code/backend/internal/module/instance/application/commands/instance_service.go:119`
       - `code/backend/internal/module/instance/application/queries/instance_service.go:97`
   - What changed:
     - The old implementation queried `identitycontracts.User`, which carries `gorm.DeletedAt` and therefore applied GORM's implicit `deleted_at IS NULL`.
     - The new implementation uses `Table("users").Select("id, role, class_name").First(&runtimeports.InstanceUser)`, which does not apply the soft-delete scope.
   - Trigger:
     - A teacher requester is soft-deleted but still reaches the service via stale auth/session context.
     - An instance owner user is soft-deleted.
   - Impact:
     - `ListTeacherInstances` can still scope by a deleted teacher's `class_name` instead of returning `ErrUnauthorized` or an empty result path.
     - `DestroyTeacherInstance` can authorize class checks against deleted requester/owner rows and allow operations that previously would have been denied.
     - This is a real behavior regression, not just a contract refactor.
   - Evidence:
     - Repo code inspection confirms the missing `deleted_at IS NULL`.
     - Independent GORM repro run during review confirmed the behavior difference:
       - querying a soft-deleted `User` model with `First(&user)` returns not found
       - querying `Table("users").Select(...).First(&projection)` returns the deleted row
   - Fix direction:
     - Preserve the projection, but restore soft-delete filtering explicitly, for example by adding `Where("id = ? AND deleted_at IS NULL", userID)` or by querying through a model that still carries the soft-delete scope and scanning into the projection.

## Material Findings

- Restore deleted-user filtering in `runtime/infrastructure.Repository.FindUserByID`, then re-run the affected runtime/instance tests plus the architecture baseline check.

## Non-blocking Suggestions

- `code/backend/internal/module/instance/ports/ports.go:44`
  - `InstanceUser` is sufficient for the current production call sites: only `ClassName` is read today, and the new projection still carries `ID` plus `Role` if future instance-owned checks need it.
  - `InstanceUserRoleStudent = "student"` is acceptable as the narrowest owner-local stopgap because the runtime side already consumes `instance/ports`, and the docs position `ports` as consumer-owned contracts. The tradeoff is duplicated role semantics. If more role-based branching lands on this surface, prefer converging on a richer instance-owned predicate/projection instead of spreading more copied identity role strings.

## Missing Validation

- No current test covers `FindUserByID` against a soft-deleted user row after this refactor.
- The provided validation evidence is otherwise aligned with the touched surface:
  - `go test ./internal/module/instance/ports ./internal/module/instance/application/commands ./internal/module/instance/application/queries -count=1`
  - `go test ./internal/module/runtime/ports ./internal/module/runtime/infrastructure -count=1`
  - `go test ./internal/module/runtime/application -run 'TestInstanceServiceDestroyTeacherInstancePropagatesContextToRepository' -count=1`
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- As independent reviewer I additionally ran:
  - `cd code/backend && timeout 180s go test ./internal/module/runtime/infrastructure -run 'Test.*(FindUser|TeacherInstance|Destroy).*' -count=1`
  - a standalone GORM repro proving `Table("users").Select(...).First(&projection)` returns soft-deleted rows while `First(&identity-like model)` does not

## Open Questions Or Assumptions

- This review assumes the existing user soft-delete contract is intentional and relied upon by teacher-instance authorization flows. That assumption matches current `identity/entity.User` shape, the previous implementation, and the database design doc's soft-delete convention for `users`.
- I did not find evidence that this diff caused the known unrelated failure in `TestInstanceServiceListTeacherInstancesAppliesStatusAndPagination`; the touched test changes are signature-only and do not alter summary logic.

## Senior Implementation Assessment

- The owner direction is mostly correct:
  - moving the requester/owner projection to `instance/ports`
  - re-exporting it through `runtime/ports`
  - removing the physical `runtime -> identity` import
  - deleting the baseline entry only after the import disappears
- That is the simpler maintainable shape for this batch, because it keeps the capability implemented in runtime but makes the data contract owned by the consuming instance boundary.
- The main problem is the repository rewrite picked a lower-level GORM query shape without preserving the old behavior contract. This is exactly the kind of regression that module-boundary cleanup can introduce if behavior parity is not locked down with a focused repository test.

## Required Re-validation

- `cd code/backend && timeout 180s go test ./internal/module/runtime/infrastructure -count=1`
- `cd code/backend && timeout 180s go test ./internal/module/instance/application/commands ./internal/module/instance/application/queries -count=1`
- `cd code/backend && timeout 180s go test ./internal/module/runtime/application -run 'TestInstanceServiceDestroyTeacherInstancePropagatesContextToRepository' -count=1`
- `cd code/backend && timeout 180s go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
- `cd code/backend && timeout 180s bash scripts/check-backend-architecture.sh --full`

## Residual Risk

- Even after fixing the soft-delete regression, there will still be no direct test asserting `FindUserByID` parity for deleted users unless one is added.
- The broader task still has remaining module baseline edges (`runtime -> instance`, `instance -> identity`, `practice -> identity`, etc.); this review only judges the current `runtime -> identity` slice.

## Touched Known-debt Status

- The touched surface is a tracked structural cleanup area from the active implementation plan.
- This slice does remove the real physical `runtime -> identity` import and makes the baseline deletion honest.
- It is still blocked because the touched authorization surface regressed behavior while closing the dependency edge.
