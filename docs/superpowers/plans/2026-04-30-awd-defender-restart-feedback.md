# AWD Defender Restart Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the student AWD restart button actually restart the defender's own service and make attacker-facing access return controlled `503` during startup/restart instead of generic failures.

**Architecture:** Add an explicit restart command for contest AWD services, preserving the existing scoped instance identity while clearing runtime fields and requeuing it for provisioning. Extend the AWD target proxy read path so it can distinguish a missing/forbidden target from an existing non-running target and return a deliberate service-unavailable response. Update the student workspace to expose non-running own-service state and call the restart endpoint from the defense action.

**Tech Stack:** Go backend with Gin, GORM, Docker runtime orchestration, Vue 3 frontend, Vitest, Go `testing`.

---

## File Structure

- Modify: `docs/architecture/features/awd-student-battle-workspace-design.md`
  - Records the no-grace-window restart and proxy feedback contract.
- Modify: `code/backend/internal/module/practice/api/http/handler.go`
  - Adds the student `RestartContestAWDService` HTTP handler.
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
  - Adds the restart command and preserves scoped instance identity while requeuing.
- Modify: `code/backend/internal/module/practice/ports/ports.go`
  - Adds repository methods needed to lock, find, and reset an AWD scoped instance for restart.
- Modify: `code/backend/internal/module/practice/infrastructure/repository.go`
  - Implements the restart reset transaction helpers.
- Modify: `code/backend/internal/app/router_routes.go`
  - Registers `POST /contests/:id/awd/services/:sid/instances/restart`.
- Modify: `code/backend/internal/app/router_test.go`
  - Verifies the new route is registered to the practice handler.
- Modify: `code/backend/internal/module/runtime/ports/http.go`
  - Adds target runtime status to the AWD proxy scope.
- Modify: `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
  - Returns the latest scoped target instance even when it is not running.
- Modify: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
  - Converts existing non-running targets into explicit service-unavailable errors.
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
  - Keeps proxy behavior on service-unavailable errors deterministic.
- Modify: `code/backend/internal/module/contest/ports/awd.go`
  - Adds instance status to student workspace service instances.
- Modify: `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`
  - Includes active non-running instances for workspace display.
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
  - Exposes own-team `instance_id` and `instance_status` for pending/creating/failed services, while only marking attack targets reachable when running.
- Modify: `code/backend/internal/dto/contest_awd*.go`
  - Adds `instance_status` to the student workspace service DTO if needed.
- Modify: `code/frontend/src/api/contracts.ts`
  - Adds workspace service `instance_status`.
- Modify: `code/frontend/src/api/contest.ts`
  - Normalizes `instance_status` and adds `restartContestAWDServiceInstance`.
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - Calls restart endpoint, guards duplicate clicks, refreshes workspace, and displays pending/creating/failed states.
- Test: `code/backend/internal/module/practice/application/commands/service_test.go`
  - Restart preserves `service_id`, `nonce`, `host_port`, clears runtime fields, and requeues.
- Test: `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
  - Non-running target returns service unavailable, not forbidden/internal.
- Test: `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository_test.go`
  - Finds latest scoped target instance and exposes status.
- Test: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
  - Workspace includes own-team non-running instance status.
- Test: `code/frontend/src/api/__tests__/contest.test.ts`
  - Restart endpoint request and workspace normalization.

## Task 1: Document Contract

- [x] **Step 1: Update design document**

Run: `git diff -- docs/architecture/features/awd-student-battle-workspace-design.md`

Expected: Document states explicit restart API, no grace window, attacker-facing `503`, and `flag_mismatch` as the compromised boundary.

## Task 2: Backend Restart Command

- [ ] **Step 1: Write failing service test**

Add a test in `code/backend/internal/module/practice/application/commands/service_test.go` that starts from a running AWD scoped instance with `service_id`, `nonce`, `host_port`, `container_id`, `network_id`, `runtime_details`, and `access_url`.

Expected after restart:

