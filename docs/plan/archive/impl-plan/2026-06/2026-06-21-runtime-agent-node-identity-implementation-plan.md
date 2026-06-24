<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# Runtime Agent Node Identity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to execute this plan task-by-task in this task worktree. This plan intentionally stops before implementation until the user confirms execution. Production code changes require TDD: write each failing test, run it red, then implement the smallest code to pass.

**Goal:** Make runtime-agent node identity explicit and verifiable so CTF/AWD runtime operations route to the intended Docker node and operational failures can identify the affected node.

**Architecture:** Keep `runtime_nodes.id` as the API/database internal primary key and use `runtime_nodes.name` as the stable cross-process node identity. Add API-side `runtime_agent.node_name` for default node bootstrap and agent-side `runtime_agent.server.node_name` for health self-reporting, then verify the agent-reported name when constructing a remote runtime node client. Do not move topology/AWD placement ownership: existing `node_id` routing remains the execution owner for instances, topology, checker, AWD service and file operations.

**Tech Stack:** Go, Viper config/env binding, gRPC JSON codec runtime-agent contracts, GORM-backed `runtime_nodes`, existing container runtime composition and package-level Go tests.

---

## Task Metadata

- Task Slug: `2026-06-21-runtime-agent-node-identity`
- Started At: `2026-06-21T07:20:52Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-21-runtime-agent-node-identity`
- Branch: `task/2026-06-21-runtime-agent-node-identity`
- Plan Type: `slice`

## Plan Status

- Status: `implemented-pending-re-review`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled
- Review fix status:
  - [x] Draft review blocker fixed: runtime-agent identity Health check now has an explicit timeout.
  - [ ] Independent re-review after fix.

## Objective And Non-Goals

- Objective:
  - Add a stable, human-readable runtime node identity that is configured on both API and runtime-agent.
  - Make API default remote node bootstrap use the configured node name instead of always using `agent-default`.
  - Make runtime-agent health responses expose the configured node identity and hostname for troubleshooting.
  - Reject a remote runtime client if an explicitly configured API node name does not match the agent-reported node name.
  - Document how operators map an API `runtime_nodes.name` entry to a specific runtime-agent host.
- Non-Goals:
  - Do not expose or configure database `runtime_nodes.id` in runtime-agent.
  - Do not change the scheduler, capacity scoring, failover strategy, or `node_id` persistence model.
  - Do not implement live migration or topology-internal cross-node placement.
  - Do not add schema migrations; `runtime_nodes.name` already exists and is unique.
  - Do not make CTF/AWD topology placement more granular than one runtime node per topology/service execution unit.

## Problem Statement

- Current behavior / structure:
  - `code/backend/internal/app/composition/container_runtime_module.go` returns `agent-default` whenever `runtime_agent.enabled` is true.
  - `runtime_nodes.name` is already the unique stable node label, while `runtime_nodes.id` is the DB primary key used by `instances.node_id`, AWD service routing, checker metadata and runtime router caches.
  - `code/backend/internal/module/container_runtime/agentcontracts.HealthResponse` only returns `Ready` and `Capabilities`.
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver.Service` has no configured identity and cannot report which runtime-agent process handled a request.
  - `buildRuntimeNodeClientFromNode()` dials the node endpoint and caches the bridge without checking whether the remote agent is the node the API intended to use.
- Target behavior / structure:
  - API config may set `runtime_agent.node_name: <stable-runtime-node-name>`.
  - runtime-agent config may set `runtime_agent.server.node_name: <same-stable-runtime-node-name>`.
  - Default bootstrap uses `runtime_agent.node_name` when configured, otherwise preserves the existing `agent-default` / `local-default` fallback.
  - Agent `Health()` includes `node_name` and `hostname`.
  - Remote client construction checks `Health().node_name` against `runtime_nodes.name` when the API is explicitly using a configured node identity, and fails before caching on mismatch.
- Why this task is needed now:
  - Operators need to distinguish which Docker node failed when a runtime-agent or node-local Docker engine has a problem.
  - Fixed `agent-default` makes single remote deployments work but is unsafe as a default for true multi-node mapping.
  - CTF topologies and AWD runtime operations already route by `node_id`; they need a stable way to map that DB binding back to an actual runtime-agent host.

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
  - `docs/plan/README.md`
  - `code/backend/tests/README.md`
  - `harness/policies/reuse-first.yaml`
- Related architecture/contracts:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Related code:
  - `code/backend/internal/config/types.go`
  - `code/backend/internal/config/defaults.go`
  - `code/backend/internal/config/validate.go`
  - `code/backend/internal/config/load.go`
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/module/container_runtime/agentcontracts/messages.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go`
  - `code/backend/internal/module/container_runtime/application/node_health_service.go`
