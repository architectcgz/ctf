# 2026-06-13 Backend Review Context Logging Contract Round 1

- Review target:
  - Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract`
  - Branch: `task/2026-06-13-backend-context-logging-contract`
  - Commit: `3d3abe1211e0c255c26842371c4eace8c2c5d8e5`
  - Diff basis: `ae161299f4d9db9f6468b037530e831b36d52608..3d3abe1211e0c255c26842371c4eace8c2c5d8e5`
  - Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-context-logging-contract-implementation-plan.md`
  - Files reviewed:
    - `code/backend/internal/middleware/request_id.go`
    - `code/backend/internal/middleware/request_id_test.go`
    - `code/backend/internal/platform/requestctx/request_id.go`
    - `code/backend/internal/platform/requestctx/request_id_test.go`
    - `code/backend/internal/platform/logctx/logger.go`
    - `code/backend/internal/platform/logctx/logger_test.go`
    - `code/backend/internal/module/auth/application/commands/service.go`
    - `code/backend/internal/module/auth/application/commands/context_logging_test.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
    - `code/backend/internal/module/instance/application/commands/context_logging_test.go`
    - `code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
    - `code/backend/internal/module/container_runtime/application/commands/context_logging_test.go`
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
    - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-context-logging-contract-implementation-plan.md`
- Reviewer mode: independent gate review
- Classification check: agree with pipeline classification `非琐碎任务`
- Gate verdict: `pass`

## Findings

No blocking findings or major regressions found in the reviewed diff.

## Material Findings

None.

## Non-blocking Suggestions

1. `code/backend/internal/module/instance/application/commands/maintenance_service.go:142,261,303,307` still leaves `Debug` outside `logctx`. For this slice that is acceptable as a scope choice rather than a regression, because the plan and guardrail intentionally pilot only `Info/Warn/Error`. If later slices need request-scoped debug correlation, extend `logctx` and the pilot guardrail together instead of reintroducing ad hoc raw calls.
2. `code/backend/tests/architecture/test_architecture_test.go:83-97,310-328` uses exact string snippets as a narrow pilot guardrail. That is fine for the current three files, but a broader rollout should move from receiver-name matching to a slightly more structural check so simple local variable aliases cannot bypass it.

## Missing Validation

1. `code/backend/internal/platform/logctx/logger_test.go:13-43` proves request-id attachment and empty-context behavior, but does not directly exercise the documented nil fallbacks (`ctx == nil`, `logger == nil`).
2. The current test set proves middleware propagation, helper behavior, pilot call-site migration, and source guardrails separately. It does not include a focused HTTP-path test that goes through Gin middleware into one pilot service and asserts the emitted structured log contains the middleware-provided `request_id`.

## Senior Implementation Assessment

The `requestctx + logctx` split is the lowest-risk shape for this slice. It keeps request-id ownership at the transport edge, avoids pulling application code toward `internal/infrastructure/logger`, and gives later slices a reusable context-field hook without forcing a full-repo logging rewrite. The middleware canonicalization fix also closes the prior owner-drift issue by ensuring Gin context, request context, request header, response header, and `logctx` all observe the same canonical request id.

## Validation Reviewed

- Existing implementation-context evidence was inspected and considered sufficient for the broad self-checks.
- Independent reviewer reran:
  - `timeout 240s go test ./internal/middleware ./internal/platform/requestctx ./internal/platform/logctx -count=1`
  - `timeout 240s go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - `timeout 240s go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
- All rerun commands passed.

## Required Re-validation

None for the current verdict.

## Residual Risk

- The contract is still a pilot, not a repository-wide logging migration. Files outside the allowlisted pilot set can still emit request-unaware logs until later slices migrate them.
- The pilot guardrail is intentionally narrow and source-based; it protects the selected files but is not yet a general backend logging policy.

## Touched Known-debt Status

The diff touches active owner surfaces in auth, instance maintenance, and container runtime provisioning, but I did not find a current fact-source entry marking these exact touched areas as unresolved structural debt that must be closed in this slice. No debt-based blocker is raised from the current registry.
