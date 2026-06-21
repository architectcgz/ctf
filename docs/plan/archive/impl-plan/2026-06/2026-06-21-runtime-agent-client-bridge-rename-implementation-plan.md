<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# runtime-agent-client-bridge-rename Implementation Plan

**Goal:** 将 runtime-agent gRPC 适配器从模糊的 `agentclient.Bridge` 收口命名为 `agentclient.Client`，让代码表达它是远端 runtime-agent client。

**Architecture:** 保持 `agentclient` 包、`DialContext` / `New` 构造入口、`RuntimeHostExecutor` / `SandboxExecutor` 接口实现不变，只改内部导出类型名、文件名、测试名和 composition 类型引用。该 slice 不改变 RPC 协议、运行时路由、健康检查、mTLS 或 sandbox checker 语义。

**Tech Stack:** Go, gRPC, mTLS, CTF modular monolith backend, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-21-runtime-agent-client-bridge-rename`
- Started At: `2026-06-21T08:50:14Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-21-runtime-agent-client-bridge-rename`
- Branch: `task/2026-06-21-runtime-agent-client-bridge-rename`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `review-pending` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 删除后端 remote runtime adapter 上的 `Bridge` 类型名，改为 `Client`，降低“这是业务 service / bridge 层”的误解。
- Non-Goals:
  - 不拆 `RuntimeHostExecutor` / `SandboxExecutor` 能力接口。
  - 不改 runtime-agent RPC request / response。
  - 不改节点选择、failover、health probe、ACL、container exec 或 checker runner 行为。
  - 不改 challenge 模块中的 transaction bridge 文件。

## Problem Statement

- Current behavior / structure: `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go` 暴露 `type Bridge`，它同时实现 runtime host executor 和 sandbox executor，并被 app composition 用作 remote runtime-agent client。
- Target behavior / structure: 同一 package 中暴露 `type Client`，调用点使用 `*agentclient.Client`，文件和测试名同步使用 `client`。
- Why this task is needed now: 用户明确指出后端仍有一个“service 还是东西叫 bridge”的不清晰命名；当前类型职责是 gRPC client adapter，不是业务 bridge/service。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `code/backend/tests/README.md`
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `harness/policies/reuse-first.yaml`
- Related architecture/contracts:
  - `docs/architecture/backend/03-container-architecture.md`
  - `code/backend/internal/module/container_runtime/ports/runtime_host_executor.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/container_runtime_module.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-container-runtime-tail-migration-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: 触达后端 infrastructure adapter 类型、composition 接线测试和架构 baseline，虽然行为不变，但属于受保护实现面。

## Files

- Create: none
- Modify:
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go` -> rename to `client.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge_integration_test.go` -> rename to `client_integration_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
- Review:
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/module/container_runtime/ports/runtime_host_executor.go`
- Test:
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/client_integration_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `rg -n "\bbridge\b|Bridge|BRIDGE" .`
  - `rg -n "agentclient\.|DialContext\(|\*agentclient\.Bridge" code/backend/internal code/backend/tests`
- Reuse / extend / split / create-new decision: refactor existing; no new adapter, interface, service, or package.
- Owner boundary: `agentclient` owns the remote runtime-agent gRPC client adapter. Composition owns node routing and identity verification. Ports own capability contracts.
- Why this is the narrowest safe surface: only the exported concrete type name and file/test naming change; all capability methods, constructor names, config loading, mTLS setup, stream handling, and interface assertions remain in place.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming` by repository inspection, without user question.
- Why this pass fits: the request was fuzzy; code search clarified that the actionable backend target is `agentclient.Bridge`, not frontend layout bridges or challenge transaction bridges.
- grill-with-docs findings:
  - `docs/architecture/backend/03-container-architecture.md` defines the execution surface as `runtime-agent` / runtime node, not a bridge service.
  - The file implements client-side delegation over gRPC and is consumed by composition as a remote node client.
  - No ADR or architecture fact update is needed because the runtime-agent architecture stays unchanged.
