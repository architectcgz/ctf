<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# backend-application-service-boundary-convergence Implementation Plan

**Goal:** 收敛后端 application 根目录中的历史 service 形态，降低 CQRS / use-case 包混用带来的阅读噪音。

**Architecture:** 保持现有 Onion Architecture，不改变 API、DTO、持久化或 runtime 行为。第一批只处理低风险结构噪音：`container_runtime/application` 根目录的具体 service 迁入 `commands` / `queries` / `jobs`，删除未接线的 retired `instance/application/AWDDefenseWorkbenchService`。

**Tech Stack:** Go backend, module-level architecture tests, package-local Go tests.

---

## Task Metadata

- Task Slug: `2026-06-21-backend-application-service-boundary-convergence`
- Started At: `2026-06-21T09:19:23Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-21-backend-application-service-boundary-convergence`
- Branch: `task/2026-06-21-backend-application-service-boundary-convergence`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `review-passed` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 让 application 根目录只承载共享 contract/helper，不再混放具体 service 实现；删除已由架构测试标记为 retired 且无生产接线的 workbench service。
- Non-Goals: 不拆 `contest/application/commands/AWDService`，不拆 `container_runtime/application/commands/provisioning_service.go`，不改变 HTTP API、DTO、数据库、Redis、后台任务语义。

## Problem Statement

- Current behavior / structure: 多数模块已使用 `application/commands` 与 `application/queries`；`container_runtime/application` 根目录仍放 `ContainerFileService`、`ImageRuntimeService`、`ContainerStatsService`、`NodeHealthService`；`instance/application/AWDDefenseWorkbenchService` 仍保留测试与 contract，但 app composition 已有 guard 明确不再注入该 retired service。
- Target behavior / structure: `container_runtime/application` 根目录只保留 `contracts.go` 这类共享 application contract；写侧 runtime 能力进入 `commands`，只读 stats 进入 `queries`，后台 node health loop 进入 `jobs`；retired workbench service 和仅服务它的 contract/test 删除。
- Why this task is needed now: 用户指出 application 层 CQRS / 非 CQRS 混杂影响阅读；该 slice 先收敛最明显且低风险的历史噪音。

## Inputs

