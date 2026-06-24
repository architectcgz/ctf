<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# 后端 application commands 边界拆分 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or an equivalent task-by-task execution discipline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 收口本次检查确认的 P1 `application/commands` 边界问题：具体基础设施实现移到 infrastructure，经 ports 注入；过宽 practice command facade 拆成明确 use-case service；todo 与架构事实同步更新。

**Architecture:** 后端继续遵循模块化单体 Onion 边界：`api -> application -> ports/domain`，`infrastructure -> ports`，`runtime/app composition` 负责装配。commands 包只保留确实属于 command-side use case 的服务，不再承载 Docker CLI、LocalFS store、Redis lock concrete、host boot id reader，practice 不再通过单个 `Service` 挂载所有写路径。

**Tech Stack:** Go 1.x、Gin、GORM、Redis、LocalFS、Docker CLI adapter、项目现有 `code-workflow` / architecture tests。

---

## Task Metadata

- Task Slug: `2026-06-11-backend-command-boundary-decomposition`
- Started At: `2026-06-11T07:43:27Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-backend-command-boundary-decomposition`
- Branch: `task/2026-06-11-backend-command-boundary-decomposition`

## Objective And Non-Goals

- Objective:
  - 记录本轮发现的 `application/commands` 边界债务到 `docs/todos/`。
  - 将 challenge 中 Docker CLI image builder、artifact GC LocalFS 逻辑、AWD checker artifact store 移出 `application/commands`。
  - 将 assessment report 输出文件路径/存在性检查抽成 port，由 infrastructure LocalFS adapter 实现。
  - 将 instance startup recovery 的 Redis lock concrete 与 host boot id 文件读取抽成 ports。
  - 拆开 practice 的单一 `Service` facade，让 HTTP handler/runtime 显式依赖 instance lifecycle、submission、manual review 等 use-case service。
- Non-Goals:
  - 不改变 API 路由、响应 JSON shape、数据库 schema、Redis key 或任务调度语义。
  - 不处理 P2 观察项：contest AWDService 子用例拆分、container_runtime provisioning 文件拆分。
  - 不重写业务规则；只迁移 owner 和依赖边界。

## Inputs

- Source docs:
  - `docs/文档规范.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/reviews/backend/2026-06-11-gate-review-challenge-command-service-decomposition.md`
- Related architecture/contracts:
  - `code/backend/internal/module/*/architecture_test.go`
  - `code/backend/internal/module/{challenge,practice,assessment,instance}/ports`
  - `code/backend/internal/module/*/runtime`
- Related prior work:
  - `2026-06-11-challenge-command-service-decomposition` 已把普通题 core/import/self-check/publish-check/export 移出 `commands`，但 review 已记录 AWD checker artifact LocalFS 仍是残留。

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 跨 challenge/practice/assessment/instance 多模块，触达 application/infrastructure/runtime/api handler/architecture tests/docs。
  - 属于结构性 service/adapter 边界拆分，会影响编译依赖和后台任务 wiring。
  - 需要 completion-full、独立 review gate 和 workflow governance。

## Files

- Create:
  - `docs/todos/2026-06-11-backend-command-boundary-debt.md`
  - `code/backend/internal/module/challenge/infrastructure/docker_image_builder.go`
  - `code/backend/internal/module/challenge/infrastructure/awd_checker_artifact_store.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - `code/backend/internal/module/instance/infrastructure/boot_id_reader.go`
  - practice focused service packages/files as determined by slice 4.
- Modify:
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/artifact_gc_service.go` or its final landing package
  - `code/backend/internal/module/challenge/runtime/wiring.go`
  - `code/backend/cmd/storage-gc/main.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/commands/report_file_output.go`
  - `code/backend/internal/module/assessment/application/commands/report_generation.go`
  - `code/backend/internal/module/assessment/runtime/module.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/infrastructure/platform_runtime_state_store.go`
  - `code/backend/internal/module/instance/ports/ports.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/module/practice/api/http/handler.go`
  - `code/backend/internal/module/practice/runtime/module.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/todos/2026-06-11-backend-command-boundary-debt.md`
