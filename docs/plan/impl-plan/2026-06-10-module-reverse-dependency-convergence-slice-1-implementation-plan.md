<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 模块反向依赖收口：边1切片1 sandbox executor契约 Implementation Plan

**Goal:** Remove the production `container_runtime -> contest` dependency from the runtime-agent checker execution path by introducing a neutral container-runtime sandbox execution contract.

**Architecture:** `container_runtime` owns the sandbox execution port and runtime-agent wire types; `contest` keeps the AWD checker business contract and maps it to/from the neutral sandbox executor at the composition edge.

**Tech Stack:** Go backend, manual gRPC JSON codec under `container_runtime/agentcontracts`, module architecture tests.

---

## Task Metadata

- Task Slug: `2026-06-10-module-reverse-dependency-convergence-slice-1`
- Started At: `2026-06-10T08:48:05Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-module-reverse-dependency-convergence-slice-1`
- Branch: `task/2026-06-10-module-reverse-dependency-convergence-slice-1`

## Objective And Non-Goals

- Objective:
  - Introduce `container_runtime/ports.SandboxExecutor` and neutral `SandboxExecJob` / `SandboxExecResult` types.
  - Change runtime-agent client/server/messages so production `container_runtime` code no longer imports `contest`.
  - Add contest-side adapters that preserve the existing `contestports.CheckerRunner` behavior while calling the neutral sandbox executor.
  - Remove the stale `container_runtime -> contest` dependency baseline entry only if the real production import disappears in this slice.
- Non-Goals:
  - Do not change AWD checker pass/fail semantics, JSON output parsing, timeout behavior, Docker sandbox behavior, or runtime-agent TLS / transport setup.
  - Do not migrate `DockerCheckerRunner` itself out of `contest/infrastructure` in this slice; that is a later owner-cleanup step.
  - Do not start the `instance -> contest` edge work; the high-level plan says it depends on the runtime residual state owner split block 3.

## Inputs

- Source docs:
  - `docs/plan/impl-plan/2026-06-10-module-reverse-dependency-convergence-plan.md`
  - `docs/design/backend-module-boundary-target.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/container_runtime/architecture_test.go`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-09-container-runtime-module-boundary-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-09-runtime-container-runtime-port-file-split-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: Cross-module backend refactor touching runtime-agent wire contracts, composition wiring, contest checker adapters, and architecture guardrails.

## Files

- Create:
  - `code/backend/internal/module/container_runtime/ports/sandbox_executor.go`
  - contest-side checker/sandbox adapter file if no existing owner fits
