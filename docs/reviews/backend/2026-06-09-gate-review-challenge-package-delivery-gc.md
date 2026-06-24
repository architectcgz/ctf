# 2026-06-09 Gate Review - challenge-package-delivery-gc

## Review target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-challenge-package-delivery-gc`
- Branch: `task/2026-06-09-challenge-package-delivery-gc`
- Task slug: `2026-06-09-challenge-package-delivery-gc`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-09-challenge-package-delivery-gc-implementation-plan.md`
- Diff source: current worktree diff plus newly added files
- Files reviewed:
  - `code/backend/internal/module/challenge/application/commands/artifact_gc_service.go`
  - `code/backend/internal/module/challenge/application/commands/artifact_gc_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service.go`
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service_test.go`
  - `code/backend/internal/module/challenge/infrastructure/artifact_reference_repository.go`
  - `code/backend/internal/module/challenge/infrastructure/artifact_reference_repository_test.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/module/challenge/api/http/awd_challenge_handler.go`
  - `code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go`
  - `code/backend/internal/module/challenge/api/http/awd_challenge_handler_test.go`
  - `code/backend/internal/module/challenge/runtime/module.go`
  - `code/backend/internal/module/challenge/runtime/wiring.go`
  - `code/backend/internal/module/challenge/architecture_test.go`
  - `code/backend/internal/module/challenge/infrastructure/registry_client.go`
  - `code/backend/internal/module/challenge/infrastructure/registry_client_test.go`
  - `code/backend/cmd/storage-gc/main.go`
  - `docs/contracts/api-contract-v1.md`

## Classification check

- Agree with the existing `非琐碎任务` classification and independent `code-workflow` review gate.

## Gate verdict

- `pass`

## Findings

- No material findings in the reviewed diff.
- Blocker 1 is closed: `pathMatchesProtectedReference()` in `artifact_gc_service.go` now treats both `candidate inside protected` and `protected inside candidate` as referenced, which protects the GC candidate parent directory of active `image_build_jobs.source_dir`.
- Blocker 2 is closed: production HTTP preview/commit entrypoints in `handler.go` and `awd_challenge_handler.go` now route through `PackageDeliveryService`, so the facade is no longer dead code.

## Material findings

- None.

## Non-blocking suggestions

- None.

## Missing validation

- None required before completion from this review.

## Senior implementation assessment

- The current shape is the lowest-risk closure for this slice. `PackageDeliveryService` is intentionally thin and delegates straight back into the existing Jeopardy and AWD import owners, so it changes the production call path without reopening HTTP contract or error-mapping logic.
- The GC fix is also appropriately narrow. It closes the active-build-parent deletion hole in the actual candidate/reference shape used by `persistImportedImageBuildSource()` and `image_build_jobs.source_dir`, without widening the deletion surface or introducing a second source of path ownership.

## Required re-validation

- None before completion.
- Reviewer independently reran:
  - `timeout 120s go test ./internal/module/challenge/api/http -run 'TestHandler.*Import|TestAWDChallengeHandler.*Import' -count=1`
  - `timeout 120s go test ./internal/module/challenge/application/commands -run 'TestArtifactGC|TestPackageDelivery' -count=1`
  - `timeout 120s go test ./internal/module/challenge -run 'TestRuntimeOwnsChallengeWiring' -count=1`

## Residual risk

- `storage-gc --execute` still lacks an end-to-end reviewer rerun against a realistic on-disk fixture with DB-backed references. The focused command tests cover the regression and the root guards, so this is residual operational confidence risk rather than a correctness blocker.
- The GC plan is based on the current DB/reference snapshot. If future code paths introduce artifact owners outside `ArtifactReferenceRepository`, this protection model would need to be extended in the same owner location.

## Touched known-debt status

- No currently tracked touched-surface structural debt was left open by this diff.