- Review:
  - `code/backend/internal/module/challenge/architecture_test.go`
  - `code/backend/internal/module/practice/architecture_test.go`
  - `code/backend/internal/module/assessment/architecture_test.go`
  - `code/backend/internal/module/instance/architecture_test.go`
- Test:
  - Focused package tests under affected modules.
  - `go test ./internal/module -run 'Test.*Architecture|TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1`
  - `bash scripts/check-backend-architecture.sh --full`

## 复用与 Owner 决策

- Existing patterns searched:
  - challenge 已有 `application/challengecore|challengeimport|challengeselfcheck|challengepublishcheck|challengepackageexport` use-case packages。
  - challenge LocalFS/zip storage 已有 `challenge/infrastructure/challenge_import_preview_store.go`、`challenge_attachment_store.go`、`challenge_package_storage.go`。
  - practice 已有 Redis/HTTP runtime concrete adapter 下沉样例：`infrastructure/submission_rate_limit_store.go`、`instance_readiness_probe.go`、`runtime_subject_repository.go`。
  - assessment 已有 renderer 从 commands 拆到 `application/reporting` 的 guard。
  - instance 已有 `ports` 作为 app/infrastructure 边界，但 startup recovery 还漏出 `redislock.Lock`。
- Reuse / extend / split / create-new decision:
  - challenge Docker CLI builder 直接迁到 infrastructure，实现既有 `challengeports.DockerImageBuilder`。
  - challenge AWD checker artifact 新增 `challengeports.AWDCheckerArtifactStore`，LocalFS 实现放 infrastructure；AWD import service 只调用 port。
  - challenge artifact GC 优先从 commands 移出；若保留 GC policy，需要让 LocalFS 读删行为落在 infrastructure，而不是 commands。
  - assessment 新增 `assessmentports.ReportOutputStore`，application 只请求 output path / download path，LocalFS 安全路径和 `os.Stat` 在 infrastructure。
  - instance 新增 `instanceports.StartupRuntimeStateStore` / `StartupRecoveryLockLease` / `HostBootIDReader`，application 不再 import `redislock` 或 `os`。
  - practice 拆 handler dependency surface 和 runtime wiring；响应 DTO 若被多个 service/handler/test 使用，优先提升到 `practice/contracts`，避免新 application packages 反向依赖 `application/commands`。
- Owner boundary:
  - application service owns business decisions, transaction choreography, event publication, retry/job policy.
  - infrastructure owns Docker CLI process execution, LocalFS path safety, env-backed roots, Redis lock concrete, host file reads.
  - runtime/composition owns concrete construction and background job registration.
- Why this is the narrowest safe surface:
  - 不改外部 API，不改数据库/缓存 key，不改业务状态机。
  - 每个 slice 可以用现有 tests + architecture guard 验证边界收口。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`：已基于代码扫描确认候选不是单纯行数问题，而是 owner/adapter 混放。
- Why this pass fits:
  - 当前是结构性重构，需要先确认每个模块的现有 owner、runtime wiring 和 guard。
- grill-with-docs findings:
  - `docs/architecture/backend/07-modular-monolith-refactor.md` 已明确 LocalFS/zip storage adapter 属于 infrastructure，且 challenge ordinary import/export 已完成该方向。
  - 同文档对 practice、assessment、instance 的描述强调 application 不应知道 Redis/GORM concrete 和 infrastructure sentinel；startup recovery 的 `redislock.Lock` 和 boot id 文件读取违反该方向。
  - `docs/architecture/backend/06-file-storage.md` 已把 challenge import/export LocalFS/zip 事实源写到 infrastructure，AWD checker artifact 是同类残留。
- Plan adjustments after challenge:
  - 不把 artifact/boot id/Redis lock 做成 commands 内 helper；统一走 ports + infrastructure。
  - `practice` 不只移动文件名，必须拆 handler/runtime 的 service dependency surface，否则宽 facade 仍存在。

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/challenge/... -count=1`
  - `cd code/backend && go test ./internal/module/assessment/... -count=1`
  - `cd code/backend && go test ./internal/module/instance/... -count=1`
  - `cd code/backend && go test ./internal/module/practice/... -count=1`
  - `cd code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
  - `bash scripts/check-backend-architecture.sh --full`
  - `python3 scripts/check-docs-consistency.py`
  - `git diff --check`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - `rg -n "os/exec|exec\\.Command|os\\.(ReadFile|WriteFile|MkdirAll|RemoveAll|Stat)|ctf-platform/internal/infrastructure/redislock" code/backend/internal/module/*/application/commands --glob '*.go' --glob '!**/*_test.go'`
  - `find code/backend/internal/module -path '*/application/commands/*.go' -not -name '*_test.go' -print0 | xargs -0 wc -l | awk '$1 > 1000 && $2 != "total" {print}'`
