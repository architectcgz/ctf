<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# Multi-instance Startup Recovery Gate Fix Implementation Plan

**Goal:** 修复 API 多副本启动时非 leader 副本绕过 startup recovery readiness gate 的问题，保证 HTTP server 与后续 scheduler 不会早于 runtime recovery 放行。

**Architecture:** 保持 `startup_runtime_recovery` 作为启动恢复 owner，继续复用 Redis `ctf:platform:runtime:recovery:lock` 与 `platform_runtime_state`。修复点收口在 `StartupRuntimeRecoveryService.Start()`：无论首轮是否拿到 lock，都等待“当前 boot_id 已有非 stale heartbeat”或当前副本成为 leader 并完成 `initializeLeader()` 后再返回。

**Tech Stack:** Go, Redis lease, GORM-backed runtime state, project-local code-workflow

---

## Task Metadata

- Task Slug: `2026-06-08-multi-instance-startup-recovery-gate-fix`
- Started At: `2026-06-08T11:33:49Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `main`
- Execution note: `start-workflow` originally targeted an isolated task worktree, but the actual diff was applied in the repository root worktree.

## Objective And Non-Goals

- Objective:
  - 让 standby API 副本在 leader 首轮 recovery 完成前阻塞 `Start()`，从而阻止 `NewHTTPServer()` 继续启动 HTTP 服务与后续 background jobs。
  - 保持同 `boot_id` stale heartbeat 只表示 leaderless gap，不触发 runtime outage recovery。
  - 更新反向测试，改为保护 startup recovery readiness gate。
- Non-Goals:
  - 不改造 `contest` 侧旧 keepalive。
  - 不改造 `runtime_cleaner`、`assessment_cleaner`、challenge publish check 或 image build worker 的多副本 owner。
  - 不引入新 Redis key、DB schema、MQ 或独立 readiness service。

## Inputs

- Source docs:
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/plan/impl-plan/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`
- Related architecture/contracts:
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
  - `code/backend/internal/app/http_server.go`
- Related prior work:
  - `docs/reviews/backend/2026-06-08-multi-instance-distributed-lock-hardening-review.md`
  - `docs/reviews/backend/2026-06-08-gate-review-multi-instance-distributed-lock-hardening.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: 触达 API 启动顺序、后台任务门禁和 runtime outage recovery 语义；需要 TDD、架构文档一致性检查和独立 review gate。

## Files

- Create: 无
- Modify:
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
  - `code/backend/internal/app/http_server_test.go`
- Review:
  - `code/backend/internal/app/http_server.go`
  - `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Test:
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
  - `code/backend/internal/app/http_server_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `StartupRuntimeRecoveryService.observedLeaderReady()` 已经表达“当前 boot_id heartbeat 可作为 leader ready 事实”。
  - `runLeaderElectionLoop()` 已有 `initReady` channel，可在 standby 路径复用等待逻辑。
  - `HTTPServer.NewHTTPServer()` 已串行启动 background jobs，startup recovery job 先于 practice scheduler 注册。
- Reuse / extend / split / create-new decision:
  - 扩展现有 `initReady` 等待语义，不新建 readiness service 或额外 Redis 状态。
  - 保持 `platform_runtime_state` 继续作为 startup recovery readiness 的事实源。
- Owner boundary:
  - `startup_runtime_recovery_service` 负责 startup recovery 是否完成的启动门禁。
  - `platform_runtime_state_store` 只负责 Redis load/save/lock，不承载业务判断。
  - `HTTPServer` 只按 background job `Start()` 返回与否决定是否继续启动，不内嵌 recovery 语义。
- Why this is the narrowest safe surface:
  - 当前文档目标与代码漂移集中在 `Start()` 首轮没拿到锁时直接返回；修复等待 channel 即可恢复门禁，不需要改变 recovery 状态机、DB schema 或路由 readiness contract。

## Intake Analysis Gate

- Relevant superpowers analysis pass: `systematic-debugging`
- Why this pass fits: 本轮从明确 review finding 出发，重点是修复已复现的启动门禁回归，不是重新设计多实例架构。
- grill-with-docs findings:
  - `docs/architecture/backend/03-container-architecture.md` 与 `docs/operations/awd-host-reboot-recovery-drill.md` 都要求非 leader 等待当前 `boot_id` heartbeat 后继续启动。
  - 当前 `Start()` 仅在 `initialAcquired == true` 时创建 `initReady`，导致 standby 直接返回，文档目标没有被代码实现。
  - 同 `boot_id` heartbeat gap 已被代码改成 warning，不再触发 outage recovery；本轮不应回退这部分语义。
- Plan adjustments after challenge:
  - 先写 RED 测试，把现有 `StandbyReplicaStartsWithoutLeaderReady` / `NewHTTPServerDoesNotWait...` 改成期望阻塞。
  - GREEN 阶段只把 `initReady` 改为所有 startup path 都创建并等待，直到 observed ready 或自己成为 leader。
  - 验证只覆盖 startup recovery 和 HTTP server 启动链路；相邻多实例后台任务风险进入 review residual，不混入本次修复。

## Validation

- Commands:
  - `go test ./internal/module/instance/application/commands -run 'StartupRuntimeRecoveryService.*Standby|StartupRuntimeRecoveryServiceLeader|SameBoot' -count=1`
  - `go test ./internal/module/instance/application/commands -run 'StartupRuntimeRecoveryServiceStandby' -count=1`
  - `go test ./internal/app -run 'TestNewHTTPServer.*StartupRecovery|TestHTTPServerBackgroundJobsStartAndStop' -count=1`
  - `go test ./internal/app -run 'TestNewHTTPServerWaitsForStartupRecoveryLeaderReadiness' -count=1`
  - `go test ./internal/module/instance/application/commands -count=1`
  - `go test ./internal/app -run 'TestNewHTTPServer|TestHTTPServer' -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
- Manual checks:
  - 核对 `docs/architecture/backend/03-container-architecture.md` 与实现一致，不再出现反向 guardrail。
  - 确认 standby 阻塞可被 context cancel 正常打断，避免 shutdown / failed startup 卡死。
- Review focus:
  - 非 leader 是否真的不能先返回。
  - leader ready 事实是否仍只来自当前 `boot_id` 非 stale heartbeat。
  - 是否引入死锁、goroutine 泄漏或 startup context 无法取消的问题。
