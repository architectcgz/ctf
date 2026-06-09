# Running Instance Count Owner Review

Date: 2026-06-09
Scope: `2026-06-09-runtime-instance-port-alias-cleanup`
Reviewer: independent code-reviewer subagent
Verdict: pass with minor issues fixed

## Findings

- Blocker: none.
- Low: `runtime/architecture_test.go` only checked `runtime/ports/http.go` for `CountRunningRepository`, so the guard would miss the same interface if it returned in another runtime port file.
- Residual test gap: repository query semantics and adapter delegation were covered separately, but there was no direct `BuildContainerRuntimeModule -> OpsRuntimeQuery -> instanceinfra.Repository` wiring regression test.

## Resolution

- Updated the runtime ports guard to scan all non-test files under `runtime/ports/*.go`.
- Extended the `BuildContainerRuntimeModule` test to seed a running instance and verify `module.OpsRuntimeQuery.CountRunning(ctx)` reads through the instance-owned repository path.

## Review Basis

The reviewer checked the current diff for owner placement, residual `runtimeports.CountRunningRepository` / `CountRunningService` / `NewCountRunningService` references, composition wiring, repository query semantics, and test coverage.

## Residual Risk

- This slice only moves the running instance count query owner. Other instance-facing repository and proxy ticket capabilities still physically live in `runtime/infrastructure` and remain the next runtime boundary cleanup item.
