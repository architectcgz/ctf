<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# Runtime Agent Only Execution Plane Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make runtime execution agent-only by default: API and AWD defense SSH gateway must route container execution through runtime-agent unless the process is in test mode or an explicit non-production local fallback is enabled.

**Architecture:** Keep `container_runtime` as the execution-plane owner. Add one explicit config switch for local fallback, enforce it at composition/router boundaries before any local Docker executor is built, and make local dev scripts start a local `runtime-agent` instead of giving API direct Docker authority.

**Tech Stack:** Go backend, Viper config, GORM-backed runtime node registry, runtime-agent mTLS client/server, Bash dev scripts, Docker Compose dev topology, Markdown architecture and operations docs.

---

## Task Metadata

- Task Slug: `2026-06-20-runtime-agent-only-execution-plane`
- Started At: `2026-06-20T06:08:17Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-20-runtime-agent-only-execution-plane`
- Branch: `task/2026-06-20-runtime-agent-only-execution-plane`
- Plan Type: `slice`

## Plan Status

- Status: `review-pending`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: close the remaining default runtime execution paths where API/gateway can directly construct Docker-backed local executors, while preserving unit-test runtime behavior and an explicit non-production emergency fallback.
- Non-Goals:
  - Do not migrate challenge image build/push ownership; `challenge/infrastructure/docker_image_builder.go` remains a separate image-build plane.
  - Do not build production certificate automation or a full PKI workflow.
  - Do not add runtime capacity scheduling, live migration, or transparent session migration.
  - Do not remove Docker access from the `runtime-agent` process; agent nodes still own Docker-side effects.

## Problem Statement

- Current behavior / structure:
  - `code/backend/internal/config/defaults.go` defaults `runtime_agent.enabled=false`.
  - `container_runtime_module.go` bootstraps `local-default` with `local://docker` when runtime-agent is disabled.
  - `runtime_node_execution_router.go` builds `NewLocalHostExecutor` and `NewDockerSandboxExecutor` for `local://docker` in non-test env.
  - `code/backend/scripts/dev-run.sh` starts API but not a local runtime-agent.
  - `docker/docker-compose.dev.yml` mounts `/var/run/docker.sock` into `ctf-api` and `ctf-awd-defense-ssh-gateway`.
- Target behavior / structure:
  - Non-test runtime execution requires runtime-agent by default.
  - Local Docker executor is allowed only in `APP_ENV=test`, or when `runtime_agent.allow_local_fallback=true` in non-production.
  - Production rejects `runtime_agent.allow_local_fallback=true`.
  - Local development startup provisions and starts a local runtime-agent and injects runtime-agent client env into API.
  - Dev compose moves Docker socket authority from API/gateway to a dedicated runtime-agent service.
- Why this task is needed now:
  - The architecture docs already name runtime-agent as the target execution boundary, but defaults and dev topology still train contributors to run API with Docker authority.
  - The security todo `docs/todos/2026-06-02-security-review-findings.md` tracks direct docker.sock mounting as a high-risk finding.

## Inputs

- Source docs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/todos/2026-06-02-security-review-findings.md`
  - `README.md`
- Related architecture/contracts:
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver`
- Related prior work:
  - Runtime node health, `runtime_nodes` bootstrap, mTLS runtime-agent client/server, and dual-node e2e tooling already exist.

## Task Classification

- Classification: `非琐碎任务`
- Why: touches runtime architecture boundary, config validation, composition behavior, dev scripts, Docker Compose security posture, tests, and architecture/operations docs.

## Files

- Create:
  - none expected.