- Source docs: `docs/architecture/backend/01-system-architecture.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`
- Related architecture/contracts: `code/backend/internal/module/architecture_test.go`、`code/backend/internal/module/container_runtime/architecture_test.go`、`code/backend/internal/module/instance/architecture_test.go`
- Related prior work: `docs/todos/2026-06-11-backend-command-boundary-debt.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: 触达 backend module application 包边界、测试和 runtime wiring，属于受保护实现面。

## Files

- Create: `code/backend/internal/module/container_runtime/application/queries/*`、`code/backend/internal/module/container_runtime/application/jobs/*`
- Modify: `code/backend/internal/module/container_runtime/application/commands/*`、`code/backend/internal/module/container_runtime/runtime/module.go`、`code/backend/internal/app/composition/container_runtime_module.go`、相关测试与架构 guard
- Review: `code/backend/internal/module/instance/application/awd_defense_workbench_service.go`、`code/backend/internal/module/instance/contracts/services.go`
- Test: `go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/app/composition/...`

## 复用与 Owner 决策

- Existing patterns searched: `application/commands`、`application/queries`、`application/jobs` 在现有模块中的用法；`challenge/application/challengecore` 这类 use-case package；app composition 对 retired workbench 的 guard。
- Reuse / extend / split / create-new decision: 复用 commands/queries 命名；为后台 node health loop 创建 `application/jobs`，匹配 contest 现有 jobs 形态；删除无生产接线 retired code。
- Owner boundary: `container_runtime` 仍拥有 runtime capability service；app composition 只改 import 和 wiring；instance retired workbench 不再作为 application owner 暴露。
- Why this is the narrowest safe surface: 不触碰 provisioning 大文件、不改业务行为、不拆高风险 AWD service，只移动低耦合 service 和删除无生产接线代码。

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: 这是结构收敛方向选择，不是故障修复；需要先决定收敛粒度。
- grill-with-docs findings: 架构文档约束的是 Onion 依赖方向和 owner 边界，不要求所有 service 机械 CQRS；因此本 slice 不做全仓统一，而是先移除根目录 service 噪音。
- Plan adjustments after challenge: `ImageRuntimeService` 虽有 inspect/remove 两个方法，但实际只供 challenge command 路径使用，归入 `commands`；node health 是后台 loop，归入 `jobs`。

## Execution Slices

### Slice 1: container_runtime application 根目录 service 归位

- Goal: 将根目录具体 service 移入更明确的 application 子包，保留外部行为。
- Dependencies: 现有 package-local tests 作为回归护栏。
- Files:
  - Create:
    - `code/backend/internal/module/container_runtime/application/queries/container_stats_service.go`
    - `code/backend/internal/module/container_runtime/application/jobs/node_health_service.go`
  - Modify:
    - `code/backend/internal/module/container_runtime/application/commands/container_file_service.go`
    - `code/backend/internal/module/container_runtime/application/commands/image_runtime_service.go`
    - `code/backend/internal/module/container_runtime/runtime/module.go`
    - `code/backend/internal/app/composition/container_runtime_module.go`
    - `code/backend/internal/module/container_runtime/architecture_test.go`
  - Review:
    - `code/backend/internal/app/composition/runtime_node_failover_wiring_test.go`
  - Test:
    - `code/backend/internal/module/container_runtime/application/**/*_test.go`
- Steps:
  - [x] Step 1: Add architecture guard that blocks new `*Service` types in `container_runtime/application/*.go`.
  - [x] Step 2: Move file/image services to `application/commands` and adjust imports.
  - [x] Step 3: Move stats service to `application/queries` and node health loop to `application/jobs`.
  - [x] Step 4: Run focused container_runtime tests.
- Validation: `go test ./internal/module/container_runtime/...`
- Review focus: import cycles, behavior-preserving wiring, node health background job still registered.
- Done criteria: tests pass and root application package no longer contains concrete service implementations.

### Slice 2: retired instance workbench service 删除

- Goal: 删除已不接线的 `AWDDefenseWorkbenchService` 历史代码，避免读者误判它仍是当前 owner surface。
- Dependencies: Slice 1 independent.
- Files:
  - Create:
  - Modify:
    - `code/backend/internal/module/instance/contracts/services.go`
  - Delete:
    - `code/backend/internal/module/instance/application/awd_defense_workbench_service.go`
    - `code/backend/internal/module/instance/application/awd_defense_workbench_service_test.go`
  - Review:
    - `code/backend/internal/app/composition/architecture_test.go`
  - Test:
    - `code/backend/internal/module/instance/...`
- Steps:
  - [x] Step 1: Confirm production code has no `AWDDefenseWorkbenchService` references.
  - [x] Step 2: Delete retired service, test, and contract interface.
  - [x] Step 3: Run focused instance and app composition architecture tests.
- Validation: `go test ./internal/module/instance/... ./internal/app/composition/...`
- Review focus: no public API removal, no handler wiring loss, retired guard remains current.
- Done criteria: no remaining `AWDDefenseWorkbenchService` symbol outside unrelated historical text.

## Impact And Compatibility

- API / DTO: 无变化。
- Data / migration: 无变化。
- State / cache / queue / event: node health job wiring 保持原语义。
- Runtime / config: 无配置变化。
- Frontend route / state / UX: 无变化。
- Docs / contracts: implementation plan 记录本次结构收敛；不更新架构事实源，除非实现发现当前文档不准确。

## Plan Review / Architecture Fit

- Target owner boundary: `container_runtime` capability service 留在本模块 application 子包；`instance` retired workbench 不再暴露 application contract。
- Reuse points / landing zones: `commands` 放写侧/command-consumed runtime service，`queries` 放 stats read service，`jobs` 放 node health loop。
- Known structural debt touched: application 根目录具体 service；retired workbench service。
- How this plan avoids behavior-only convergence: 直接改变包落点并加 guard，不只是改注释或文档。
- Hidden second-redesign risk: `provisioning_service.go` 大文件仍是后续可读性债务，但本次不触碰；`contest/AWDService` 仍按 todo 单独跟踪。
- Decision after review: ready-for-implementation，按两个独立 slice 执行。

## Documentation Owner

- Current fact sources to read: `docs/architecture/backend/01-system-architecture.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`
- Fact sources to update after implementation: 若架构测试新增永久约束即可作为机械事实；无需修改架构文档。
- Plan-only notes that must not become architecture source: 文件移动步骤、临时验证命令。
- Archive condition: 代码合入并完成 review 后归档。

## Validation

- `go test ./internal/module/container_runtime/...`
- `go test ./internal/module/instance/... ./internal/testutil/... ./internal/app/composition/...`
- `go test ./internal/module/... ./internal/app/composition/...`
- `bash scripts/check-backend-architecture.sh`
- `bash scripts/check-workflow-complete.sh`

## Validation Plan

- Per-slice commands:
  - `go test ./internal/module/container_runtime/...`
  - `go test ./internal/module/instance/... ./internal/app/composition/...`
- Integration commands:
  - `go test ./internal/module/... ./internal/app/composition/...`
- Manual checks:
  - `rg -n "AWDDefenseWorkbenchService|container_runtime/application\""`
- Commands intentionally skipped and why: 不运行全量 E2E/runtime 容器测试；本次不改变 runtime 外部协作或容器生命周期行为。

## Validation Evidence

- Command: `go test ./internal/module/container_runtime -run TestApplicationRootDoesNotOwnConcreteServices -count=1`
  - Result: failed before migration as expected.
  - Notes: guard caught `application/container_file_service.go declares ContainerFileService`.
- Command: `go test ./internal/module/container_runtime/...`
  - Result: passed.
  - Notes: verified moved container runtime services and architecture guard.
- Command: `go test ./internal/module/instance/... ./internal/app/composition/...`
  - Result: passed.
  - Notes: verified retired workbench deletion and app composition guard.
- Command: `go test ./internal/module/instance/... ./internal/testutil/... ./internal/app/composition/...`
  - Result: passed.
  - Notes: rerun after same-context review removed residual retired workbench DTO/testutil methods.
- Command: `go test ./internal/module/... ./internal/app/composition/...`
  - Result: passed.
  - Notes: module integration range.
- Command: `bash scripts/check-backend-architecture.sh`
  - Result: passed.
  - Notes: backend module, shared, app composition, and test architecture boundaries.
- Command: `bash scripts/check-workflow-complete.sh`
  - Result: passed.
  - Notes: workflow governance and completion-full stages passed.

## Independent Review Handoff

- Review target: package moves under `container_runtime/application`, retired `instance/application/AWDDefenseWorkbenchService` deletion, new container runtime architecture guard.
- Validation evidence summary: focused package tests, module integration tests, backend architecture script, and workflow completion script all passed.
- Architecture / contract inputs: `docs/architecture/backend/01-system-architecture.md`, `docs/architecture/backend/07-modular-monolith-refactor.md`, module architecture tests.
- Known risks / review focus: ensure node health job semantics and runtime typed deps remain equivalent; ensure deleted workbench service was truly retired.
- Project-local checks to consider: `bash scripts/check-backend-architecture.sh`, `go test ./internal/module/... ./internal/app/composition/...`.
- Same-context review archive: `docs/reviews/backend/2026-06-21-backend-application-service-boundary-convergence-round-1.md`
- Independent review gate: not satisfied; no subagent tool was available in this session, and project review rules require a commit-bound independent review for the formal gate.

## Rollback / Recovery

- Safe revert boundary:
- Data / config / runtime recovery notes:
- Irreversible operations:

## Residual Risks

- Risk:
- Why acceptable:
- Follow-up owner, if any:
