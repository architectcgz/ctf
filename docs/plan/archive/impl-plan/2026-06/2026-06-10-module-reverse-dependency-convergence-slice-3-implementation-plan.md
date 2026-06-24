<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 模块反向依赖收口：边2 instance 经 contest 准入查询 Implementation Plan

**Goal:** Remove the production `instance -> contest` module dependency while keeping AWD proxy, instance visibility, startup recovery, and proxy traffic behavior unchanged.

**Architecture:** `instance` keeps consumer-side ports and neutral result/event structs. `contest/infrastructure` owns contest/AWD permission and traffic persistence rules, and app composition adapts contest repositories into the instance-facing ports. Remaining instance read models may read persisted contest table columns, but production instance code must not import `contest/contracts`, `contest/entity`, or any contest package.

**Tech Stack:** Go backend, GORM repositories, module architecture tests, CTF modular-monolith composition.

---

## Task Metadata

- Task Slug: `2026-06-10-module-reverse-dependency-convergence-slice-3`
- Started At: `2026-06-10T09:32:24Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-module-reverse-dependency-convergence-slice-3`
- Branch: `task/2026-06-10-module-reverse-dependency-convergence-slice-3`
- Base: `task/2026-06-10-module-reverse-dependency-convergence-slice-2`

## Objective And Non-Goals

- Objective:
  - Move AWD attack proxy and defense SSH scope queries out of `instance/infrastructure` into `contest/infrastructure`, returning `instanceports` neutral scope structs.
  - Build an app composition adapter so `ProxyTicketService` and `AWDDefenseSSHGateway` read `FindByID` from instance and AWD scope from contest.
  - Remove all production `instance` imports of `ctf-platform/internal/module/contest/...`.
  - Replace contest-owned event/result types used by instance startup recovery and proxy traffic with instance-owned neutral structs.
  - Remove `instance -> contest` from `moduleDependencyBaseline` only after imports are gone.