- Modify:
  - `code/backend/internal/config/types.go`
  - `code/backend/internal/config/defaults.go`
  - `code/backend/internal/config/validate.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/internal/app/composition/container_runtime_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_module_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/Dockerfile`
  - `code/backend/scripts/dev-run.sh`
  - `docker/docker-compose.dev.yml`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/todos/2026-06-02-security-review-findings.md`
  - `README.md`
- Review:
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `tools/runtime-agent-dual-node-e2e.sh`
- Test:
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/app/composition/runtime_module_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Runtime-agent client/server config uses `RuntimeAgentConfig` and `runtimeAgentConfigForNode`.
  - Runtime node selection already persists default `local-default` or `agent-default` through `runtime_nodes`.
  - Dev TLS generation patterns exist in `tools/runtime-agent-dual-node-e2e.sh`.
- Reuse / extend / split / create-new decision:
  - Extend `RuntimeAgentConfig` with `allow_local_fallback`.
  - Enforce local executor eligibility in composition/router, not in generic `Config.Validate`, because server-only `LoadRuntimeAgent` and unit tests must remain small.
  - Reuse the existing local Docker executor only behind the explicit fallback guard.
- Owner boundary:
  - `config` owns parsing and production-safety validation.
  - `app/composition` owns deciding which execution client may be built.
  - `scripts/dev-run.sh` and compose own local development wiring.
  - `runtime-agent` remains the only runtime process that may talk to Docker in default dev/prod runtime execution.
- Why this is the narrowest safe surface:
  - The existing agent protocol and runtime node router are already present; changing defaults and guardrails closes the gap without redesigning provisioning, checker, ACL, or node health flows.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: this is a behavior and architecture-direction change, not a bug reproduction.
- grill-with-docs findings:
  - Docs and code disagree: docs call direct API Docker fallback "development fallback", while target is agent-only local dev.
  - Security todo already identifies docker.sock in API compose as high risk.
  - Image build Docker access is a separate image-build plane and should not block runtime execution-plane convergence.
- Plan adjustments after challenge:
  - Added explicit fallback config instead of relying on `runtime_agent.enabled=false`.
  - Production fallback rejection is config-owned.
  - Non-test direct local executor construction is composition-owned and tested.
  - Dev script/compose changes are in scope so default workflows exercise runtime-agent.

## Execution Slices

### Slice 1: Config Contract For Explicit Local Fallback

- Goal: expose `runtime_agent.allow_local_fallback` with safe defaults and production rejection.
- Dependencies: none.
- Files:
  - Modify: `code/backend/internal/config/types.go`
  - Modify: `code/backend/internal/config/defaults.go`
  - Modify: `code/backend/internal/config/validate.go`
  - Modify: `code/backend/internal/config/config_test.go`
  - Modify: `code/backend/configs/config.yaml`
- Steps:
  - [x] Step 1: Write failing config tests for production rejecting `runtime_agent.allow_local_fallback=true` and default false.
  - [x] Step 2: Run `go test ./internal/config -run 'TestValidateRejectsRuntimeAgentLocalFallbackInProduction|TestRuntimeAgentLocalFallbackDefaultsFalse' -count=1` from `code/backend`; expected RED.
  - [x] Step 3: Add `AllowLocalFallback` field, default, YAML key, and prod validation.
  - [x] Step 4: Re-run the same config test command; expected PASS.
- Validation:
  - `go test ./internal/config -run 'TestValidateRejectsRuntimeAgentLocalFallbackInProduction|TestRuntimeAgentLocalFallbackDefaultsFalse' -count=1`
- Review focus:
  - No generic validation that breaks `LoadRuntimeAgent`.
  - Error message names `runtime_agent.allow_local_fallback`.
- Done criteria:
  - Config tests prove fallback is opt-in and forbidden in production.

### Slice 2: Composition Guardrail Against Direct Docker Runtime

- Goal: prevent non-test API/gateway composition from building local Docker executors unless explicit non-prod fallback is enabled.
- Dependencies: Slice 1.
- Files:
  - Modify: `code/backend/internal/app/composition/container_runtime_module.go`
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - Modify: `code/backend/internal/app/composition/runtime_module_test.go`
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
- Steps:
  - [x] Step 1: Write failing composition tests:
    - dev env + runtime-agent disabled + no fallback fails before local executor construction.
    - dev env + `allow_local_fallback=true` can build the local executor path.
    - test env still uses test runtime engine.
  - [x] Step 2: Run `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModuleRejectsLocalRuntimeWithoutExplicitFallback|TestBuildContainerRuntimeModuleAllowsExplicitLocalRuntimeFallback|TestRuntimeNodeClientAllowsTestRuntimeEngineWithoutFallback' -count=1`; expected RED.
  - [x] Step 3: Implement small helper(s) for local runtime eligibility and actionable errors.
  - [x] Step 4: Re-run the same composition test command; expected PASS.
- Validation:
  - `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModuleRejectsLocalRuntimeWithoutExplicitFallback|TestBuildContainerRuntimeModuleAllowsExplicitLocalRuntimeFallback|TestRuntimeNodeClientAllowsTestRuntimeEngineWithoutFallback' -count=1`
- Review focus:
  - Guard runs before `newLocalRuntimeHostRunner` and `newLocalSandboxExecutor`.
  - Existing runtime-agent path still dials by node endpoint/TLS identity.
  - Existing test runtime path is not coupled to Docker.
