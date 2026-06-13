<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# backend-async-safego Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为后台 goroutine 提供一个共享 SafeGo helper，统一捕获 panic、记录 stack，并把它接到应用层和基础设施层最风险的后台任务上。

**Architecture:** 新 helper 放在 `internal/shared/safego`，只负责 goroutine 启动、panic recovery 和日志记录，不承担业务调度、不保存业务状态。先迁移现有的后台任务入口，再用窄范围 architecture guardrail 防止重点文件回退到裸 `go func` / 无恢复 goroutine。

**Tech Stack:** Go, context.Context, zap, runtime/debug, go test, repository-local architecture tests

---

## Task Metadata

- Task Slug: `2026-06-13-backend-async-safego`
- Parent Task Group: `2026-06-13-backend-error-management-group`
- Slice Index: `4/13`
- Depends On: `2026-06-13-backend-context-logging-contract`
- Started At: `2026-06-13T03:41:28Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-async-safego`
- Branch: `task/2026-06-13-backend-async-safego`
- Plan Type: `slice`

## Plan Status

- Status: `review-passed`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 提供一个可复用的 SafeGo helper，后台 goroutine 出现 panic 时能按统一日志格式落盘并终止对应任务。
  - 把后台任务入口迁移到 SafeGo，优先覆盖 app composition 的循环后台 job、practice service 的异步任务，以及 instance cleaner 的后台启动路径。
  - 建立窄范围 guardrail，防止这几个高风险入口回退到裸 `go func`。
- Non-Goals:
  - 不在本 slice 统一替换全仓所有 `go func`。
  - 不改业务调度语义，也不把 panic 变成业务错误回传。
  - 不顺带重写 lock keepalive、HTTP recovery 或其他已有 panic 处理器。

## Problem Statement

- Current behavior / structure:
  - `internal/app/composition/root.go`、`internal/module/practice/application/commands/service_lifecycle.go` 和 `internal/module/instance/infrastructure/cleaner.go` 都各自直接启动 goroutine。
  - 这些 goroutine 没有共享的 panic recovery 入口，日志格式也不统一。
  - 现有 middleware recovery 已经记录 `panic`、`stack` 和 `request_id`，但后台任务没有对应的共享 helper。
- Target behavior / structure:
  - 新 helper 统一接住 goroutine panic，记录 `panic`、`stack`、任务名和必要的上下文字段。
  - `NewLoopBackgroundJob`、practice async task 和 cleaner 启动路径改用 helper，不再直接散落裸 goroutine 启动逻辑。
  - 针对这些入口的 guardrail 能阻止明显回退。
- Why this task is needed now:
  - Slice 3 已经把 request-scoped logging 契约铺好，slice 4 需要把它延伸到后台 goroutine。
  - 后续 slice 13 的关闭与 sleep 清理、以及更广泛的后台任务治理，都需要一个可复用的 SafeGo 基线。

## Inputs

- Source docs:
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-context-logging-contract-implementation-plan.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/app/backend_context_architecture_test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
- Related prior work:
  - `code/backend/internal/middleware/recovery.go`
  - `code/backend/internal/shared/lockkeepalive/*`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达共享 helper、app composition、application service 和 infrastructure 启动路径。
  - 需要独立 worktree、implementation plan、focused validation 和独立 review gate。

## Files

- Create:
  - `code/backend/internal/app/composition/background_job_loop.go`
  - `code/backend/internal/shared/safego/safego.go`
  - `code/backend/internal/shared/safego/safego_test.go`
  - `code/backend/internal/app/composition/root_safego_test.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle_safego_test.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner_safego_test.go`
- Modify:
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Review:
  - `code/backend/internal/middleware/recovery.go`
  - `code/backend/internal/shared/lockkeepalive/lockkeepalive.go`