- Plan adjustments after challenge: avoid broad runtime split; use a naming-only refactor to align terms with the documented runtime-agent execution surface.

## Execution Slices

### Slice 1: Rename runtime-agent bridge adapter to client

- Goal: Replace `agentclient.Bridge` with `agentclient.Client` and keep behavior unchanged.
- Dependencies: startup gate created by `scripts/start-implementation.sh`; existing tests provide behavior coverage.
- Files:
  - Create: none
  - Modify:
    - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go` -> `code/backend/internal/module/container_runtime/infrastructure/agentclient/client.go`
    - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge_integration_test.go` -> `code/backend/internal/module/container_runtime/infrastructure/agentclient/client_integration_test.go`
    - `code/backend/internal/app/composition/runtime_node_execution_router.go`
    - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
    - `code/backend/internal/module/architecture_baseline_test.go`
  - Review:
    - `code/backend/internal/app/composition/container_runtime_module.go`
    - `docs/architecture/backend/03-container-architecture.md`
  - Test:
    - `go test -count=1 ./internal/module/container_runtime/infrastructure/agentclient`
    - `go test -count=1 ./internal/app/composition -run 'TestBuildRuntimeNodeClient|TestRuntimeNodeExecutionRouter|TestBuildContainerRuntimeModule'`
    - `go test -count=1 ./internal/module -run TestRuntimeHostExecutorUsageIsRestricted`
- 步骤：
  - [x] 步骤 1：将 `agentclient` 中的 `Bridge` 类型和 receiver 命名改为 `Client`。
  - [x] 步骤 2：将 bridge 相关测试函数和局部变量改为 client 语义命名。
  - [x] 步骤 3：将 composition 中的类型引用从 `*agentclient.Bridge` 改为 `*agentclient.Client`。
  - [x] 步骤 4：将架构 baseline 路径从 `bridge.go` 更新为 `client.go`。
  - [x] 步骤 5：对触达的 Go 文件运行 gofmt。
  - [x] 步骤 6：运行针对性 Go 测试并记录验证证据。
- Validation: targeted package tests plus architecture baseline.
- Review focus: behavior must be unchanged; no temporary type alias; no broad runtime routing redesign.
- Done criteria: no `agentclient.Bridge` references remain; tests pass.

## Impact And Compatibility

- API / DTO: none
- Data / migration: none
- State / cache / queue / event: none
- Runtime / config: none
- Frontend route / state / UX: none
- Docs / contracts: no fact source update required; plan records the naming-only refactor.

## Plan Review / Architecture Fit

- Target owner boundary: `agentclient.Client` is the concrete remote runtime-agent client adapter; capability interfaces remain in `container_runtime/ports`.
- Reuse points / landing zones: reuse existing `agentclient` package and existing `DialContext` / `New` constructors.
- Known structural debt touched: `Bridge` naming ambiguity only.
- How this plan avoids behavior-only convergence: the change directly fixes the confusing exported concrete type name and file/test naming, instead of adding comments around the old name.
- Hidden second-redesign risk: low; a later split of runtime host executor vs sandbox executor can happen independently if responsibilities expand.
- Decision after review: ready for implementation as a single naming refactor slice.

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/03-container-architecture.md`
  - `AGENTS.md`
- Fact sources to update after implementation: none expected.
- Plan-only notes that must not become architecture source: exact local rename steps and validation evidence.
- Archive condition: code reviewed/validated and conclusions fully reflected by code names.

## Validation

- 计划验证范围：运行 agentclient 包测试、runtime node composition 相关测试和 module 架构 baseline。
- 命名检查范围：在触达的 backend surface 搜索 `agentclient.Bridge`、`type Bridge`、`bridge.go`、`bridge_integration_test` 等旧名残留。
- 完成判定：上述测试通过，旧 concrete type 名称不再出现在代码面，且 `git diff --check` 无格式问题。

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test -count=1 ./internal/module/container_runtime/infrastructure/agentclient`
  - `cd code/backend && go test -count=1 ./internal/app/composition -run 'TestBuildRuntimeNodeClient|TestRuntimeNodeExecutionRouter|TestBuildContainerRuntimeModule'`
  - `cd code/backend && go test -count=1 ./internal/module -run TestRuntimeHostExecutorUsageIsRestricted`