- Done criteria:
  - Non-test direct local Docker runtime is impossible by default.

### Slice 3: Local Dev Runtime-Agent Startup

- Goal: make `code/backend/scripts/dev-run.sh` start a local runtime-agent by default and inject client mTLS env into API.
- Dependencies: Slice 1 and Slice 2.
- Files:
  - Modify: `code/backend/scripts/dev-run.sh`
- Steps:
  - [x] Step 1: Add script behavior checks by static inspection plan: runtime-agent default, opt-out variable, cert generation, API env injection, background/foreground cleanup.
  - [x] Step 2: Implement bounded Bash helpers:
    - generate local CA/server/client certs under `docker/runtime/runtime-agent-certs` unless present.
    - choose runtime-agent port from `CTF_RUNTIME_AGENT_SERVER_PORT`, default `19443`.
    - start `go run ./cmd/runtime-agent` before API unless `CTF_DEV_RUNTIME_AGENT=false` or explicit `CTF_RUNTIME_AGENT_ALLOW_LOCAL_FALLBACK=true`.
    - export `CTF_RUNTIME_AGENT_ENABLED=true`, endpoint, server name, CA, client cert, and client key for API.
    - stop the foreground runtime-agent when foreground API exits.
  - [x] Step 3: Run `bash -n code/backend/scripts/dev-run.sh`; expected PASS.
- Validation:
  - `bash -n code/backend/scripts/dev-run.sh`
- Review focus:
  - No long-lived foreground process leak.
  - Background mode records enough log/PID context for operators.
  - `DOCKER_HOST` authority is scoped to runtime-agent process, not API process.
- Done criteria:
  - Default `dev-run.sh` API process uses runtime-agent client env.

### Slice 4: Dev Compose Execution Plane

- Goal: route compose dev API/gateway through `ctf-runtime-agent` and remove docker.sock from API/gateway.
- Dependencies: Slice 1 and Slice 2.
- Files:
  - Modify: `docker/docker-compose.dev.yml`
- Steps:
  - [x] Step 1: Add `ctf-runtime-agent` service using backend image, docker.sock mount, and mTLS env.
  - [x] Step 2: Add a one-shot cert generation service using `alpine` + `openssl` into shared runtime storage.
  - [x] Step 3: Set API and gateway runtime-agent client env and remove their `DOCKER_HOST` and `/var/run/docker.sock` mount.
  - [x] Step 4: Run `docker compose -f docker/docker-compose.dev.yml config >/tmp/ctf-compose-dev-config.yaml`; expected PASS.
- Validation:
  - `docker compose -f docker/docker-compose.dev.yml config >/tmp/ctf-compose-dev-config.yaml`
- Review focus:
  - API/gateway no longer receive Docker socket.
  - Only runtime-agent service owns Docker socket in compose dev.
  - Cert files are generated in runtime storage, not committed as static private keys.
- Done criteria:
  - Compose topology expresses API/gateway -> runtime-agent -> Docker.

### Slice 5: Docs And README Alignment

- Goal: make docs match the new default runtime boundary and note the explicit fallback.
- Dependencies: Slices 1-4.
- Files:
  - Modify: `docs/operations/runtime-agent-deployment.md`
  - Modify: `docs/architecture/backend/01-system-architecture.md`
  - Modify: `README.md`
- Steps:
  - [x] Step 1: Update runtime-agent deployment modes: local dev starts local runtime-agent; direct local Docker is explicit fallback only.
  - [x] Step 2: Update backend architecture sections 7.2/7.5/7.6 and runtime node binding notes.
  - [x] Step 3: Update README dev/security caveat so docker.sock is not described as API/gateway default.
  - [x] Step 4: Run `git diff --check -- docs/operations/runtime-agent-deployment.md docs/architecture/backend/01-system-architecture.md README.md`; expected PASS.
- Validation:
  - `git diff --check -- docs/operations/runtime-agent-deployment.md docs/architecture/backend/01-system-architecture.md README.md`
- Review focus:
  - Docs do not claim image build plane is solved.
  - Architecture docs remain current facts, not aspirational statements.
- Done criteria:
  - Stable docs and README describe the same runtime execution default as code/scripts.

### Slice 6: Integration Validation And Workflow Gate

- Goal: run the smallest sufficient post-change verification and record evidence.
- Dependencies: Slices 1-5.
- Files:
  - Modify: this implementation plan checklist and validation evidence.