- Related prior work:
  - Existing runtime-node health/schedulable semantics: `runtime_nodes.last_seen_at`, `health_status`, `capacity_snapshot`.
  - Existing routing guardrails in `code/backend/internal/app/composition/runtime_node_execution_router_test.go`.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Touches runtime operations, config contract, gRPC agent protocol, runtime node bootstrap and docs.
  - A wrong implementation can route instance/checker/AWD operations to the wrong Docker host.
  - Requires TDD and code-workflow startup/review gates.

## Files

- Create:
  - None expected.
- Modify:
  - `code/backend/internal/config/types.go`: add `RuntimeAgentConfig.NodeName` and `RuntimeAgentServerConfig.NodeName`.
  - `code/backend/internal/config/defaults.go`: add empty defaults for both node-name fields.
  - `code/backend/internal/config/validate.go`: validate configured node names when present; runtime-agent server should require `node_name` only if strict matching is adopted in this slice.
  - `code/backend/configs/config.yaml`: document both config keys with empty/default examples.
  - `code/backend/deploy/systemd/ctf-runtime-agent.env.example`: add `CTF_RUNTIME_AGENT_SERVER_NODE_NAME`.
  - `code/backend/internal/module/container_runtime/agentcontracts/messages.go`: add `NodeName` and `Hostname` to `HealthResponse`.
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`: store server identity and report it from `Health()`.
  - `code/backend/internal/bootstrap/runtime_agent.go`: pass server config/identity into `agentserver.NewService`.
  - `code/backend/internal/app/composition/container_runtime_module.go`: make default node name use `runtime_agent.node_name` when present.
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`: validate remote agent identity during client construction.
  - `code/backend/internal/app/composition/container_runtime_module.go`: optionally extend `runtimeNodeStatsProbe` with health identity evidence if minimal router validation is not enough.
  - `docs/architecture/backend/03-container-architecture.md`: after implementation, update current runtime-node identity fact source.
  - `docs/operations/runtime-agent-deployment.md`: after implementation, update operator deployment and troubleshooting instructions.
- Review:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/contracts/runtime_node.go`
  - `docs/architecture/backend/01-system-architecture.md`
- Test:
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/app/composition/runtime_module_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service_test.go`
  - Optional if health probing changes: `code/backend/internal/module/container_runtime/application/node_health_service_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Config ownership: `RuntimeAgentConfig`, `RuntimeAgentServerConfig`, `setDefaults()`, `LoadRuntimeAgent()`, `Validate()` and `ValidateRuntimeAgent()`.
  - Runtime bootstrap ownership: `defaultRuntimeNodeName()`, `defaultRuntimeNodeEndpoint()`, `buildDefaultRuntimeNodeSelector()`.
  - Runtime execution ownership: `runtimeNodeExecutionRouter`, `buildRuntimeNodeClientFromNode()`, `runtimeAgentConfigForNode()`.
  - Agent protocol ownership: `agentcontracts.HealthResponse`, `agentserver.Service.Health()`, `agentclient.Bridge.Health()`.
- Reuse / extend / split / create-new decision:
  - `extend_existing`: add fields to existing config structs and health response rather than creating a new identity registry.
  - `extend_existing`: verify identity in `buildRuntimeNodeClientFromNode()` because that is the single owner that turns a `RuntimeNode` row into an executable client.
  - `reuse_existing`: use `runtime_nodes.name` as the stable identity and keep `runtime_nodes.id` as internal DB routing key.
- Owner boundary:
  - API config owns "which default runtime node name should be bootstrapped/selected".
  - `runtime_nodes` owns persistent node registry and uniqueness.
  - runtime-agent server config owns "what identity this agent claims".
  - router/client construction owns "remote endpoint identity matches expected node before execution".
  - health service owns liveness/capacity, not scheduling identity assignment.
- Why this is the narrowest safe surface:
  - It prevents wrong-node remote execution without changing instance schema, scheduler, runtime details, or topology provisioning behavior.
  - It improves observability without making agents depend on API database IDs.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming` for design branch selection.
  - `grill-with-docs` for terminology and architecture/doc fit.
  - `writing-plans` for this formal implementation plan.
