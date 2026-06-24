# Review 对象

- 仓库：`/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-awd-defense-ssh-gateway-split`
- 分支：`task/2026-06-07-awd-defense-ssh-gateway-split`
- diff 来源：当前 worktree 未提交改动
- 评审范围：
  - `code/backend/internal/service/health/service.go`
  - `code/backend/internal/service/health/service_test.go`
  - `code/backend/internal/handler/health/handler.go`
  - `code/backend/internal/handler/health/handler_test.go`
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/app/http_server.go`
  - `code/backend/internal/app/http_server_test.go`
  - `code/backend/internal/app/full_router_integration_test.go`
  - `docker/ctf/docker-compose.dev.yml`
  - `README.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-07-awd-defense-ssh-gateway-split-implementation-plan.md`

# 分类检查

- 结论：同意 `非琐碎任务` 分类，并按 `code-workflow` 独立 review gate 处理。

# 结论

- Gate verdict：`blocked`
- 原因：`/live`、`/ready`、`/health` 的 endpoint 语义本身基本符合目标，但当前 API 启动链路仍会在监听 HTTP 之前同步等待 `startup_runtime_recovery` 的 leader-ready。这样 standby API 副本在依赖健康时仍可能无法及时提供 `/ready`，与“readiness 不依赖 startup recovery / lock owner”的任务约束冲突。

> 复核状态：已由实现方修复，等待独立 re-review 更新最终 verdict。
> 修复摘要：standby 副本未拿到 startup recovery lock 时，`StartupRuntimeRecoveryService.Start()` 只启动后台选举循环并立即返回；初始拿到 lock 的 leader 仍同步完成恢复初始化后再返回。
> 新增验证：`TestStartupRuntimeRecoveryServiceStandbyReplicaStartsWithoutLeaderReady`、`TestStartupRuntimeRecoveryServiceStandbyReplicaCanTakeOverAfterAsyncStart`、`TestNewHTTPServerDoesNotWaitForStartupRecoveryLeaderReadiness`。

# Findings

## Blocker

1. `code/backend/internal/app/http_server.go` 这次新增了 shutdown draining，但没有解除 API 启动对 `startup_runtime_recovery` leader-ready 的同步等待，导致 standby 副本仍可能在健康依赖下无法对外提供 `/ready`。
   - 位置：
     - `code/backend/internal/app/http_server.go:39`
     - `code/backend/internal/app/http_server.go:65`
     - `code/backend/internal/app/composition/instance_module.go:91`
     - `code/backend/internal/app/composition/instance_module.go:99`
     - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go:109`
     - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go:152`
     - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go:318`
     - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go:552`
   - 触发条件：
     - 当前实例没有拿到 startup recovery 锁。
     - leader 还没有把 runtime state 写到 `observedLeaderReady()` 可观察的 ready 状态。
     - `NewHTTPServer()` 仍按当前顺序同步调用 `server.startBackgroundJobs()`。
   - 影响：
     - standby API 副本在 Postgres/Redis 健康时，仍可能因为启动阶段卡在 `startupRecovery.Start(ctx)` 而没有开始监听 HTTP。
     - `/ready` endpoint 即使实现正确，也无法被 LB / compose 命中，因此不能满足“standby 实例只要依赖健康且未 draining 就 ready”。
   - 证据：
     - `HTTPServer.NewHTTPServer()` 在返回前同步执行 `server.startBackgroundJobs()`。
     - `instance_module.go` 把 `startupRecovery.Start` 直接注册为 background job 的 `start`。
     - `StartupRuntimeRecoveryService.Start()` 会等待 `initReady`，而 standby 分支会在 `observedLeaderReady()` 为真前持续等待。
     - 现有测试 `TestStartupRuntimeRecoveryServiceStandbyReplicaWaitsForLeaderReady` 明确锁定了这个等待行为。
   - 修正方向：
     - 让 startup recovery 的 leader 选举 / 观察逻辑不再阻塞 API 监听启动。
     - 最小安全修法不是改 `/ready` handler，而是调整 startup recovery 的启动语义或 HTTP server 的启动编排，使 standby 副本可以先启动 HTTP，再由后台异步完成 recovery 选举与跟随。

# Material findings

- 上述 Blocker 必须先修复，当前 diff 才能满足本次 task context。

# Non-blocking suggestions

- 暂无需要单独阻塞的次要建议。

# 必须重跑的验证

- `cd code/backend && go test ./internal/service/health ./internal/handler/health ./internal/app -run 'TestCheckLive|TestCheckReady|TestGetLive|TestGetReady|TestHTTPServerShutdownMarksReadiness|TestFullRouter' -count=1`
- `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryServiceStandbyReplicaWaitsForLeaderReady|TestStartupRuntimeRecoveryServiceStandbyReplicaSkipsRecoveryWithoutLeaderLock' -count=1`
- 修复后新增一条直接证明“standby API 不等待 startup recovery leader-ready 也能开始提供 `/ready`”的 app/bootstrap 级测试。
- `docker compose -f docker/ctf/docker-compose.dev.yml config`

# Senior implementation assessment

- 当前 diff 对 endpoint 语义的拆分是直接且可维护的：`health` service 统一 owner live/ready/health 状态，`http_server` 持有 draining state，`compose` 改 `/ready` 也合理。
- 更低风险的完整实现应把“HTTP 接流量 readiness”与“startup recovery leader ready”彻底解耦。否则 `/ready` 只是新增了 endpoint，没真正改变多实例 standby 副本的接流量行为。

# 残余风险

- 本次独立验证已复跑目标测试和 `docker compose ... config`。
- 项目本地 `bash scripts/check-backend-architecture.sh --full` 当前失败，但失败点在未纳入本次 diff 的 `internal/bootstrap/awd_defense_ssh_gateway.go` 中多处 `context.Background()`，属于当前 branch 的额外问题，不作为这次未提交 diff 的直接 finding。

# Touched known-debt status

- 未发现本次评审范围内另有已登记且被本次 diff 触达却未收口的结构债 surface。

## Re-review Update (2026-06-08)

- Gate verdict：`pass`
- 上一轮 blocker 已修复：
  - `StartupRuntimeRecoveryService.Start()` 在初始未拿到 startup recovery lock 时，只启动后台 leader election loop 后立即返回，不再等待 leader ready。
  - 初始 leader 仍同步等待 `initializeLeader()` 完成后再返回，恢复初始化与 safe failover 语义保持不变。
  - `NewHTTPServer()` 新增 app 级测试，证明 foreign owner 持有 startup recovery lock 时不会阻塞 HTTP server 构建。
- 独立复跑验证：
  - `go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryServiceStandbyReplicaStartsWithoutLeaderReady|TestStartupRuntimeRecoveryServiceStandbyReplicaCanTakeOverAfterAsyncStart|TestStartupRuntimeRecoveryServiceLeaderFailoverWithinSafeWindowSkipsRecovery|TestStartupRuntimeRecoveryServiceSameBootLeaderGapSkipsRecovery' -count=1`
  - `go test ./internal/app -run 'TestNewHTTPServerDoesNotWaitForStartupRecoveryLeaderReadiness|TestHTTPServerShutdownMarksReadinessDrainingBeforeStoppingBackgroundJobs|TestNewHTTPServerBuildsAndShutsDown' -count=1`
  - `go test ./internal/service/health -run 'TestCheckLive|TestCheckReady' -count=1`
  - `go test ./internal/handler/health -run 'TestGetLive|TestGetReady' -count=1`
  - `go test ./internal/app -run 'TestFullRouter' -count=1 -timeout 120s`
  - `docker compose -f docker/ctf/docker-compose.dev.yml config`
