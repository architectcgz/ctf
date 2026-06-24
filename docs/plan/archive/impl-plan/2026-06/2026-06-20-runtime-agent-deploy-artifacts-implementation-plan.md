<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# runtime-agent-deploy-artifacts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or execute the slices inline while keeping checkboxes current. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide an actual runtime-agent deployment artifact path instead of leaving only `go run ./cmd/runtime-agent` in docs.

**Architecture:** Keep `runtime-agent` as the existing Go command and add packaging around it: a tracked backend build script for host binaries, Docker image inclusion for command override use, and systemd templates for the recommended host-process deployment. Do not change the runtime-agent protocol, config schema, or node routing behavior.

**Tech Stack:** Go, shell script, Dockerfile, systemd unit templates, existing `CTF_*` Viper environment mapping.

---

## Task Metadata

- Task Slug: `2026-06-20-runtime-agent-deploy-artifacts`
- Started At: `2026-06-20T02:29:31Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-20-runtime-agent-deploy-artifacts`
- Branch: `task/2026-06-20-runtime-agent-deploy-artifacts`
- Plan Type: `slice`

## Plan Status

- Status: `review-pending`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: Add minimal, reviewable packaging and deployment artifacts for `cmd/runtime-agent`.
- Non-Goals:
  - Do not redesign `runtime_agent` protocol, mTLS, node routing, or health probing.
  - Do not make runtime-agent a required dev compose service.
  - Do not claim containerized runtime-agent is the preferred production path.
  - Do not add certificate generation automation in this slice.

## Problem Statement

- Current behavior / structure: `cmd/runtime-agent` exists, and docs say to use `go run` or compile a binary manually, but tracked build and Docker artifacts omit runtime-agent.
- Target behavior / structure: operators can run `code/backend/scripts/build-runtime-agent.sh`, find a systemd unit/env template, and optionally use the backend image with `/app/ctf-runtime-agent`.
- Why this task is needed now: The documented deployment model depends on a deployable runtime-agent, but the repository does not currently package one.

## Inputs

- Source docs:
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `code/backend/cmd/runtime-agent/main.go`
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentclient/bridge.go`
  - `code/backend/internal/module/container_runtime/infrastructure/agentserver/service.go`
- Related prior work:
  - `tools/runtime-agent-dual-node-e2e.sh` already builds `./cmd/runtime-agent` temporarily for test use.

## Task Classification

- Classification: `非琐碎任务`
- Why: Touches build packaging, Docker image contents, deployment templates, and operations docs.

## Files

- Create:
  - `code/backend/scripts/build-runtime-agent.sh`
  - `code/backend/deploy/systemd/ctf-runtime-agent.service`
  - `code/backend/deploy/systemd/ctf-runtime-agent.env.example`
- Modify:
  - `code/backend/Dockerfile`
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - `code/backend/internal/config/load.go`
  - `code/backend/internal/config/validate.go`
  - `code/backend/internal/config/config_test.go`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-20-runtime-agent-deploy-artifacts-implementation-plan.md`
- Review:
  - `code/backend/internal/config/load.go`
  - `code/backend/internal/bootstrap/runtime_agent.go`
- Test:
  - Build target and Dockerfile smoke checks.

## 复用与 Owner 决策

- Existing patterns searched:
  - Current Dockerfile builds colocated backend commands.
  - Backend `scripts/` already owns developer/operator helper commands such as `dev-run.sh`.
  - Existing docs under `docs/operations/` own deployment runbooks.
- Reuse / extend / split / create-new decision:
  - Extend existing Dockerfile.
  - Add a tracked backend-local build script instead of depending on the untracked local Makefile found in the original worktree.
  - Create `code/backend/deploy/systemd/` as backend-command deployment templates colocated with backend command code.
- Owner boundary:
  - `code/backend` owns command artifacts.
  - `docs/operations/runtime-agent-deployment.md` owns deployment instructions.
- Why this is the narrowest safe surface:
  - No runtime behavior changes are required; the gap is packaging and operator instructions only.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: This adds a developer/operator-visible capability and needs deployment-shape tradeoff review.
- grill-with-docs findings:
  - Architecture already requires runtime-agent on the Docker host.
  - Docs explicitly prefer host-process deployment via systemd or equivalent.
  - Code uses `CTF_*` env mapping, so no config parser changes are needed for an EnvironmentFile template.
  - A Docker artifact can be useful, but must be documented as optional because containerizing runtime-agent requires high host privileges.
- Plan adjustments after challenge:
  - Include both host binary/systemd and Docker image inclusion, but keep systemd as the recommended path.

## Execution Slices

### Slice 1: Backend Build Artifacts

- Goal: Make `cmd/runtime-agent` available through local build targets and backend image contents.
- Dependencies: none.
- Files:
  - Create: `code/backend/scripts/build-runtime-agent.sh`
  - Modify: `code/backend/Dockerfile`
