<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 模块反向依赖收口：边1切片2 contest经sandbox capability执行 Implementation Plan

**Goal:** Finish edge 1 convergence by making local AWD checker execution use the neutral `container_runtime` sandbox capability instead of a contest-owned Docker runner.

**Architecture:** `container_runtime/infrastructure` owns Docker sandbox process execution; `contest/infrastructure` owns only `contestports.CheckerRunner` adapters and checker business metadata mapping.

**Tech Stack:** Go backend, Docker SDK adapter, module architecture tests, runtime-agent JSON gRPC bridge from slice 1.

---

## Task Metadata

- Task Slug: `2026-06-10-module-reverse-dependency-convergence-slice-2`
- Started At: `2026-06-10T09:14:11Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-module-reverse-dependency-convergence-slice-2`
- Branch: `task/2026-06-10-module-reverse-dependency-convergence-slice-2`

## Objective And Non-Goals

- Objective:
  - Move Docker sandbox execution implementation from `contest/infrastructure` to `container_runtime/infrastructure` as a `runtimeports.SandboxExecutor`.
  - Wire local runtime nodes and runtime-agent server to construct the container-runtime sandbox executor.
  - Keep contest-facing checker behavior through `contestinfra.SandboxCheckerRunner`.
  - Add a contest architecture guard preventing production contest infrastructure from importing Docker SDK packages.
- Non-Goals:
  - Do not change AWD checker HTTP/script/TCP business behavior.
  - Do not change runtime-agent wire method names, TLS setup, JSON codec, or API DTOs.
  - Do not start `instance -> contest` edge work in this slice.

## Inputs

- Source docs:
  - `docs/plan/impl-plan/2026-06-10-module-reverse-dependency-convergence-plan.md`
  - `docs/design/backend-module-boundary-target.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/module/container_runtime/ports/sandbox_executor.go`
  - `code/backend/internal/module/contest/infrastructure/sandbox_checker_runner.go`
- Related prior work:
  - `7207c3343 refactor(backend): 收口 runtime agent checker 反向依赖`

## Task Classification

- Classification: `非琐碎任务`
- Why: Backend structural refactor touching Docker adapter ownership, app composition wiring, runtime-agent bootstrap, and module architecture guards.

## Files

- Create:
  - `code/backend/internal/module/container_runtime/infrastructure/docker_sandbox_executor.go`
  - container-runtime Docker sandbox executor tests, migrated from contest tests
- Modify:
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go`
  - `code/backend/internal/module/contest/infrastructure/sandbox_checker_runner.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/runtime_module_test.go`
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - `code/backend/internal/module/contest/architecture_test.go`
- Review:
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge_integration_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
- Test:
  - `code/backend/internal/module/contest/architecture_test.go`
  - Docker sandbox executor behavior tests after migration
  - runtime module local checker init failure test

## 复用与 Owner 决策

- Existing patterns searched:
  - Slice 1 `runtimeports.SandboxExecutor` and contest `SandboxCheckerRunner` adapter.
  - Existing contest `DockerCheckerRunner` Docker SDK implementation and tests.
  - Local/remote runtime node wiring in `app/composition/runtime_module.go` and `runtime_node_execution_router.go`.
- Reuse / extend / split / create-new decision:
  - Reuse the existing Docker runner implementation by moving its owner to container_runtime and adapting types from `contestports.CheckerRun*` to `runtimeports.SandboxExec*`.
  - Keep contest behavior via adapter mapping rather than duplicating Docker execution code.
  - Preserve existing Docker sandbox behavior tests by moving them to the new owner package.
- Owner boundary:
  - `container_runtime`: Docker SDK, container spec, file archive, logs, timeout, resource limits, and sandbox execution status.
  - `contest`: checker business interface, metadata labels, and mapping sandbox result reasons back to checker reasons.
- Why this is the narrowest safe surface:
  - It removes the remaining owner-mixed Docker implementation without changing checker callers or public API behavior.
  - It avoids cross-module private imports by constructing container_runtime infrastructure in app/bootstrap, then passing only `runtimeports.SandboxExecutor` to contest adapters.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: This is a structural owner migration with an existing plan; the main decisions are ownership, compatibility, and validation placement.
- Evidence inspected:
  - Slice 1 already removed production `container_runtime -> contest` imports and baseline entry.
  - Local runtime nodes still call `newLocalCheckerRunner(cfg.Contest.AWD.CheckerSandbox)`, which constructs contest-owned Docker execution.
  - `runtime_agent.go` still constructs a contest checker runner and wraps it back into a sandbox executor for the server.
  - Existing Docker behavior tests live under contest infrastructure and should move with the owner.
- grill-with-docs findings:
  - No user question blocks execution. The code confirms slice 2 should close the local Docker owner gap, not repeat slice 1.
  - Do not make `contest/infrastructure` import `container_runtime/infrastructure`; app composition/bootstrap are the correct construction sites for private adapters.
- Plan adjustments after challenge:
  - Add a red architecture test to forbid Docker SDK imports from production contest infrastructure.
  - Keep wire compatibility by preserving the slice 1 agentcontracts method path.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/contest -run TestInfrastructureDoesNotOwnDockerSandboxExecution -count=1` (red before implementation, green after)
  - `cd code/backend && go test ./internal/module/container_runtime/infrastructure -count=1`
  - `cd code/backend && go test ./internal/module/contest/infrastructure -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter' -count=1`
  - `cd code/backend && go test ./internal/bootstrap -count=1`
  - `cd code/backend && go test ./internal/module/container_runtime/... -count=1`
  - `cd code/backend && go test ./internal/module/contest/... -count=1`
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
- Manual checks:
  - `rg 'github.com/docker/docker' code/backend/internal/module/contest/infrastructure -g '*.go' -g '!**/*_test.go'` should return no production matches.
  - `rg 'NewLocalCheckerRunner|NewDockerCheckerRunner' code/backend/internal -g '*.go'` should show no production wiring through old contest-owned Docker runner.
- Review focus:
  - container_runtime must not import contest.
  - contest must not import container_runtime infrastructure.
  - checker metadata labels must remain contest-owned and opaque to container_runtime.
  - checker result reasons must preserve contest-facing values through adapter mapping.

## Validation Evidence

- `cd code/backend && go test ./internal/module/contest -run TestInfrastructureDoesNotOwnDockerSandboxExecution -count=1`: red before migration while production contest infrastructure still imported Docker SDK; green after Docker executor owner moved to container_runtime.
- `cd code/backend && go test ./internal/module/container_runtime/infrastructure ./internal/module/contest/infrastructure ./internal/module/contest -count=1`: pass.
- `cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter' -count=1`: pass.
- `cd code/backend && go test ./internal/bootstrap -count=1`: pass.
- `cd code/backend && go test ./internal/module/container_runtime/... -count=1`: pass.
- `cd code/backend && go test ./internal/module/contest/... -count=1`: pass.
- `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`: pass.
- `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`: pass.
- `bash scripts/run-workflow-stage.sh completion-full`: pass.
- `bash scripts/run-workflow-stage.sh pre-commit-quick`: pass.

## Checklist

- [x] Add red architecture guard for contest Docker SDK ownership.
- [x] Move Docker sandbox executor implementation to container_runtime.
- [x] Rewrite contest Docker runner surface to adapter-only behavior or remove production use.
- [x] Update local runtime node and runtime-agent bootstrap wiring.
- [x] Move/adjust Docker sandbox tests under the new owner and keep contest adapter tests.
- [x] Run validation and record evidence.