- Why this pass fits:
  - The request started as an architecture/operations question and affects config, runtime routing, and docs.
  - The key decision is identity ownership, not a local code tweak.
- grill-with-docs findings:
  - "nodeId" must be disambiguated: API DB `runtime_nodes.id` is not the stable deploy-time identity.
  - Stable deploy-time identity should be named `node_name`, matching `runtime_nodes.name`.
  - CTF topology and AWD runtime placement already bind to one runtime node via existing `node_id` fields; no container-level node ID should be introduced.
  - Health self-report is useful for troubleshooting, but wrong-node prevention must happen before remote client execution/caching.
- Plan adjustments after challenge:
  - Do not add `node_id` to runtime-agent.
  - Add `node_name` to both API and agent config.
  - Make mismatch fail closed only when API has an explicit configured node name; preserve old single-agent defaults for local/non-explicit deployments.
  - Keep health probe changes optional/minimal so routing correctness is not coupled to capacity polling.

## Execution Slices

### Slice 1: Config Contract And Default Node Bootstrap

- Goal:
  - Load and validate API-side `runtime_agent.node_name` and agent-side `runtime_agent.server.node_name`.
  - Use API `runtime_agent.node_name` for default remote runtime node bootstrap.
- Dependencies:
  - None beyond existing config package.
- Files:
  - Modify: `code/backend/internal/config/types.go`
  - Modify: `code/backend/internal/config/defaults.go`
  - Modify: `code/backend/internal/config/validate.go`
  - Modify: `code/backend/internal/app/composition/container_runtime_module.go`
  - Test: `code/backend/internal/config/config_test.go`
  - Test: `code/backend/internal/app/composition/runtime_module_test.go`
- Steps:
  - [x] Write failing config tests proving `CTF_RUNTIME_AGENT_NODE_NAME` and `CTF_RUNTIME_AGENT_SERVER_NODE_NAME` load into the expected structs.
  - [x] Run: `cd code/backend && go test ./internal/config -run 'TestLoadReadsRuntimeAgent.*NodeName|TestLoadRuntimeAgent.*NodeName' -count=1`
  - [x] Confirm RED: tests fail because fields do not exist or values are empty.
  - [x] Write failing runtime module test proving `BuildContainerRuntimeModule()` selects/bootstraps the configured node name instead of `agent-default`.
  - [x] Run: `cd code/backend && go test ./internal/app/composition -run TestBuildContainerRuntimeModuleSelectsConfiguredRuntimeAgentNodeName -count=1`
  - [x] Confirm RED: test gets `agent-default`.
  - [x] Implement minimal config fields/defaults/validation and `defaultRuntimeNodeName()` change.
  - [x] Re-run the two focused tests and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/config ./internal/app/composition -run 'TestLoadReadsRuntimeAgent.*NodeName|TestLoadRuntimeAgent.*NodeName|TestBuildContainerRuntimeModuleSelectsConfiguredRuntimeAgentNodeName' -count=1`
- Review focus:
  - `node_name` normalization is trim-only and does not duplicate DB uniqueness.
  - No DB `runtime_nodes.id` is introduced into config.
- Done criteria:
  - Explicit API node name controls the default remote node binding.
  - Existing fallback remains `agent-default` for `runtime_agent.enabled=true` without `node_name`, and `local-default` otherwise.

### Slice 2: Runtime-Agent Health Identity

- Goal:
  - Make runtime-agent report configured `node_name` and host `hostname` in `Health()`.
- Dependencies:
  - Slice 1 config fields.
- Files:
  - Modify: `code/backend/internal/module/container_runtime/agentcontracts/messages.go`
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`
  - Modify: `code/backend/internal/bootstrap/runtime_agent.go`
  - Test: `code/backend/internal/module/container_runtime/infrastructure/agentserver/service_test.go`