- Steps:
  - [x] Step 1: Run targeted config tests.
  - [x] Step 2: Run targeted composition tests.
  - [x] Step 3: Run script syntax check.
  - [x] Step 4: Run compose config check.
  - [x] Step 5: Run `bash scripts/check-startup-gate.sh`.
  - [x] Step 6: Run project completion gate if feasible: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`.
  - [x] Step 7: Record validation evidence and independent review handoff details.
- Validation:
  - `go test ./internal/config -run 'TestValidate.*RuntimeAgent|TestRuntimeAgentLocalFallback' -count=1`
  - `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter|TestRuntimeNodeClient' -count=1`
  - `bash -n code/backend/scripts/dev-run.sh`
  - `docker compose -f docker/docker-compose.dev.yml config >/tmp/ctf-compose-dev-config.yaml`
  - `bash scripts/check-startup-gate.sh`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Review focus:
  - Any test failures caused by existing unrelated repo state must be separated from this task's changes.
  - Independent review gate remains required after local validation.
- Done criteria:
  - Targeted checks pass or blockers are explicitly recorded with command output.

## Impact And Compatibility

- API / DTO: no external HTTP API or DTO contract changes.
- Data / migration: no schema migration expected; default `runtime_nodes` bootstrap data changes by config path only.
- State / cache / queue / event: no queue/cache/event contract change.
- Runtime / config:
  - New config key: `runtime_agent.allow_local_fallback`.
  - Default remains false.
  - Production rejects true.
  - Non-test local Docker runtime requires explicit true.
- Frontend route / state / UX: none.
- Docs / contracts: architecture, operations, and README text updated.

## Plan Review / Architecture Fit

- Target owner boundary:
  - API/gateway are runtime execution clients.
  - runtime-agent is the runtime execution authority that touches Docker.
  - local test runtime remains package-local to composition tests.
- Reuse points / landing zones:
  - `RuntimeAgentConfig` for fallback flag.
  - `buildRuntimeNodeClientFromNode` for the runtime client creation guard.
  - `dev-run.sh` and compose for local execution topology.
- Known structural debt touched:
  - Security todo for API docker.sock mount.
  - Existing `runtime_agent.enabled=false` local default.
- How this plan avoids behavior-only convergence:
  - It changes default construction behavior, startup wiring, and docs together.
  - It rejects unsafe production config instead of only documenting it.
- Hidden second-redesign risk:
  - Image build plane still calls Docker directly and is intentionally out of scope; this is recorded as a non-goal so runtime execution work is not presented as full Docker authority removal.
- Decision after review:
  - Proceed with implementation. Same-context architecture-fit review found no immediate second redesign for the runtime execution plane. Tooling in this session does not provide a separate plan-document-reviewer subagent, so final independent implementation review remains a separate completion gate.

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `README.md`
- Fact sources to update after implementation:
  - Same three files above.
- Plan-only notes that must not become architecture source:
  - The exact Bash implementation steps for local cert generation.
  - Temporary validation outputs.
- Archive condition:
  - After implementation, validation, review, and absorption into architecture/operations docs, archive through `harness/workflow-plugins/code-workflow/archive_task_artifacts.sh`.

## Validation

- Per-slice commands:
  - See each slice.
- Integration commands:
  - `go test ./internal/config -run 'TestValidate.*RuntimeAgent|TestRuntimeAgentLocalFallback' -count=1`
  - `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter|TestRuntimeNodeClient' -count=1`
  - `bash -n code/backend/scripts/dev-run.sh`
  - `docker compose -f docker/docker-compose.dev.yml config >/tmp/ctf-compose-dev-config.yaml`
  - `bash scripts/check-startup-gate.sh`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - Inspect `docker compose config` output for no `/var/run/docker.sock` under `ctf-api` or `ctf-awd-defense-ssh-gateway`.
- Commands intentionally skipped and why:
  - Full runtime-agent dual-node e2e is not required for this default-boundary change unless targeted tests or compose validation expose protocol-level uncertainty.

## Validation Evidence

- Command: `go test ./internal/config -run 'TestValidateRejectsRuntimeAgentLocalFallbackInProduction|TestRuntimeAgentLocalFallbackDefaultsFalse' -count=1`
  - Result: PASS, `ok ctf-platform/internal/config 0.003s`
  - Notes: RED first failed because `RuntimeAgentConfig.AllowLocalFallback` did not exist.
