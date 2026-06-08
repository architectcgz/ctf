<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 多实例分布式锁补齐 Implementation Plan

**Goal:** 为 `startup_runtime_recovery` 和 `practice_instance_scheduler` 补齐多实例 leader / scheduler 分布式锁，避免 API 多副本部署时重复恢复或突破全局容量上限。

**Architecture:** 复用现有 Redis lease 模型，把“谁能执行全局后台编排”收口成显式 owner，而不是继续依赖单实例假设或只靠条件更新兜底。`startup_runtime_recovery` 改成 leader-only heartbeat/recovery；`practice_instance_scheduler` 改成单 owner 调度循环，在锁持有窗口内执行 desired reconcile 与 pending instance claim，并在长于 TTL 的执行窗口内做 lease keepalive。

**Tech Stack:** Go, Redis, GORM, miniredis, project-local code-workflow

---

## Task Metadata

- Task Slug: `2026-06-08-multi-instance-distributed-lock-hardening`
- Started At: `2026-06-08T04:02:27Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/.worktrees/2026-06-07-awd-defense-ssh-gateway-split/2026-06-08-multi-instance-distributed-lock-hardening`
- Branch: `task/2026-06-08-multi-instance-distributed-lock-hardening`

## Objective And Non-Goals

- Objective:
  - 为 `startup_runtime_recovery` 增加分布式 leader lock，确保同一时刻只有一个 API 实例负责平台 runtime heartbeat 与 outage recovery。
  - 为 `practice_instance_scheduler` 增加分布式 scheduler lock，确保 `max_concurrent_starts` / `max_active_instances` 在多实例下仍由单 owner 计算和 claim。
  - 保持现有 practice instance scope 锁、contest recovery 幂等账本和 pending/creating/running 状态机语义不变。
- Non-Goals:
  - 不把当前单机 Docker runtime 架构扩展成跨物理机 runtime owner 模型。
  - 不改造 challenge image build / publish check / assessment cleaner 等其它后台任务。
  - 不引入 MQ、数据库新表或新的后台服务进程。

## Inputs

- Source docs:
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/Q&A/容器状态机调度是怎么实现的.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Related architecture/contracts:
  - `code/backend/internal/infrastructure/redislock/lock.go`
  - `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
  - `code/backend/internal/module/runtime/infrastructure/cleaner.go`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-07-awd-defense-ssh-gateway-split-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-06-awd-service-operation-active-scope-recovery-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达后台调度 owner、Redis 锁配置、运行时恢复链路和 practice provisioning 主循环。
  - 需要补测试与架构事实源，且必须经过独立 review 才能确认没有引入新的死锁 / 错误恢复路径。

## Files

- Create:
  - 如有必要，新增 lock store / helper 文件，放在对应模块的 infrastructure / application 目录内。
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
  - `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
  - `code/backend/internal/module/practice/infrastructure/cachekeys/redis_keys.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/Q&A/容器状态机调度是怎么实现的.md`
- Review:
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/module/practice/runtime/module.go`
  - `code/backend/internal/module/runtime/infrastructure/cleaner.go`
  - `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- Test:
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `runtime_cleaner` 的 Redis 单实例锁
  - `contest_status_updater` / `AWDRoundUpdater` 的 lease + keepalive 模式
  - practice 现有 scope 级事务锁与状态 claim
- Reuse / extend / split / create-new decision:
  - 复用 `redislock.Acquire/Refresh/Release` 和现有 keepalive 设计。
  - `startup_runtime_recovery` 继续复用 `PlatformRuntimeStateStore` 承担 Redis 状态 owner，不新建额外 persistence owner。
  - `practice_instance_scheduler` 在 practice 模块内增加 scheduler lock owner，而不是把调度锁硬塞进 runtime cleaner 或 contest 锁抽象。
- Owner boundary:
  - `instance/startup_runtime_recovery_service` 负责“什么时候需要恢复”与“恢复执行顺序”。
  - `runtime/platform_runtime_state_store` 负责平台 runtime state + startup recovery lock 的 Redis 实现细节。
  - `practice/instance_provisioning_scheduler` 负责全局 provisioning 领取 owner。
- Why this is the narrowest safe surface:
  - 不改变 instance scope 锁、容器创建路径、DB schema 或 contest pause 账本，只把“全局后台 job 是否有权执行”补成显式 owner。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这是现有后台编排语义的收口，不是单点 bugfix；先确认哪些任务真的缺锁，哪些已经靠 DB claim 或事务锁足够安全。
- grill-with-docs findings:
  - 当前架构文档明确写的是单机 Docker runtime，但没有把多 API 副本下的 leader/scheduler owner 说清楚。
  - `startup_runtime_recovery` 共享 `platform_runtime_state`，不加 leader lock 就会把“谁写 heartbeat”与“谁执行恢复”暴露给所有实例。
  - `practice_instance_scheduler` 现有 `TryTransitionStatus(pending->creating)` 只能保证单实例不重复 claim 同一行，不能保证全局容量判断只算一次。
- Plan adjustments after challenge:
  - 先修两个 correctness 风险点。
  - `assessment_cleaner` 和 `instance_stopping_cleanup` 暂不纳入本任务，避免把“负载优化”和“必须修的 correctness”混成同一提交。

## Validation

- Commands:
  - `go test ./internal/module/instance/application/commands -run StartupRuntimeRecoveryService -count=1`
  - `go test ./internal/module/practice/application/commands -run 'RunProvisioningLoop|DesiredAWD' -count=1`
  - `go test ./internal/module/runtime/infrastructure -count=1`
  - `go test ./internal/module/contest/application/jobs -run 'Lock|Scheduler' -count=1`
- Manual checks:
  - 检查新增配置默认值与校验逻辑不破坏现有启动基线。
  - 检查 docs 事实源对“多实例下 single leader / single scheduler owner”表述是否一致。
- Review focus:
  - 锁丢失后的停止语义是否正确。
  - 长于 TTL 的执行窗口是否真的做了 keepalive。
  - 非 owner 实例是否只待命、不写 heartbeat、不误触发 recovery。