- Steps:
  - [x] Write failing agentserver test proving `NewService(..., identity)` returns `HealthResponse.NodeName` and non-empty/stubbed `Hostname`.
  - [x] Run: `cd code/backend && go test ./internal/module/container_runtime/infrastructure/agentserver -run TestServiceHealthReportsNodeIdentity -count=1`
  - [x] Confirm RED: health response has no identity fields.
  - [x] Implement minimal identity storage in `agentserver.Service` and extend `HealthResponse`.
  - [x] Update `runtime_agent.go` to pass `cfg.RuntimeAgent.Server` or a narrow identity struct into `NewService`.
  - [x] Re-run focused test and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/module/container_runtime/infrastructure/agentserver -count=1`
- Review focus:
  - Health remains backward-compatible for existing callers because added JSON fields are optional additive fields.
  - Hostname is operational metadata only; routing still uses `node_name`.
- Done criteria:
  - Runtime-agent health identifies the configured logical node and physical host.

### Slice 3: Remote Client Identity Verification

- Goal:
  - Prevent API from caching/using a remote runtime client when the agent self-reports a different node name than the `runtime_nodes.name` row being dialed.
- Dependencies:
  - Slice 2 health identity.
- Files:
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - Test: `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
- Steps:
  - [x] Write failing router/client builder test with a stub `dialRuntimeAgent` returning `Health().NodeName="node-b"` while the selected DB node is `node-a`.
  - [x] Run: `cd code/backend && go test ./internal/app/composition -run TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch -count=1`
  - [x] Confirm RED: client is accepted/cached despite mismatch.
  - [x] Implement a small `verifyRuntimeAgentNodeIdentity(ctx, expectedName, bridge, strict)` helper in composition.
  - [x] Make strict mode true when API `runtime_agent.node_name` is non-empty, or when the selected node name equals that explicit config; preserve compatibility when agent health lacks `node_name` and API did not explicitly configure one.
  - [x] Close the bridge on identity mismatch before returning the error.
  - [x] Re-run focused test and confirm GREEN.
- Validation:
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch|TestRuntimeNodeExecutionRouter' -count=1`
- Review focus:
  - Mismatch errors include expected node name, reported node name, endpoint and hostname if available; no secrets.
  - The router does not fallback to another node for explicit old-container operations.
  - Cached client map remains untouched on mismatch.
- Done criteria:
  - Wrong runtime-agent endpoint cannot be used for an explicitly named remote node.

### Slice 4: Health Probe Evidence And Documentation

- Goal:
  - Update operator-facing docs and examples; optionally surface identity in health probing logs if needed.
- Dependencies:
  - Slices 1-3.