- Same instance ID is returned.
- `status = pending` when scheduler is enabled.
- `service_id`, `nonce`, `host_port`, `contest_id`, `team_id`, and `challenge_id` are preserved.
- `container_id`, `network_id`, `runtime_details`, and `access_url` are empty.

- [ ] **Step 2: Run failing test**

Run:

```bash
go test ./internal/module/practice/application/commands -run 'TestRestartContestAWDService' -count=1
```

Expected: FAIL because restart API does not exist.

- [ ] **Step 3: Implement restart service and repository methods**

Add:

- `RestartContestAWDService(ctx, userID, contestID, serviceID int64)`
- scoped instance lookup for current user's AWD team service
- runtime cleanup before reset
- reset to `pending` with runtime fields cleared

- [ ] **Step 4: Run passing test**

Run:

```bash
go test ./internal/module/practice/application/commands -run 'TestRestartContestAWDService' -count=1
```

Expected: PASS.

## Task 3: Backend Route and Handler

- [ ] **Step 1: Add route test**

Update `code/backend/internal/app/router_test.go` to assert:

```text
POST /api/v1/contests/:id/awd/services/:sid/instances/restart
```

- [ ] **Step 2: Implement handler and route**

Add handler method to practice HTTP handler and register route next to the existing start endpoint.

- [ ] **Step 3: Run route test**

Run:

```bash
go test ./internal/app -run 'TestRouter' -count=1
```

Expected: PASS.

## Task 4: Proxy Controlled 503

- [ ] **Step 1: Write failing runtime tests**

Cover:

- Repository returns latest target instance with `status = pending`.
- Proxy ticket resolution for an existing non-running target returns `ErrServiceUnavailable`.

- [ ] **Step 2: Implement target status read path**

Add `Status` to `AWDTargetProxyScope`; repository should find the latest scoped instance without filtering to running only, but only return `AccessURL` for running instances.

- [ ] **Step 3: Convert non-running target to 503**

In `ResolveAWDTargetAccessURL`, return a service-unavailable app error when target exists but is not running or has no access URL.

- [ ] **Step 4: Run runtime tests**

Run:

```bash
go test ./internal/module/runtime/... -run 'AWDTarget|ProxyTicket' -count=1
```

Expected: PASS.

## Task 5: Workspace Instance State

- [ ] **Step 1: Write failing workspace test**

Add a test that seeds own-team `pending` or `failed` AWD instance and expects workspace service to include `instance_id` and `instance_status`.

- [ ] **Step 2: Implement status in repository/query/DTO**

Return active and failed own-team instance state for the student's services. Target services remain reachable only when `status = running` and `access_url` is non-empty.

- [ ] **Step 3: Run contest query tests**

Run:

```bash
go test ./internal/module/contest/application/queries -run 'Workspace|AWDService' -count=1
```

Expected: PASS.

## Task 6: Frontend Restart Action

- [ ] **Step 1: Add API tests**

Add tests for:

- `restartContestAWDServiceInstance` URL.
- `instance_status` normalization in workspace service data.

- [ ] **Step 2: Implement API and UI behavior**

Update `ContestAWDWorkspacePanel.vue`:

- Button calls restart endpoint.
- Duplicate clicks are guarded per service.
- Button label/state distinguishes `pending`, `creating`, `running`, and `failed`.
- Workspace refreshes after restart response.

- [ ] **Step 3: Run frontend focused tests**

Run:

```bash
npm --prefix code/frontend test -- contest.test.ts
```

Expected: PASS.

## Task 7: Final Verification

- [ ] **Step 1: Run focused backend tests**

Run:

```bash
go test ./internal/module/practice/application/commands ./internal/module/runtime/... ./internal/module/contest/application/queries ./internal/app -count=1
```

Expected: PASS.

- [ ] **Step 2: Run frontend focused tests**

Run:

```bash
npm --prefix code/frontend test -- contest.test.ts
```

Expected: PASS.

- [ ] **Step 3: Optional dev smoke**

If backend and frontend dev servers are running, smoke test:

- login as student
- open `/contests/8?panel=challenges`
- click restart on own AWD service
- confirm workspace shows startup state
- confirm attacker proxy for that service returns controlled `503` during restart

