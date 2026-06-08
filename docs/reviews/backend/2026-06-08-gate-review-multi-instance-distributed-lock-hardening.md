# 多实例分布式锁补齐 Gate Review

- Review target:
  - Repository: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-07-awd-defense-ssh-gateway-split/2026-06-08-multi-instance-distributed-lock-hardening`
  - Worktree: `task/2026-06-08-multi-instance-distributed-lock-hardening`
  - Diff source: 当前 worktree 未提交改动
  - Files reviewed:
    - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
    - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
    - `code/backend/internal/module/practice/application/commands/service.go`
    - `code/backend/internal/module/practice/infrastructure/scheduler_state_store.go`
    - `code/backend/internal/module/practice/infrastructure/cachekeys/redis_keys.go`
    - `code/backend/internal/module/practice/ports/ports.go`
    - `code/backend/internal/module/practice/runtime/module.go`
    - `code/backend/internal/config/config.go`
    - `code/backend/internal/config/config_test.go`
    - `code/backend/internal/shared/lockkeepalive/lockkeepalive.go`
    - `code/backend/internal/shared/lockkeepalive/lockkeepalive_test.go`
  - Related call chain checked:
    - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
    - `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
    - `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
    - `code/backend/internal/app/composition/instance_module.go`
    - `code/backend/internal/infrastructure/redis/redis.go`

- Classification check:
  - Agree with `非琐碎任务` classification.

- Gate verdict:
  - `blocked`

## Findings

1. **Blocker** `code/backend/internal/shared/lockkeepalive/lockkeepalive.go:39-55`, `code/backend/internal/config/config.go:427-431`, `code/backend/internal/config/config.go:487-519`
   - 风险：新的 scheduler/startup recovery 分布式锁允许把 TTL 配成任意正数，但 keepalive 刷新调用没有单次 deadline，只是把外层 `ctx` 原样传给 `lease.Refresh(...)`。一旦 Redis 刷新调用因为网络抖动或 Redis I/O timeout 被卡住超过 TTL，当前 owner 会在租约实际过期后继续运行，另一副本已经可以拿到同一把锁，出现双 leader / 双 scheduler。
   - 证据：
     - `lockkeepalive.Start` 只有在 `Refresh` 返回后才检查 `time.Since(lastConfirmedAt) >= ttl` 并 `runCancel()`；如果 `Refresh` 本身阻塞，fail-closed 逻辑不会提前生效。
     - 配置校验只要求 `container.startup_recovery_lock_ttl > 0` 且 `< 60s`，以及 `container.scheduler.lock_ttl > 0`；没有约束它们必须大于 Redis `read_timeout` / `write_timeout`，也没有给单次刷新包一个更短的 `context.WithTimeout(...)`。
     - 当前 Redis 客户端默认 `read_timeout` / `write_timeout` 是 `3s`，但这两个 TTL 现在都可以被配置成更小的值。
   - 触发条件：
     - 例如把 `container.scheduler.lock_ttl=1s`，或者把 `container.startup_recovery_lock_ttl=1s`；此时一次 Redis 刷新请求只要卡住到 TTL 之后才返回错误，旧 owner 会继续执行，而新 owner 已可抢锁。
   - 影响：
     - `practice_instance_scheduler` 会重新出现“同一轮容量判断被两个副本同时推进”的 split-brain。
     - `startup_runtime_recovery` 会重新出现双副本同时写 heartbeat / 执行 recovery 的窗口。
   - 修正方向：
     - 二选一至少做一个：
       - 在 `lockkeepalive.Start` 里为每次 `Refresh` 包一个不大于剩余 TTL 的子 deadline，确保阻塞调用也会在 TTL 内 fail-closed。
       - 或者在配置校验中强制 `startup_recovery_lock_ttl`、`scheduler.lock_ttl` 大于 Redis I/O timeout，并给出明确下界。
     - 更稳妥的做法是两者都做：运行时 deadline 保证 correctness，配置校验避免危险组合进入生产。

## Material findings

- 必修：
  - 修掉 `lockkeepalive` 刷新阻塞超过 TTL 时仍可能继续持有 owner 行为的 split-brain 风险。

## Senior implementation assessment

- 当前方案的 owner 收口方向是对的：practice scheduler 和 startup recovery 都被显式包进 Redis lease，且 release / keepalive 都抽成了共享逻辑。
- 但分布式锁 correctness 还差最后一层边界条件：现在只处理了“刷新返回 error/false”的场景，没有处理“刷新调用本身被卡住直到 TTL 之后才返回”的场景。对这类 lease 逻辑，单次刷新 deadline 不是优化项，而是 correctness 条件。

## Required re-validation

- 修复后至少补一类证据，再跑当前已有最小验证：
  - 新增一个 `lockkeepalive` 定向测试，覆盖“`Refresh` 调用阻塞超过 TTL`”后 run context 会在 TTL 内停止推进，或
  - 新增配置测试，证明危险 TTL/Redis-timeout 组合会被 `Validate()` 拒绝。
- 重新执行：
  - `go test ./internal/shared/lockkeepalive -count=1`
  - `go test ./internal/config -count=1`
  - `go test ./internal/module/practice/application/commands -run 'RunProvisioningLoop|DesiredAWD' -count=1`
  - `go test ./internal/module/practice/infrastructure -run 'StateStore|Scheduler|ReadinessProbe' -count=1`

## Residual risk

- 这次 review 按用户要求只覆盖指定文件和直接相关调用链，没有重新审 `contest` 侧旧的 keepalive 实现。
- 未独立复跑用户已提供的 Go 测试命令；本结论基于代码审读和用户给出的验证证据。

## Touched known-debt status

- 本次 touched surface 未命中当前 fact source 里要求在同面收口的已知结构债；阻塞点是新增分布式锁实现自身的 correctness 边界。