- Steps:
  - [x] Step 1: Run `./scripts/build-runtime-agent.sh` before editing and confirm it fails because the script is missing.
  - [x] Step 2: Add `code/backend/scripts/build-runtime-agent.sh` for a tracked host binary build path.
  - [x] Step 3: Add `/out/ctf-runtime-agent` build and `/app/ctf-runtime-agent` copy to Dockerfile; expose `9443`.
  - [x] Step 4: Run `./scripts/build-runtime-agent.sh`.
- Validation:
  - `cd code/backend && ./scripts/build-runtime-agent.sh`
- Review focus:
  - Build script defaults are explicit and overridable.
  - Runtime-agent binary name is stable.
- Done criteria:
  - `bin/ctf-runtime-agent` is produced.
  - Dockerfile contains `/app/ctf-runtime-agent`.

### Slice 2: Host Deployment Templates

- Goal: Provide operator-copyable systemd templates for the recommended deployment mode.
- Dependencies: Slice 1.
- Files:
  - Create: `code/backend/deploy/systemd/ctf-runtime-agent.service`
  - Create: `code/backend/deploy/systemd/ctf-runtime-agent.env.example`
- Steps:
  - [x] Step 1: Add a systemd unit that runs `/opt/ctf/backend/ctf-runtime-agent` with `APP_ENV=prod` and an EnvironmentFile.
  - [x] Step 2: Add an env example using existing `CTF_RUNTIME_AGENT_SERVER_*`, `CTF_CONTEST_AWD_CHECKER_SANDBOX_*`, and optional `DOCKER_HOST`.
  - [x] Step 3: Keep the template host-process oriented; do not require Docker socket mounts.
- Validation:
  - Shell/text smoke checks that referenced paths and env names are consistent.
- Review focus:
  - Unit does not run migrations or API.
  - Env names match existing Viper mapping.
- Done criteria:
  - Templates are present and referenced from operations docs.

### Slice 3: Operations Documentation

- Goal: Update runtime-agent deployment docs so the new artifacts are discoverable.
- Dependencies: Slices 1 and 2.
- Files:
  - Modify: `docs/operations/runtime-agent-deployment.md`
- Steps:
  - [x] Step 1: Replace the deploy-time `go run` emphasis with build script + binary install instructions.
  - [x] Step 2: Document systemd template locations and required env/config values.
  - [x] Step 3: Document optional Docker image command override and the security caveat.
  - [x] Step 4: Record validation evidence in this plan.
- Validation:
  - `python3 scripts/check-docs-consistency.py`
  - `bash scripts/check-workflow-governance.sh` if documentation or governance checks require it.
- Review focus:
  - Docs match actual artifact paths and keep systemd as preferred production path.
- Done criteria:
  - Operators can identify the binary target, systemd templates, runtime-agent config, and Docker caveat from one runbook.

### Slice 4: Runtime-Agent Config Boundary

- Goal: Make the deployed runtime-agent usable without copying API-only production secrets onto the Docker host.
- Dependencies: Slices 1 and 2.
- Files:
  - Modify: `code/backend/internal/bootstrap/runtime_agent.go`
  - Modify: `code/backend/internal/config/load.go`
  - Modify: `code/backend/internal/config/validate.go`
  - Modify: `code/backend/internal/config/config_test.go`
  - Modify: `code/backend/deploy/systemd/ctf-runtime-agent.env.example`
  - Modify: `docs/operations/runtime-agent-deployment.md`
- Steps:
  - [x] Step 1: Add a failing config test showing `LoadRuntimeAgent("prod")` can load runtime-agent server config without API PostgreSQL / Redis / CORS / flag secret production requirements.
  - [x] Step 2: Add `LoadRuntimeAgent` and `ValidateRuntimeAgent` so the runtime-agent process validates only its runtime boundary.
  - [x] Step 3: Switch `cmd/runtime-agent` bootstrap to the runtime-agent loader.
  - [x] Step 4: Document that Docker hosts do not need API-only secrets.
- Validation:
  - `cd code/backend && go test ./internal/config -run TestLoadRuntimeAgentDoesNotRequireAPIProductionSecrets -count=1`
  - `cd code/backend && go test ./internal/config -count=1`
- Review focus:
  - Full API `Load` behavior and error wrapping remain unchanged.
  - Runtime-agent validation still checks TLS server config, Docker runtime defaults, registry credentials, and checker sandbox limits.
- Done criteria:
  - `runtime-agent` deployment templates are not dependent on API-only production config.

## Impact And Compatibility

- API / DTO: none.
- Data / migration: none.
- State / cache / queue / event: none.
- Runtime / config: adds deployment templates that use existing env/config keys; no schema change.
- Runtime / config: runtime-agent bootstrap now uses a narrower config loader that avoids API-only production config requirements on Docker hosts.
- Frontend route / state / UX: none.
- Docs / contracts: updates operations runbook only.

## Plan Review / Architecture Fit