- Review focus:
  - No behavior drift in API response shape, task loops, lock release semantics, report download security, artifact deletion safety.
  - No new wide provider-owned ports.
  - No application package importing concrete infrastructure.

## Task Slices

### Slice 1: Challenge concrete adapters out of commands

- [x] Add failing/updated architecture guard that blocks Docker CLI, LocalFS artifact store helpers, and AWD checker artifact env/FS operations in `challenge/application/commands`.
- [x] Move Docker CLI image builder implementation and tests to `challenge/infrastructure`.
- [x] Add `AWDCheckerArtifactStore` port and LocalFS infrastructure implementation.
- [x] Inject the artifact store into `AWDChallengeImportService`; remove LocalFS/env helper code from commands.
- [x] Move or rehome artifact GC so `code/backend/cmd/storage-gc` no longer imports it from `application/commands`.
- [x] Re-run focused challenge tests.

### Slice 2: Assessment report output store

- [x] Add `assessmentports.ReportOutputStore`.
- [x] Implement LocalFS report output store in `assessment/infrastructure`.
- [x] Inject store into `ReportService`; remove `os`/path safety from commands.
- [x] Update report output tests and renderer structure guard.
- [x] Re-run focused assessment tests.

### Slice 3: Instance startup recovery ports

- [x] Add startup recovery lock lease/state store and host boot id reader ports.
- [x] Adapt `PlatformRuntimeStateStore` to return port lease instead of `*redislock.Lock`.
- [x] Add infrastructure boot id reader and inject it from app composition.
- [x] Update startup recovery tests with test boot id reader and no application redislock import.
- [x] Re-run focused instance tests.

### Slice 4: Practice focused service owners

- [x] Split HTTP handler service interface into instance lifecycle, submission, manual review dependencies.
- [x] Move shared practice response DTOs used by handler/tests into `practice/contracts`.
- [x] Extract submission service owner; keep flag submission/history behavior and score/event side effects.
- [x] Extract manual review service owner; keep review/list/get behavior and score/event side effects.
- [x] Extract instance lifecycle/AWD runtime owner; keep scheduler, desired reconcile, AWD scope control, provisioning, runtime helpers and background task closer.
- [x] Update runtime wiring to construct and pass focused services; remove single wide `practicecmd.Service` facade.
- [x] Add/adjust architecture guard proving the wide `type Service struct` is gone from commands and handler no longer accepts one all-method interface.
- [x] Re-run focused practice tests.

### Slice 5: Docs, todo, full validation, review

- [x] Update backend architecture/file-storage docs with final landed paths.
- [x] Mark completed P1 todo items as done and leave P2 open.
- [x] Run full validation commands.
- [ ] Run independent review gate and archive review evidence.
- [ ] Fix review findings, re-run impacted validation, then run workflow governance.