- Files:
  - Modify: `code/backend/configs/config.yaml`
  - Modify: `code/backend/deploy/systemd/ctf-runtime-agent.env.example`
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/operations/runtime-agent-deployment.md`
  - Optional modify: `code/backend/internal/app/composition/container_runtime_module.go`
  - Optional test: `code/backend/internal/module/container_runtime/application/node_health_service_test.go`
- Steps:
  - [x] Add commented/default config examples for API and agent node names.
  - [x] Add `CTF_RUNTIME_AGENT_SERVER_NODE_NAME=<stable-node-name>` to the systemd env example.
  - [x] Update operations doc to state that operators set the same stable `node_name` on API node registration and runtime-agent server config.
  - [x] Update architecture doc to state `runtime_nodes.name` is the deploy-time stable node identity and `runtime_nodes.id` remains internal routing key.
  - [x] If health probe is extended, write a failing test first; otherwise keep health probing unchanged and rely on router verification plus health response observability.
  - [x] Run docs whitespace check for touched docs.
- Validation:
  - `git diff --check -- code/backend/configs/config.yaml code/backend/deploy/systemd/ctf-runtime-agent.env.example docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md`
  - If architecture doc changes: `bash scripts/check-architecture.sh --full`
- Review focus:
  - Docs do not imply operators configure numeric DB `node_id`.
  - Docs preserve the one-topology/one-runtime-node invariant for CTF and AWD.
- Done criteria:
  - Operator docs can answer "which Docker node is broken?" through `runtime_nodes.name`, agent `node_name`, and health `hostname`.

## Impact And Compatibility

- API / DTO:
  - No HTTP API contract change.
  - Runtime-agent internal gRPC JSON health response gets additive `node_name` and `hostname` fields.
- Data / migration:
  - No schema migration.
  - Reuses existing `runtime_nodes.name` unique constraint.
- State / cache / queue / event:
  - No Redis/cache/event changes.
  - Runtime router cache refuses to store mismatched remote client.
- Runtime / config:
  - Adds API config `runtime_agent.node_name`.
  - Adds runtime-agent server config `runtime_agent.server.node_name`.
  - Existing deployments without explicit node name keep current `agent-default` behavior.
- Frontend route / state / UX:
  - No frontend change.
- Docs / contracts:
  - Update runtime-agent deployment and backend container architecture docs after code lands.

## Plan Review / Architecture Fit

- Target owner boundary:
  - Config owns deploy-time names; `runtime_nodes` owns persisted registry; router owns remote client identity verification.
- Reuse points / landing zones:
  - `defaultRuntimeNodeName()` is the default selection landing zone.
  - `HealthResponse` is the observability landing zone.
  - `buildRuntimeNodeClientFromNode()` is the wrong-node prevention landing zone.
- Known structural debt touched:
  - Fixed `agent-default` was a single-agent convenience in a code path now used for multi-node architecture.
  - Health probe currently conflates "agent reachable enough to list stats" with "this is the intended agent"; this plan avoids that by checking identity in client construction.
- How this plan avoids behavior-only convergence:
  - It changes the identity owner and client construction guard, not only docs or logs.
- Hidden second-redesign risk:
  - If later multi-node registration becomes API-driven CRUD, `runtime_agent.node_name` will only bootstrap the default node and explicit DB rows will still need operator/admin registration. This is acceptable because this task does not introduce node management UI.
- Decision after review:
  - Proceed with `runtime_nodes.name` / `runtime_agent.node_name` as the stable identity.
  - Do not configure or report DB `node_id` from runtime-agent.
  - Keep health probe identity logging optional for this slice; wrong-node execution must be blocked in router/client creation.

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Fact sources to update after implementation:
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Plan-only notes that must not become architecture source:
  - TDD execution order, temporary compatibility stance, and test command choices.
- Archive condition:
  - After code, docs, validation and independent review gate pass, archive this plan through `harness/workflow-plugins/code-workflow/archive_task_artifacts.sh`.

## Validation

- Required validation owner:
  - Focused RED/GREEN tests for config loading, default node bootstrap, runtime-agent Health identity and remote client mismatch / timeout behavior.
  - Affected package tests for `internal/config`, `internal/module/container_runtime/infrastructure/agentserver` and `internal/app/composition`.
  - Project workflow checks for startup gate, architecture guard and completion-full before merge.
- Current evidence:
  - Detailed command history and results are recorded in `## Validation Evidence`.
  - The draft-review blocker around identity Health deadline has a RED test and a GREEN re-run after the fix.
- Merge condition:
  - No known code blocker remains after the identity Health timeout fix.
  - The independent re-review still needs to be run or explicitly treated as part of the merge handoff.

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/config -run 'TestLoadReadsRuntimeAgent.*NodeName|TestLoadRuntimeAgent.*NodeName' -count=1`
  - `cd code/backend && go test ./internal/module/container_runtime/infrastructure/agentserver -run TestServiceHealthReportsNodeIdentity -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModuleSelectsConfiguredRuntimeAgentNodeName|TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch' -count=1`
- Integration commands:
  - `cd code/backend && go test ./internal/config ./internal/module/container_runtime/infrastructure/agentserver ./internal/app/composition -count=1`
- Workflow commands:
  - `bash scripts/check-startup-gate.sh`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Documentation commands:
  - `git diff --check -- <touched-files>`
  - `bash scripts/check-architecture.sh --full` if `docs/architecture/backend/03-container-architecture.md` changes.
- Manual checks:
  - Inspect health JSON/struct output to confirm `node_name` and `hostname` are visible.
  - Confirm mismatch error names expected/reported node and endpoint without leaking TLS paths or secrets.
- Commands intentionally skipped and why:
  - Full runtime/container integration tests are not required for the plan-only gate; revisit after implementation if health or real agent dialing behavior changes beyond package-level tests.

## Validation Evidence

- Command: `cd code/backend && go test ./internal/config -run 'TestLoadReadsRuntimeAgent.*NodeName|TestLoadRuntimeAgent.*NodeName' -count=1`
  - Result: RED first, then GREEN after implementation.
  - Notes: RED failed because `RuntimeAgentConfig.NodeName` and `RuntimeAgentServerConfig.NodeName` did not exist; GREEN passed after config fields/defaults/load normalization landed.
- Command: `cd code/backend && go test ./internal/app/composition -run TestBuildContainerRuntimeModuleSelectsConfiguredRuntimeAgentNodeName -count=1`
  - Result: RED first, then GREEN after implementation.
  - Notes: RED initially failed at compile-time before the config field existed; GREEN passed after default node bootstrap used `runtime_agent.node_name`.
