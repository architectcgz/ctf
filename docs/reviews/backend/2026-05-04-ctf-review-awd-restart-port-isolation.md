# AWD 重启端口隔离 Review

- Review target: `/home/azhi/workspace/projects/.worktrees/ctf-awd-restart-port-fix`
- Branch: `fix/awd-restart-port-isolation`
- Base: `6418b705`
- Diff source: local working tree diff
- Files reviewed:
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
  - `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
  - `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
  - `code/backend/internal/module/practice/infrastructure/repository.go`
  - `code/backend/internal/module/practice/infrastructure/repository_test.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `code/backend/internal/module/practice/ports/instance_context_contract_test.go`
  - `docs/plan/impl-plan/2026-05-04-awd-restart-port-isolation-implementation-plan.md`

## Classification Check

Agree with pipeline classification: non-trivial backend bugfix. The change touches restart state transition, repository transaction behavior, and port allocation ownership.

## Gate Verdict

Pass.

## Findings

No material findings.

## Senior Implementation Assessment

The implementation keeps the ownership boundary clear:

- `instance_start_service.go` decides whether the scope needs a published host port.
- `repository.go` owns transactional reset, port allocation validation, and release.
- AWD service restart now passes `preserveHostPort=false`, while non-AWD restart behavior remains capable of preserving the port.

The approach is lower risk than a broad cleanup job or schema change because it makes the restart path self-healing for historical AWD instances without changing normal instance allocation semantics.

## Required Re-validation

Already run before review:

```bash
go test ./internal/module/practice/application/commands ./internal/module/practice/infrastructure ./internal/module/practice/ports
go test ./internal/module/practice/...
```

Re-run `go test ./internal/module/practice/...` after any review-driven changes.

## Residual Risk

- Existing running AWD instances that still have `host_port` are not proactively rewritten; they are corrected on their next restart.
- If an external deployment later introduces a browser-facing AWD defense proxy that depends on host ports, that should be modeled as a separate stable proxy mapping, not by reusing instance `host_port`.
