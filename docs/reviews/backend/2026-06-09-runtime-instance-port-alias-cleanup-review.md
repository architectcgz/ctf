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

- This slice intentionally keeps runtime infrastructure implementing some instance-facing repository capabilities. Proxy ticket storage and AWD ticket scope reader ownership have since moved to `instance/infrastructure`.

## Follow-up Review: Proxy Ticket Infrastructure Owner

Date: 2026-06-09
Reviewer: independent code-reviewer subagent
Verdict: pass after blocker fixes

### Findings

- Blocker fixed: app composition structure guards still expected `runtimeInstanceRepository` and `buildRuntimeProxyTicketService(root, repo)` after proxy ticket owner moved to instance infrastructure.
- Blocker fixed: runtime infrastructure guard only checked old symbol strings and did not reject renamed implementations of the same proxy ticket interfaces.
- Blocker fixed: instance infrastructure briefly used runtime persistence aliases for AWD workspace and scope-control test fixtures; these were replaced with local SQL constants and local test rows.

### Resolution

- Updated app structure guards to require `instanceinfra.NewRepository(root.DB())` and `buildRuntimeProxyTicketService(root, instanceRepo)`, and to reject `runtimeInstanceRepository`.
- Replaced the runtime infrastructure guard with method-level checks for `SaveProxyTicket`, `FindProxyTicket`, `FindAWDTargetProxyScope`, and `FindAWDDefenseSSHScope`, plus a proxy ticket key guard.
- Kept `runtimecontracts.ResolveRuntimeAliasAccessURL` as a runtime access URL contract function, not as a persistence alias.

### Review Validation

Reviewer reran:

```bash
go test ./internal/app -run 'TestRuntimeModuleUsesTypedDeps|TestBuildInstanceModuleDelegatesToSubBuilders' -count=1
go test ./internal/module/runtime -run TestRuntimeInfrastructureDoesNotOwnInstanceProxyTickets -count=1
```

Both reviewer-run commands passed, and the reviewer confirmed blocker count was zero.