- Command: `cd code/backend && go test ./internal/module/container_runtime/infrastructure/agentserver -run TestServiceHealthReportsNodeIdentity -count=1`
  - Result: RED first, then GREEN after implementation.
  - Notes: RED failed because `HealthResponse` lacked identity fields and `NewService` did not accept identity.
- Command: `cd code/backend && go test ./internal/module/container_runtime/infrastructure/agentserver -count=1`
  - Result: PASS.
  - Notes: Agent server package tests passed after Health identity support.
- Command: `cd code/backend && go test ./internal/app/composition -run TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch -count=1`
  - Result: RED first, then GREEN after implementation.
  - Notes: RED accepted/cached the mismatched remote client; GREEN rejected the mismatch before returning a client.
- Command: `cd code/backend && go test ./internal/app/composition -run TestBuildRuntimeNodeClientTimesOutRuntimeAgentIdentityCheck -count=1 -timeout=2s`
  - Result: RED.
  - Notes: Draft review blocker reproduced; identity Health reused a root context without deadline and did not return within the test timeout window.
- Command: `cd code/backend && go test ./internal/app/composition -run 'TestBuildRuntimeNodeClientTimesOutRuntimeAgentIdentityCheck|TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch' -count=1 -timeout=5s`
  - Result: PASS.
  - Notes: Identity Health now derives a timeout from `runtime_agent.dial_timeout`; mismatch behavior still passes.
- Command: `cd code/backend && go test ./internal/app/composition -run 'TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch|TestRuntimeNodeExecutionRouter' -count=1`
  - Result: PASS.
  - Notes: Router mismatch check and existing router behavior tests passed.
- Command: `git diff --check -- code/backend/configs/config.yaml code/backend/deploy/systemd/ctf-runtime-agent.env.example docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md`
  - Result: PASS.
  - Notes: Touched config/docs files have no whitespace errors.
- Command: `bash scripts/check-architecture.sh --full`
  - Result: PASS.
  - Notes: Backend architecture tests and frontend architecture guardrails passed.
- Command: `cd code/backend && go test ./internal/config ./internal/module/container_runtime/infrastructure/agentserver ./internal/app/composition -count=1`
  - Result: PASS.
  - Notes: Affected backend packages passed.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS.
  - Notes: Startup gate accepted the current task context.
- Command: `/home/azhi/.codex/skills/development-pipeline/scripts/check_impl_plan_done.sh docs/plan/archive/impl-plan/2026-06/2026-06-21-runtime-agent-node-identity-implementation-plan.md`
  - Result: PASS.
  - Notes: All 30 implementation-plan checklist items are marked complete.
- Command: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS.
  - Notes: Project completion-full stage passed.

## Independent Review Handoff

- Review target:
  - Branch `task/2026-06-21-runtime-agent-node-identity`
  - Plan `docs/plan/archive/impl-plan/2026-06/2026-06-21-runtime-agent-node-identity-implementation-plan.md`
  - Implementation diff after code lands.
- Validation evidence summary:
  - Include RED/GREEN focused test output and final package test output from `code/backend`.
- Architecture / contract inputs:
  - `AGENTS.md`
  - `code/backend/tests/README.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Known risks / review focus:
  - API DB `node_id` must not leak into runtime-agent config.
  - Existing non-explicit `agent-default` behavior should remain usable.
  - Wrong-node mismatch must fail before client caching.
  - CTF topology/AWD one-runtime-node placement invariant must remain unchanged.
- Project-local checks to consider:
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash scripts/check-workflow-governance.sh`

## Rollback / Recovery

- Safe revert boundary:
  - Revert this task's config/protocol/router/docs changes as one slice.
- Data / config / runtime recovery notes:
  - No data migration rollback required.
  - If an operator misconfigures names, fix API `runtime_agent.node_name`, runtime-agent `runtime_agent.server.node_name`, or the corresponding `runtime_nodes.name` row, then restart affected processes.
- Irreversible operations:
  - None.

## Residual Risks

- Risk:
  - Existing runtime-agent binaries do not report `node_name`.
- Why acceptable:
  - Plan keeps compatibility for deployments without explicit API `runtime_agent.node_name`; strict matching applies when an operator opts into explicit identity.
- Follow-up owner, if any:
  - Future node-management UI/API can manage `runtime_nodes.name` rows, but this is outside this slice.
