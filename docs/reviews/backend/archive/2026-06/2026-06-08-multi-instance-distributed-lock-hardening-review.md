# Multi-instance Distributed Lock Hardening Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-07-awd-defense-ssh-gateway-split/2026-06-08-multi-instance-distributed-lock-hardening`
- Branch: `task/2026-06-08-multi-instance-distributed-lock-hardening`
- Task slug: `2026-06-08-multi-instance-distributed-lock-hardening`
- Review scope:
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
  - `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
  - `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/shared/lockkeepalive/lockkeepalive.go`
  - `code/backend/internal/shared/lockkeepalive/lockkeepalive_test.go`
- Related call chain read during review:
  - `code/backend/internal/app/http_server.go`
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/app/composition/practice_module.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/infrastructure/redislock/lock.go`

## Classification Check

同意按 `非琐碎任务` 的独立 gate review 处理。

## Gate Verdict

`blocked`

## Findings

1. `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go:148-161`
   - `Start()` 现在只在当前副本**首轮就抢到 recovery lock**时才等待 `initializeLeader()` 完成；拿不到锁的副本会立刻返回成功。
   - 但 `code/backend/internal/app/http_server.go:62-65,136-147` 会在所有 background job `Start()` 返回后继续启动 HTTP 服务，`code/backend/internal/app/router.go:89,157,163` 与 `code/backend/internal/app/composition/practice_module.go:23-28` 还会在同一次启动里继续注册并启动 practice scheduler。
   - 结果是：宿主机重启后的恢复窗口里，非 leader API 副本可以先进入 ready 并开始对外提供请求，甚至先跑 `practice_instance_scheduler`，而真正负责 `paused_seconds` 顺延、active runtime recovery、desired reconcile 的 leader 还没完成恢复。这会把恢复前的中间态暴露给流量，并允许 scheduler 与 startup recovery 并发运行。
   - 这是 correctness / regression blocker。之前单实例语义是 HTTP server 在 startup recovery 完成前不会启动；现在多副本下丢了这个门禁。

2. `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go:282-299,396-405`
   - 这次把 `platform_runtime_state` heartbeat 改成了 **leader-only 写入**，但 stale heartbeat 仍然被直接解释成 “runtime outage” 并触发 `recoverFromRuntimeOutage()`。
   - 结合 `code/backend/internal/shared/lockkeepalive/lockkeepalive.go:39-53` 与 `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go:331-348`，只要发生一段超过 60 秒的无 leader 窗口（例如 Redis 锁刷新/获取连续失败、leader 抖动、所有 API 副本的 leader loop 暂时都没持锁），下一次拿到锁的副本就会在 **同 boot_id** 下误判为宿主机 outage。
   - 误判后会执行 `recoverFromRuntimeOutage()`，把 `paused_seconds`、实例 `expires_at` 和 AWD desired reconcile 一起推进。这相当于把“leader/Redis 控制面故障”错误地升级成“宿主机运行时停机”。
   - 这是 distributed locking blocker。锁层的 leaderless gap 现在和真正的 runtime outage 共用同一个 heartbeat 事实源，语义已经不再等价。

## Material Findings

- 必须恢复一个全局启动门禁：在 recovery leader 完成首次 `initializeLeader()` 前，其它副本不能把 API 服务和后续 scheduler 当作 ready。
- 必须把 “leader liveness / lease continuity” 和 “runtime outage heartbeat” 的语义拆开，或者至少让同 boot_id 的 stale heartbeat 不会仅因 leaderless gap 就触发 outage recovery。

## Non-blocking Suggestions

- 无。

## Required Re-validation

修复后至少补这些证据：

```bash
cd code/backend
go test ./internal/module/instance/application/commands -run StartupRuntimeRecoveryService -count=1
go test ./internal/module/practice/application/commands -run 'RunProvisioningLoop|DesiredAWD' -count=1
go test ./internal/app -run 'TestNewHTTPServer|TestBuildRouterRuntime' -count=1
```

还需要新增或补强行为测试，至少覆盖：

- 非 leader 副本在 leader 完成 startup recovery 前不会进入对外可服务状态，或不会启动后续 scheduler。
- 同 boot_id 下仅出现 leaderless / lock-acquire gap 时，不会错误触发 `recoverFromRuntimeOutage()`。

## Senior Implementation Assessment

- 复用现有 Redis lease + keepalive 方向本身是合理的，`lockkeepalive` 抽到 shared 也能减少重复逻辑。
- 但当前实现把“谁能执行恢复”和“系统是否已经恢复完毕”混成了一个局部 leader loop；同时又继续用 leader heartbeat 代表 runtime outage 事实，导致启动门禁和 outage 语义都发生了漂移。
- 更低风险的做法是：
  - 明确一个独立的 startup recovery readiness gate，让所有 API 副本在该 gate 完成前都不对外 ready。
  - 保持 runtime outage 判断只依赖真正代表宿主机/runtime 连续性的事实源，而不是 leader-only heartbeat。

## Residual Risk

- 本次按用户要求只审指定文件及相关调用链，没有重审 practice 模块其它未列出的改动。
- 已知的两个已修 blocker（refresh error 超 TTL fail closed、`startup_recovery_lock_ttl < stale heartbeat window`）本次未重复报。

## Touched Known-debt Status

- 本次 touched surface 直接涉及多实例分布式锁 owner 和恢复门禁，当前仍有未收口的 correctness / distributed locking blocker，不能以下轮 follow-up 形式带过。