- Target owner boundary: backend command packaging belongs to `code/backend`; operational procedure belongs to `docs/operations`.
- Reuse points / landing zones: existing `cmd/runtime-agent`, backend scripts, Dockerfile, Viper `CTF_*` env mapping.
- Known structural debt touched: Docker socket risk remains tracked separately in `docs/todos/2026-06-02-security-review-findings.md`; this task helps the alternative path but does not close that todo.
- How this plan avoids behavior-only convergence: The artifact path is made executable by build targets and templates, not just described in docs.
- Hidden second-redesign risk: low; no runtime protocol changes.
- Decision after review: ready for implementation.

## Documentation Owner

- Current fact sources to read:
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/文档规范.md`
- Fact sources to update after implementation:
  - `docs/operations/runtime-agent-deployment.md`
- Plan-only notes that must not become architecture source:
  - Task slice details and validation logs.
- Archive condition:
  - After implementation, validation, review, and absorption into operations docs.

## Validation

- Required proof: runtime-agent host binary build, runtime-agent config loader tests, backend image executable smoke, docs consistency, and workflow completion checks.
- Evidence owner: detailed commands and results are recorded in `## Validation Evidence`.

## Validation Plan

- Per-slice commands:
  - `cd code/backend && ./scripts/build-runtime-agent.sh`
- Integration commands:
  - `cd code/backend && docker build -t ctf-backend-runtime-agent-check .`
  - `docker run --rm --entrypoint /bin/sh ctf-backend-runtime-agent-check -c 'test -x /app/ctf-runtime-agent'`
- Manual checks:
  - Inspect Dockerfile copy path and systemd `ExecStart`.
- Commands intentionally skipped and why:
  - Full runtime-agent host-daemon E2E is not required in this slice; the behavior change is limited to config loading and packaging, while real rollout still needs operator-side certificates and Docker host access.

## Validation Evidence

- Command: `cd code/backend && ./scripts/build-runtime-agent.sh`
  - Result: exit 0; produced `code/backend/bin/ctf-runtime-agent` for `linux/amd64`.
  - Notes: `bin/` is ignored and not part of the patch.
- Command: `cd code/backend && go test ./cmd/runtime-agent ./internal/bootstrap ./internal/config -count=1`
  - Result: exit 0.
  - Notes: compiles the runtime-agent entrypoint and covers the config loader split.
- Command: `timeout 600 docker build -t ctf-backend-runtime-agent-check . && docker run --rm --entrypoint /bin/sh ctf-backend-runtime-agent-check -c 'test -x /app/ctf-runtime-agent' && docker image rm ctf-backend-runtime-agent-check >/dev/null`
  - Result: exit 0.
  - Notes: the image-contained runtime-agent executable was checked and the temporary image was removed.
- Command: `python3 scripts/check-docs-consistency.py`
  - Result: exit 0; `PASS — documentation references and diagram sources`.
  - Notes: `bash scripts/check-docs-consistency.py` was attempted first and failed because the file is a Python script; `python3` is the correct invocation.
- Command: `cd code/backend && go test ./internal/config -run TestLoadRuntimeAgentDoesNotRequireAPIProductionSecrets -count=1`
  - Result: red before implementation with `undefined: LoadRuntimeAgent`; exit 0 after implementation.
  - Notes: proves runtime-agent production loading no longer depends on API-only production secrets.
- Command: `cd code/backend && go test ./internal/config -count=1`
  - Result: exit 0.
  - Notes: confirms the config loader split did not break existing config package tests.
- Command: `cd code/backend && APP_ENV=prod bin/ctf-runtime-agent` with `deploy/systemd/ctf-runtime-agent.env.example` sourced
  - Result: expected exit 1 at `runtime_agent_tls_init_failed` because `/etc/ctf/runtime-agent/server.pem` does not exist in the local environment.
  - Notes: confirms the example env reaches TLS loading instead of failing on API-only PostgreSQL / Redis / CORS / flag secret config.
- Command: `timeout 900 bash scripts/check-workflow-complete.sh`
  - Result: exit 0; workflow-governance and completion-full stages passed.
  - Notes: The task worktree needed `.claude/skills -> ../.agents/skills` to satisfy project agent-entrypoint governance.

## Independent Review Handoff

- Review target:
  - Runtime-agent deployment artifact changes.
- Validation evidence summary:
  - Runtime-agent host build, Docker image executable check, docs consistency, and workflow governance all passed in this worktree.
- Architecture / contract inputs:
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/architecture/backend/01-system-architecture.md`
- Known risks / review focus:
  - Avoid implying containerized runtime-agent is the preferred safe production deployment.
  - Confirm env names match `CTF_*` mapping.
- Project-local checks to consider:
  - `bash scripts/check-workflow-governance.sh`

## Rollback / Recovery

- Safe revert boundary:
  - Revert build script, Dockerfile, systemd templates, docs, and this plan together.
- Data / config / runtime recovery notes:
  - No persistent data is changed.
- Irreversible operations:
  - None.

## Residual Risks

- Risk: The systemd template is not exercised against a real host daemon in this task.
- Why acceptable: It only wraps the existing command and uses existing config keys; true host rollout still needs operator-side certificate and daemon checks.
- Follow-up owner, if any: Runtime deployment runbook validation.
