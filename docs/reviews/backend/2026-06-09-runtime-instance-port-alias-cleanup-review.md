# Runtime Instance Port Alias Cleanup Review

Date: 2026-06-09
Scope: `2026-06-09-runtime-instance-port-alias-cleanup`
Reviewer: independent code-reviewer subagent
Verdict: pass with minor issue fixed

## Findings

- Blocker: none.
- Minor: `code/backend/internal/module/runtime/application/contracts.go` still exposed instance-owned contracts through runtime application aliases. This preserved a secondary non-owner export path even after `runtime/ports` stopped re-exporting `instance/ports`.

## Resolution

- Removed the unused instance-owned aliases from `runtime/application/contracts.go`.
- Kept runtime-owned aliases in `runtime/application/contracts.go`, such as runtime metrics, file, image, and topology contracts.

## Review Validation

Reviewer reran:

```bash
go test ./internal/module/runtime -run TestRuntimePortsDoNotReexportInstancePorts -count=1
go test ./internal/module/runtime/ports ./internal/module/runtime/infrastructure ./internal/module/instance/ports -count=1
go test ./internal/app -run 'TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestAWDDefenseSSHGateway' -count=1
go test ./internal/module/instance/... -count=1
go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1
bash scripts/check-backend-architecture.sh --full
```

All reviewer-run commands passed.

## Residual Risk

- This slice intentionally keeps runtime infrastructure implementing some instance-facing repository and proxy ticket capabilities. That is now explicit through `instance/ports`, but the physical owner of those implementations is still a later architecture decision.