- Modify:
  - `code/backend/internal/module/container_runtime/agentcontracts/messages.go`
  - `code/backend/internal/module/container_runtime/agentcontracts/service.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - affected tests and architecture baseline
- Review:
  - `code/backend/internal/module/contest/ports/checker_runner.go`
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge_integration_test.go`
- Test:
  - `code/backend/internal/module/container_runtime/architecture_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge_integration_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Existing `runtimeports.RuntimeHostExecutor` capability boundary.
  - Existing manual runtime-agent JSON gRPC service and bridge/server tests.
  - Existing `contestports.CheckerRunner` business contract and `DockerCheckerRunner`.
- Reuse / extend / split / create-new decision:
  - Create a new narrow `SandboxExecutor` port instead of widening `RuntimeHostExecutor`; sandbox process execution is used by checker execution but is not equivalent to host lifecycle/file/ACL operations.
  - Keep runtime-agent wire method path compatible while changing Go request/response payloads to neutral sandbox types.
  - Add contest-owned adapters for `CheckerRunner <-> SandboxExecutor` mapping instead of making `container_runtime` import contest business ports.
- Owner boundary:
  - `container_runtime`: neutral sandbox process execution capability and agent transport.
  - `contest`: AWD checker semantics, metadata labels, checker reason interpretation, and business-facing `CheckerRunner`.
  - `app/composition` / `bootstrap`: wires contest adapters to runtime capabilities.
- Why this is the narrowest safe surface:
  - It removes the reverse import at the transport boundary without rewriting checker Docker execution internals or changing AWD behavior.
  - The remaining contest-side adapter is an explicit transitional owner point for the next slice.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: This is a structural refactor with an existing plan; the main risk is owner placement and compatibility, not debugging a runtime failure.
- Evidence inspected:
  - `container_runtime/agentcontracts/messages.go` embeds `contestports.CheckerRunJob` / `CheckerRunResult`.
  - `agentclient.Bridge` implements `contestports.CheckerRunner`.
  - `agentserver.Service` holds `contestports.CheckerRunner`.
  - `runtime_node_execution_router` and `runtime_module` expose contest-facing checker runners from runtime clients.
  - `DockerCheckerRunner` remains the existing Docker sandbox implementation behind contest's checker contract.
- grill-with-docs findings:
  - No user question is blocking this slice; code confirms the high-level plan's root cause.
  - Avoid renaming the runtime-agent wire method unless required; compatibility goal is type ownership and import direction, not transport shape churn.
  - The term for the new neutral capability is `SandboxExecutor`; the contest-facing term remains `CheckerRunner`.
- Plan adjustments after challenge:
  - Slice 1 will not just add types; it will remove production `container_runtime -> contest` imports and the stale baseline entry if tests confirm the edge is gone.
  - `DockerCheckerRunner` migration is deferred, but only behind an explicit contest adapter surface, not by leaving business ports in container_runtime.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/container_runtime -run TestContainerRuntimeDoesNotDependOnBusinessOwnerModules -count=1` (red before implementation)
  - `cd code/backend && go test ./internal/module/container_runtime/... -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestBuildContainerRuntimeModule|TestBuildContainerRuntimeModuleFailsWhenRemoteRuntimeAgentDialFails' -count=1`
  - `cd code/backend && go test ./internal/bootstrap -count=1`
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`
- Manual checks:
  - Confirm `rg 'internal/module/contest' code/backend/internal/module/container_runtime -g '*.go'` returns only tests, or no production matches.
  - Confirm runtime-agent JSON method path remains compatible if method path is preserved.
- Review focus:
  - `container_runtime` must not interpret contest/team/service metadata; any business labels are set in contest-owned adapter code.
  - Agent client/server should expose neutral sandbox execution to app composition.
  - Checker result semantics and existing tests should remain unchanged.

## Validation Evidence

- `cd code/backend && go test ./internal/module/container_runtime -run TestContainerRuntimeDoesNotDependOnBusinessOwnerModules -count=1`
  - Red before implementation: failed on `agentcontracts/messages.go` importing `contest/ports`.
  - Green after implementation: passed.
- `cd code/backend && go test ./internal/module/container_runtime/... -count=1`: passed.
- `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - Failed once with stale `container_runtime -> contest` baseline after production imports were removed.
  - Passed after removing the stale baseline entry.
- `cd code/backend && go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestBuildContainerRuntimeModule|TestBuildContainerRuntimeModuleFailsWhenRemoteRuntimeAgentDialFails' -count=1`: passed.
- `cd code/backend && go test ./internal/bootstrap -count=1`: passed.
- `cd code/backend && go test ./internal/app -run 'TestNewRouterFailsWhenRemoteRuntimeAgentDialFails|Test.*RuntimeAgent|Test.*Composition.*Deps' -count=1`: passed.
- `cd code/backend && go test ./internal/module/contest/infrastructure -count=1`: passed.
- `cd code/backend && go test ./internal/module/contest/... -count=1`: passed.
- `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`: passed.
- `bash scripts/run-workflow-stage.sh completion-full`: passed.
- `bash scripts/run-workflow-stage.sh pre-commit-quick`: passed.

Independent review gate: not satisfied in this session because no independent reviewer/subagent tool is available. Same-context self-check found no material blocker, but it does not count as the workflow independent gate.

## Checklist

- [x] Add the failing architecture guard for `container_runtime` business-owner imports.
- [x] Define `runtimeports.SandboxExecutor` and neutral sandbox job/result types.
- [x] Update runtime-agent message/service/client/server code to use sandbox types.
- [x] Add contest-owned adapters between `contestports.CheckerRunner` and `runtimeports.SandboxExecutor`.
- [x] Update composition/bootstrap wiring and focused tests.
- [x] Remove stale `container_runtime -> contest` baseline only after the real import edge disappears.
- [x] Run validation commands and record results.