- Test:
  - `code/backend/internal/shared/safego/safego_test.go`
  - `code/backend/internal/app/composition/root_test.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner_test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `internal/middleware/recovery.go`：已有 panic log 样式，记录 `panic` / `stack` / `request_id`。
  - `internal/shared/lockkeepalive`：已有共享后台协作 helper 的落点和测试风格。
  - `internal/app/composition/root.go`：已有 background job 启动包装。
  - `internal/module/practice/application/commands/service_lifecycle.go`：已有 async task 包装。
  - `internal/module/instance/infrastructure/cleaner.go`：已有后台运行入口。
- Reuse / extend / split / create-new decision:
  - 新建 `internal/shared/safego` 作为共享 goroutine recovery owner。
  - 复用 `zap` 和 `runtime/debug`，不把 panic recovery 分散到各个调用方。
  - 保留各业务模块自己的任务语义，只把 goroutine 启动和 panic recovery 收到 helper。
- Owner boundary:
  - `internal/shared/safego` 只负责 goroutine 启动、panic recovery、日志字段和任务完成同步。
  - request-scoped 字段继续由调用方传入的 logger owner 决定；`internal/shared` 不反向依赖 `internal/platform/*`。
  - `root.go`、`service_lifecycle.go`、`cleaner.go` 只决定“何时启动后台任务”和“任务本身做什么”。
- Why this is the narrowest safe surface:
  - 先把共享 recovery helper 建起来，再迁移三个最直接的后台入口，能提供后续 slice 复用点，同时避免全仓 `go func` 一次性重写。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这不是单点 bug，而是共享后台任务契约和 owner 边界设计，需要先确认 helper 落点、panic 日志格式和试点范围。
- grill-with-docs findings:
  - 后台任务入口已经分散在 app composition、practice service 和 instance cleaner。
  - middleware recovery 的 panic log 已经有现成字段样式，可作为 SafeGo 的日志基线。
  - shared helper 更适合落在 `internal/shared`，而不是 app / module 层的某个单点。
- Plan adjustments after challenge:
  - 试点范围收敛为 `root.go`、`service_lifecycle.go`、`cleaner.go` 三个最直接后台入口。
  - guardrail 先锁这三个文件，不把全仓 `go func` 一次性纳入。

## Execution Slices

### Slice 1: SafeGo helper 基线

- Goal:
  - 新建共享 SafeGo helper，并用测试锁住 panic recovery 行为。
- Dependencies:
  - 无
- Files:
  - Create:
    - `code/backend/internal/shared/safego/safego.go`
    - `code/backend/internal/shared/safego/safego_test.go`
  - Review:
    - `code/backend/internal/middleware/recovery.go`
  - Test:
    - `code/backend/internal/shared/safego/safego_test.go`
- Steps:
  - [x] Step 1: 写 failing tests，证明 helper 会捕获 panic、记录 stack，并在 nil logger / nil ctx 下保持可预期。
  - [x] Step 2: 跑最小测试确认红灯。
  - [x] Step 3: 实现 `Go` / `GoWait` / `GoWithRecover` 之类最小 API。
  - [x] Step 4: 跑 helper tests，确认转绿。
- Validation:
  - `go test ./internal/shared/safego -count=1`
- Review focus:
  - panic recovery 日志字段要和 middleware recovery 保持一致的语义。
  - helper 不保存业务状态，不引入额外调度策略。
- Done criteria:
  - helper 能稳定处理 panic 和正常返回。
  - helper 的日志契约有单测覆盖。

### Slice 2: 试点迁移与 guardrail

- Goal:
  - 把三个最高风险后台入口迁移到 SafeGo，并加窄范围 guardrail。
- Dependencies:
  - Slice 1
- Files:
  - Create:
    - 无
  - Modify:
    - `code/backend/internal/app/composition/root.go`
    - `code/backend/internal/app/composition/root_test.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner_test.go`
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
  - Review:
    - `code/backend/internal/app/composition/root.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - Test:
    - 对应 package 现有单测 + 新增 panic / completion / guardrail tests
- Steps:
  - [x] Step 1: 先为三个试点写 failing tests，锁住 panic recovery 或 SafeGo 路径。
  - [x] Step 2: 跑 targeted tests，确认红灯。
  - [x] Step 3: 用 SafeGo 迁移三个后台入口。
  - [x] Step 4: 在 architecture test 中为三个试点加回退 guardrail。
  - [x] Step 5: 同步 task group index 当前事实。
  - [x] Step 6: 跑 focused tests 与 architecture guardrail，确认转绿。
- Validation:
  - `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure -count=1`
  - `go test ./tests/architecture -run 'Test.*SafeGo|Test.*Background.*' -count=1`
- Review focus:
  - 只改后台任务契约，不顺带改变业务调度语义。
  - guardrail 先锁试点文件，不扩大到全仓。
- Done criteria:
  - 三个试点后台入口都通过 SafeGo 启动。
  - panic recovery 日志契约统一。
  - task group index 与当前实际状态一致。

## Impact And Compatibility

- API / DTO:
  - 无公开 API 结构变化。
- Data / migration:
  - 无。
- State / cache / queue / event:
  - 无持久化协议变化；仅补 goroutine 级 panic recovery。
- Runtime / config:
  - 无新增配置项。
- Frontend route / state / UX:
  - 无。
- Docs / contracts:
  - task group index 状态更新；如 helper owner 明显影响架构事实，再补 backend architecture 文档。

## Plan Review / Architecture Fit

- Target owner boundary:
  - SafeGo owner 在 `internal/shared/safego`，后台任务语义 owner 留在各自业务包。
- Reuse points / landing zones:
  - 复用 middleware recovery 的 panic log 语义。
  - 复用 shared helper 风格（如 lockkeepalive）。
  - 复用现有测试里的 goroutine / cancellation 断言风格。
- Known structural debt touched:
  - `root.go`、`service_lifecycle.go`、`cleaner.go` 的裸 goroutine 启动是本 slice 的直接治理对象。
- How this plan avoids behavior-only convergence:
  - 不是只把 panic 包成日志，而是把共享 recovery owner 抽出来，再把三个高风险入口迁移进去。
- Hidden second-redesign risk:
  - 如果后续希望把更多后台任务统一纳入 SafeGo，helper API 需要保持简单，避免演化成重量级任务框架。
- Decision after review:
  - 通过；按“共享 recovery owner + 三个试点 + 窄 guardrail”执行。

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Fact sources to update after implementation:
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
  - 如 helper owner 需要沉淀，再更新 `docs/architecture/backend/04-api-design.md` 或相关 backend 架构事实源。
- Plan-only notes that must not become architecture source:
  - “先试点三个后台入口”的 rollout 说明属于实施策略，不是长期架构事实。
- Archive condition:
  - focused tests、`completion-full`、独立 backend review 与 `workflow-governance` 通过后归档到 `docs/plan/archive/impl-plan/2026-06/`。

## Validation Plan

- Per-slice commands:
  - Slice 1: `go test ./internal/shared/safego -count=1`
  - Slice 2: `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure -count=1`
  - Slice 2: `go test ./tests/architecture -run 'Test.*SafeGo|Test.*Background.*' -count=1`
- Integration commands:
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - 无额外手工操作；goroutine 行为和 panic recovery 用测试证明。
- Commands intentionally skipped and why:
  - 全仓 `go test ./...`：本 slice 只改共享 helper 与三个试点入口，先跑最小充分范围。

## Validation Evidence

- Command: `go test ./internal/shared/safego -count=1`
  - Result: pass
  - Notes: 验证共享 helper 的 panic recovery、`stack` 字段、`request_id` 透传以及 nil context / nil logger fallback。
- Command: `go test ./internal/app/composition -run 'TestNewLoopBackgroundJobLogsPanicThroughSafeGo|TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
  - Result: pass
  - Notes: 验证 loop background job 试点经 SafeGo 启动，panic 通过 logger 记录，原始 context 传递保持不变。
- Command: `go test ./internal/module/practice/application/commands -run 'TestRunAsyncTaskRequiresSafeGoForPanicRecovery|TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
  - Result: pass
  - Notes: 验证 practice 异步任务路径不再因 panic 直接炸进程，并保留 close / cancel 行为。
- Command: `go test ./internal/module/instance/infrastructure -run 'TestCleanerStartRunOnceRequiresSafeGoRecovery|TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
  - Result: pass
  - Notes: 验证 cleaner 试点后台启动与 stop/wait 路径经 SafeGo 包装后不破坏现有生命周期语义。
- Command: `go test ./tests/architecture -run 'TestSafeGoContractPilots' -count=1`
  - Result: pass
  - Notes: 确认三个试点文件已导入 `internal/shared/safego`，并禁止回退到裸 `go func()`.
- Command: `go test ./internal/shared/safego ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
  - Result: pass
  - Notes: 对 touched surface 跑完整 focused package tests，确认 helper、3 个试点入口和源码 guardrail 一起转绿。
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: pass
  - Notes: 通过 code change contracts、backend architecture、app composition boundaries 和 backend test architecture gate。
- Command: `timeout 240s bash scripts/check-workflow-governance.sh`
  - Result: pass
  - Notes: 通过 workflow governance stage；open todos 仅作为提醒输出，无 blocker。

## Independent Review Handoff

- Review target:
  - SafeGo helper 的 panic recovery 语义、三个试点入口的迁移、guardrail 是否只锁试点文件。
- Validation evidence summary:
  - focused shared helper tests
  - app composition / practice / cleaner focused tests
  - architecture guardrail
  - `completion-full`
- Architecture / contract inputs:
  - `AGENTS.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Known risks / review focus:
  - panic recovery 日志字段是否一致
  - helper 是否引入调度语义或业务状态
  - guardrail 是否只锁试点，不误伤非试点文件
- Project-local checks to consider:
  - `go test ./tests/architecture -run 'Test.*SafeGo|Test.*Background.*' -count=1`
  - `go test ./internal/shared/safego ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure -count=1`

## Rollback / Recovery

- Safe revert boundary:
  - 单次 task branch / merge commit 可整体回退。
- Data / config / runtime recovery notes:
  - 无数据迁移；回退后仅恢复到旧 goroutine / panic 处理行为。
- Irreversible operations:
  - 无。

## Residual Risks

- Risk:
  - 本 slice 只覆盖三个高风险后台入口，仓库内其他裸 goroutine 仍然存在。
- Why acceptable:
  - 这是分片推进的第一步，先建立共享 owner 和试点 guardrail，再决定后续扩面范围。
- Follow-up owner, if any:
  - 后续 `slice 13` 或额外后台任务 slice。