- Command: `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModuleRejectsLocalRuntimeWithoutExplicitFallback|TestBuildContainerRuntimeModuleAllowsExplicitLocalRuntimeFallback|TestRuntimeNodeClientAllowsTestRuntimeEngineWithoutFallback' -count=1`
  - Result: PASS, `ok ctf-platform/internal/app/composition 0.168s`
  - Notes: RED first failed because dev config still built the local runtime without explicit fallback.
- Command: `go test ./internal/config -run 'TestValidate.*RuntimeAgent|TestRuntimeAgentLocalFallback' -count=1`
  - Result: PASS, `ok ctf-platform/internal/config 0.004s`
- Command: `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter|TestRuntimeNodeClient' -count=1`
  - Result: PASS, `ok ctf-platform/internal/app/composition 0.850s`
- Command: `bash -n code/backend/scripts/dev-run.sh`
  - Result: PASS
- Command: `docker compose -f docker/docker-compose.dev.yml config >/tmp/ctf-compose-dev-config.yaml`
  - Result: PASS
- Command: `rg -n "ctf-api:|ctf-awd-defense-ssh-gateway:|ctf-runtime-agent:|/var/run/docker.sock|DOCKER_HOST|CTF_RUNTIME_AGENT_ENDPOINT" /tmp/ctf-compose-dev-config.yaml`
  - Result: PASS by inspection; `DOCKER_HOST` and `/var/run/docker.sock` appear only under `ctf-runtime-agent`, while API/gateway contain `CTF_RUNTIME_AGENT_ENDPOINT`.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS, `PASS: no startup-gated changes in diff`
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS, workflow stage passed
- Command: `git diff --check`
  - Result: PASS
- Command: `timeout 180s bash scripts/check-workflow-governance.sh`
  - Result: PASS, workflow governance checks passed
  - Notes: after closing the Docker socket todo, open todo reminder dropped from 11 to 10 items.
- Command: `timeout 180s bash scripts/check-architecture.sh --full`
  - Result: PASS, architecture checks passed
- Post-review fix command: `bash -n code/backend/scripts/dev-run.sh`
  - Result: PASS
  - Notes: after fixing review M1/L2.
- Post-review fix command: `docker compose -f docker/docker-compose.dev.yml config >/tmp/ctf-compose-dev-config.yaml`
  - Result: PASS
  - Notes: after changing runtime-agent healthcheck to `openssl s_client` and adding `openssl` to the backend runtime image.
- Post-review fix command: `git diff --check -- code/backend/scripts/dev-run.sh code/backend/Dockerfile docker/docker-compose.dev.yml`
  - Result: PASS
- Post-review fix command: `go test ./internal/config -run 'TestValidate.*RuntimeAgent|TestRuntimeAgentLocalFallback' -count=1`
  - Result: PASS, `ok ctf-platform/internal/config 0.004s`
- Post-review fix command: `go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestRuntimeNodeExecutionRouter|TestRuntimeNodeClient' -count=1`
  - Result: PASS, `ok ctf-platform/internal/app/composition 0.787s`

## Independent Review Handoff

- Review target: current branch diff for `task/2026-06-20-runtime-agent-only-execution-plane`.
- Validation evidence summary: targeted config/composition tests, script syntax, compose config, startup gate, completion-full, workflow governance, architecture full, and `git diff --check` all passed.
- Architecture / contract inputs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/todos/2026-06-02-security-review-findings.md`
- Known risks / review focus:
  - local fallback guard bypasses
  - accidental API/gateway Docker authority in compose or dev-run
  - docs overstating challenge image-build ownership
  - production config safety
- Project-local checks to consider:
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash scripts/check-workflow-governance.sh`
- Gate status: independent review is still pending because this Codex tool session has no separate reviewer/subagent dispatch tool available.

## Rollback / Recovery

- Safe revert boundary:
  - Revert this branch to restore previous local Docker default.
- Data / config / runtime recovery notes:
  - If local runtime-agent startup fails during development, temporarily set `CTF_RUNTIME_AGENT_ALLOW_LOCAL_FALLBACK=true` and `CTF_DEV_RUNTIME_AGENT=false` in non-production only.
  - Production must not use fallback; fix runtime-agent connectivity or certificates instead.
- Irreversible operations:
  - None expected.

## Residual Risks

- Risk: challenge image build still has Docker authority outside runtime execution.
- Why acceptable: it belongs to image build plane, explicitly out of this task's scope.
- Follow-up owner, if any: challenge/image build architecture task if the user wants "API never touches Docker" extended to build pipelines.