- Integration commands: not needed; no runtime behavior or protocol changes.
- Manual checks:
  - `rg -n "\*agentclient\.Bridge|type Bridge|bridge.go|bridge_integration_test" code/backend/internal/module/container_runtime/infrastructure/agentclient code/backend/internal/app/composition code/backend/internal/module/architecture_baseline_test.go`
- Commands intentionally skipped and why: full backend test suite is out of scope for a naming-only internal adapter refactor unless targeted tests reveal broader breakage.

## Validation Evidence

- Command: `cd code/backend && timeout 120s go test -count=1 ./internal/module/container_runtime/infrastructure/agentclient`
  - Result: PASS, `ok ctf-platform/internal/module/container_runtime/infrastructure/agentclient 0.138s`
  - Notes: covers runtime-agent client mTLS delegation for runtime and checker calls.
- Command: `cd code/backend && timeout 180s go test -count=1 ./internal/app/composition -run 'TestBuildRuntimeNodeClient|TestRuntimeNodeExecutionRouter|TestBuildContainerRuntimeModule'`
  - Result: PASS, `ok ctf-platform/internal/app/composition 1.089s`
  - Notes: covers runtime node client construction, identity checks, and container runtime module wiring.
- Command: `cd code/backend && timeout 120s go test -count=1 ./internal/module -run TestRuntimeHostExecutorUsageIsRestricted`
  - Result: PASS, `ok ctf-platform/internal/module 0.016s`
  - Notes: covers architecture baseline path update from `bridge.go` to `client.go`.
- Command: `rg -n "Bridge|bridge|agentclient\.Bridge|bridge\.go|bridge_integration_test" code/backend/internal/module/container_runtime/infrastructure/agentclient code/backend/internal/app/composition code/backend/internal/module/architecture_baseline_test.go`
  - Result: PASS, no matches
  - Notes: confirms old runtime-agent bridge naming is removed from the touched backend surface.
- Command: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: ran API contract unchanged check, backend module/shared/app/test architecture guards, frontend architecture no-op detection, and backend test architecture guard.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS, `no startup-gated changes in diff`
  - Notes: task gate is present; current changed paths were accepted by completion-full.

## Independent Review Handoff

- Review target: `agentclient` type rename and composition references.
- Validation evidence summary: targeted `agentclient`, composition runtime node, and module architecture tests passed; old `Bridge` naming search returned no matches on touched backend surface.
- Architecture / contract inputs:
  - `docs/architecture/backend/03-container-architecture.md`
  - `code/backend/internal/module/container_runtime/ports/runtime_host_executor.go`
- Known risks / review focus:
  - missed `Bridge` type reference
  - behavior change in gRPC client adapter
  - architecture baseline path drift
- Project-local checks to consider:
  - `go test -count=1 ./internal/module/container_runtime/infrastructure/agentclient`
  - `go test -count=1 ./internal/app/composition -run 'TestBuildRuntimeNodeClient|TestRuntimeNodeExecutionRouter|TestBuildContainerRuntimeModule'`
  - `go test -count=1 ./internal/module -run TestRuntimeHostExecutorUsageIsRestricted`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## Rollback / Recovery

- Safe revert boundary: revert this task branch or the touched files.
- Data / config / runtime recovery notes: none.
- Irreversible operations: none.

## Residual Risks

- Risk: `agentclient.Client` still implements two capability interfaces, so the adapter remains broad.
- Why acceptable: this task addresses the confusing bridge/service naming without changing runtime behavior; splitting capabilities would be a separate architectural change with larger test surface.
- Follow-up owner, if any: container runtime owner if future methods continue to expand.