- Non-Goals:
  - Do not change tables, columns, migrations, routes, API payload shapes, proxy ticket shape, or AWD business rules.
  - Do not rewrite teacher/admin instance list behavior beyond import and owner cleanup.
  - Do not remove test fixture imports of contest contracts where they seed contest-owned tables.

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-module-reverse-dependency-convergence-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-runtime-residual-state-owner-split-plan.md`
  - `docs/design/backend-module-boundary-target.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/contest/infrastructure/contest_awd_runtime_recovery_repository.go`
  - `code/backend/internal/module/contest/infrastructure/awd_traffic_event_repository.go`
- Related prior work:
  - `7207c3343 refactor(backend): 收口 runtime agent checker 反向依赖`
  - `5b1c16f59 refactor(backend): 迁移 checker docker sandbox 执行 owner`

## Task Classification

- Classification: `非琐碎任务`
- Why: Backend structural refactor across `instance`, `contest`, and `app/composition`, changing repository owner boundaries and module dependency baseline.

## Files

- Create:
  - `code/backend/internal/module/contest/infrastructure/awd_proxy_scope_repository.go`
  - `code/backend/internal/module/contest/infrastructure/awd_proxy_scope_repository_test.go`
- Modify:
  - `code/backend/internal/module/instance/architecture_test.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/module/instance/contracts/errors.go`
  - `code/backend/internal/module/instance/api/http/handler.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
  - `code/backend/internal/module/instance/application/queries/instance_service.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/instance/infrastructure/awd_target_proxy_repository.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_builder.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/runtime_http_service_adapter.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
- Review:
  - `code/backend/internal/module/contest/infrastructure/awd_traffic_event_repository.go`
  - `code/backend/internal/module/contest/infrastructure/contest_awd_runtime_recovery_repository.go`
  - `code/backend/internal/module/instance/application/queries/proxy_ticket_service.go`
- Test:
  - `code/backend/internal/module/instance/architecture_test.go`
  - `code/backend/internal/module/contest/infrastructure/awd_proxy_scope_repository_test.go`
  - Existing instance startup recovery, proxy ticket, HTTP proxy traffic, and repository/query tests.

## 复用与 Owner 决策

- Existing patterns searched:
  - Instance consumer-side ports in `instance/ports/ports.go`.
  - Existing contest AWD repositories and runtime recovery repository.
  - Existing composition adapters in `app/composition/instance_module.go`.
  - Existing AWD proxy scope SQL in `instance/infrastructure/awd_target_proxy_repository.go`.
- Reuse / extend / split / create-new decision:
  - Reuse the existing SQL semantics by moving the AWD proxy scope implementation to contest infrastructure.
  - Create instance-owned neutral structs for startup recovery contest pause results and AWD proxy traffic events.
  - Keep `FindByID` in instance repo and compose it with contest scope queries in app composition instead of making instance import contest.
  - Keep user/teacher instance list queries in instance repository for this slice, but replace direct contest package imports with local persisted-value constants and a local snapshot read-model decoder.
- Owner boundary:
  - `contest`: AWD mode/status constants as persisted values, active round/scope-control admission rules, contest service visibility, AWD traffic persistence, contest pause extension.
  - `instance`: instance visibility rows, proxy ticket claims, proxy traffic event input shape consumed by its handler, startup recovery orchestration.
  - `app/composition`: private adapter that wires contest implementations into instance-facing ports.
- Why this is the narrowest safe surface:
  - It removes the import edge and the main owner-mixed AWD admission SQL without changing route behavior or schema.
  - It avoids a broad query rewrite of all instance list/read models in the same slice.
  - It keeps future reviewable work possible if list/read-model ownership needs further tightening.

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: This slice has an existing high-level plan, but the real code surface includes more production imports than the plan's headline proxy scope query.
- Evidence inspected:
  - `instance` production imports `contest/contracts` from startup recovery, HTTP handler, query service, infrastructure repository, and AWD proxy scope repository.
  - `runtime` residual split plan is `Implemented`; AWD runtime state and `awd_scope_controls` now live in contest.
  - `contest/infrastructure` already imports instance contracts in other AWD runtime repositories, so `contest -> instance` is an existing allowed direction.
  - `moduleDependencyBaseline` still contains `instance -> contest` and should fail once the edge is removed unless the baseline is updated.
- grill-with-docs findings:
  - Slice 3 can proceed because the runtime block 3/4 prerequisite is complete.
  - The completion condition is not just moving two methods; all production instance imports of contest must disappear.
  - Tests may continue importing contest fixtures because `archtest.RuntimeGoFiles` ignores `_test.go`.
- Plan adjustments after challenge:
  - Add a red architecture guard in `instance/architecture_test.go` forbidding production `instance` imports of contest packages.
  - Include proxy traffic and startup recovery neutral type changes.
  - Preserve `ErrAWDDefenseSSHUnavailable` API code/message/status by defining the same visible contract from instance.

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/instance -run TestProductionInstanceDoesNotImportContestModule -count=1` (red before implementation, green after)
  - `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TestAWDProxyScope|TestFindAWD' -count=1`
  - `cd code/backend && go test ./internal/module/instance/application/queries ./internal/module/instance/application/commands ./internal/module/instance/api/http ./internal/module/instance/infrastructure -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildInstanceModule|TestAWDDefenseSSH|TestRuntimeHTTP' -count=1`
  - `cd code/backend && go test ./internal/module/contest/... ./internal/module/instance/... -count=1`
  - `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
- Manual checks:
  - `rg 'ctf-platform/internal/module/contest' code/backend/internal/module/instance -g '*.go' -g '!**/*_test.go'` returns no production matches.
  - `rg '\"instance -> contest\"' code/backend/internal/module/architecture_baseline_test.go` returns no matches.
- Review focus:
  - No production `instance -> contest` imports remain.
  - Contest-owned AWD admission rules are not duplicated in instance production code.
  - App composition adapter does not leak contest types into instance packages.
  - API-visible AWD defense SSH unavailable error semantics stay unchanged.

## Validation Evidence

- `cd code/backend && go test ./internal/module/instance -run TestProductionInstanceDoesNotImportContestModule -count=1`: red before implementation on `api/http/handler.go`, green after production instance imports were removed.
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TestAWDProxyScopeRepository' -count=1`: pass.
- `cd code/backend && go test ./internal/module/instance/application/queries ./internal/module/instance/application/commands ./internal/module/instance/api/http ./internal/module/instance/infrastructure -count=1`: pass after updating HTTP recorder test stubs to the instance-owned event type.
- `cd code/backend && go test ./internal/app/composition -run 'TestBuildInstanceModule|TestBuildAWDDefenseSSHGateway|TestAWDDefenseSSHGateway|TestRuntimeHTTPServiceAdapter' -count=1`: pass.
- `cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`: pass after removing `instance -> contest`.
- `cd code/backend && go test ./internal/module/contest/... ./internal/module/instance/... -count=1`: pass.
- `cd code/backend && bash ../../scripts/check-backend-architecture.sh --full`: pass after updating `internal/testutil/systemapp` to use a composite proxy ticket reader.
- `bash scripts/run-workflow-stage.sh completion-full`: pass.
- `bash scripts/run-workflow-stage.sh pre-commit-quick`: pass with startup gate covering the staged diff.
- `rg 'ctf-platform/internal/module/contest' code/backend/internal/module/instance -g '*.go' -g '!**/*_test.go'`: no production matches.
- `rg --fixed-strings '"instance -> contest"' code/backend/internal/module/architecture_baseline_test.go`: no matches.

## Checklist

- [x] Add red architecture guard for production `instance` not importing contest.
- [x] Add contest-owned AWD proxy scope repository and migrate behavior tests.
- [x] Wire composite proxy ticket reader in composition for HTTP and SSH gateway paths.
- [x] Replace instance startup recovery and proxy traffic contest types with instance-owned neutral structs.
- [x] Remove remaining production contest imports from instance query/repository/service code.
- [x] Remove `instance -> contest` baseline edge.
- [x] Run validation and record evidence.
