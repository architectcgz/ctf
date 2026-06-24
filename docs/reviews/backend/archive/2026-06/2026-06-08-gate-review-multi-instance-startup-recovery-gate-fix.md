# Startup Recovery Gate Fix Gate Review

- Review target:
  - Repository: `/home/azhi/workspace/projects/ctf`
  - Task slug: `2026-06-08-multi-instance-startup-recovery-gate-fix`
  - Diff source: 当前主工作区未提交改动
  - Files reviewed:
    - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
    - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
    - `code/backend/internal/app/http_server_test.go`
    - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-startup-recovery-gate-fix-implementation-plan.md`
  - Related call chain checked:
    - `code/backend/internal/app/http_server.go`
    - `code/backend/internal/app/router.go`
    - `code/backend/internal/app/composition/instance_module.go`
    - `code/backend/internal/app/composition/root.go`
    - `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
    - `docs/architecture/backend/03-container-architecture.md`
    - `docs/operations/awd-host-reboot-recovery-drill.md`

- Classification check:
  - Agree with `非琐碎任务` classification.

- Gate verdict:
  - `pass`
  - 首轮 review verdict 为 `pass with minor issues`；minor finding 已通过新增 foreign leader heartbeat 放行测试收口。

## Findings

- 无未收口 finding。

## Resolved Findings

1. **Minor** `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go:517-625`, `code/backend/internal/app/http_server_test.go:197-245`
   - 风险：当前新增测试已经覆盖了“standby 在未 ready 时阻塞”和“standby 在拿到 lock 并完成初始化前不返回”，但还没有直接覆盖另一条同样关键的放行语义：standby 在自己始终拿不到 startup recovery lock 的情况下，应当在“其它 leader 写入当前 `boot_id` 的非 stale heartbeat”后放行。
   - 证据：
     - `TestStartupRuntimeRecoveryServiceStandbyReplicaWaitsForLeaderReady` 只验证了阻塞直到 `ctx` cancel。
     - `TestStartupRuntimeRecoveryServiceStandbyReplicaCanTakeOverBeforeStartReturns` 与 `TestNewHTTPServerWaitsForStartupRecoveryLeaderReadiness` 都是靠“本副本后来拿到 lock”这条路径结束等待。
     - `runLeaderElectionLoop()` 新行为的另一半 owner 其实是 `observedLeaderReady()` 分支，但这次 diff 没有直接把它锁成回归测试。
   - 影响：
     - 当前实现本身我没有看到错误，但后续如果有人调整轮询顺序、ready 判定或 `LoadPlatformRuntimeState()` 交互，现有测试不一定能及时报出“standby 不再根据 foreign leader heartbeat 放行”的回归。
   - 修正方向：
     - 补一个定向单测：让 standby 一直拿不到 lock，先返回 `boot_id` 不匹配或 stale 状态，再切到“同 `boot_id` + fresh heartbeat”，断言 `Start()` / `NewHTTPServer()` 在不抢锁的前提下放行。
   - Resolution:
     - 已新增 `TestStartupRuntimeRecoveryServiceStandbyReplicaReturnsAfterForeignLeaderReady`，覆盖 standby 始终拿不到 lock、等待其它 leader 写入当前 `boot_id` fresh heartbeat 后放行。
     - 复核 verdict: `pass`。

## Material findings

- 无。

## Senior implementation assessment

- 这次 owner 收口是对的：启动门禁继续放在 `StartupRuntimeRecoveryService.Start()`，`HTTPServer.NewHTTPServer()` 不内嵌 recovery 语义，只通过 background job 的 `Start()` 串行阻塞来继承 gate。
- 我额外核对了 `router.go -> composition/instance_module.go -> root.BackgroundJobs()` 这条链路，`startup_runtime_recovery` 仍然先于 `runtime_cleaner` 和后续 practice/contest/assessment jobs 注册并启动，所以这次改动确实能挡住 HTTP server 返回以及后续 job 启动，不只是把单个 service 的 `Start()` 改慢。

## Required re-validation

- 这次没有 blocker，不要求返修后重跑。
- 如果补上上面的定向测试，建议最小重跑：
  - `go test ./internal/module/instance/application/commands -run 'StartupRuntimeRecoveryServiceStandby' -count=1`
  - `go test ./internal/app -run 'TestNewHTTPServerWaitsForStartupRecoveryLeaderReadiness' -count=1`

## Residual risk

- 按用户要求，本次没有把 `contest` 旧 keepalive、`runtime_cleaner`、`assessment_cleaner` 的相邻多副本风险升格为 blocker。
- `completion-full` 和 `pre-commit-quick` 的通过结果来自实现上下文提供的证据；我独立补跑了两条最关键的回归测试：
  - `go test ./internal/module/instance/application/commands -run 'StartupRuntimeRecoveryServiceStandby' -count=1`
  - `go test ./internal/app -run 'TestNewHTTPServerWaitsForStartupRecoveryLeaderReadiness' -count=1`

## Touched known-debt status

- 本次 touched surface 命中了 2026-06-08 分布式锁 hardening review 里要求补回的 startup recovery readiness gate。
- 当前 diff 已把这项 gate blocker 收口回文档要求的 owner 边界；未看到把同一 surface 上的已知结构债继续留作 blocker 的情况。
