# Multi-instance Governance Owner Review

- Review target:
  - Repository: `/home/azhi/workspace/projects/ctf`
  - Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-multi-instance-governance-owner-review`
  - Task slug: `2026-06-08-multi-instance-governance-owner-review`
  - Diff source: 当前 task worktree 未提交改动
  - Files reviewed:
    - `tools/multi-instance-nginx-proxy-smoke.sh`
    - `scripts/lib/multi-instance-nginx-proxy-smoke/run.sh`
    - `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
    - `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
    - `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
    - `code/backend/internal/module/runtime/infrastructure/cleaner.go`
    - `code/backend/internal/module/assessment/application/commands/cleaner.go`
    - `code/backend/internal/module/assessment/application/commands/profile_service.go`
    - `code/backend/internal/module/assessment/infrastructure/state_store.go`

- Classification check:
  - Agree with `非琐碎任务`.

- Gate verdict:
  - Initial independent gate: `pass with minor issues`
  - Final gate after minor test fix: `pass`

## Findings

- 无未收口 material finding。
- Independent reviewer minor issue: `runtime_cleaner` 已改用 `context.WithoutCancel(ctx)` 释放 cleanup lock，但初版测试只覆盖 lock loss，没有直接覆盖 shutdown 时 cleanup lock 仍会释放。
- Follow-up fix: 新增 `TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation`，覆盖 Stop 取消 base context 后 Redis `ctf:container:cleanup:lock` 会被释放。

## Owner Review Conclusions

1. `contest` 旧 keepalive
   - 结论：原 `contest/application/jobs/lock_keepalive.go` 维护了一套本地续租循环，和 `internal/shared/lockkeepalive` 重复，且缺少单次 refresh deadline。
   - 处理：改为复用 `lockkeepalive.Start()` 和 `lockkeepalive.RefreshInterval()`；`contest_status_updater` 与 `AWDRoundUpdater` 继续保持 Redis scheduler lock owner，只是续租策略统一到 shared owner。
   - Guardrail：`TestRedisLockKeepaliveCancelsWhenRefreshBlocksPastTTL` 覆盖 refresh 阻塞超过 TTL 时必须 fail-closed。

2. `runtime_cleaner`
   - 结论：原实现只在单轮清理开始时 acquire `ctf:container:cleanup:lock`，没有续租；`ReconcileLostActiveRuntimes / CleanExpiredInstances / CleanupOrphans` 超过 TTL 时，另一个副本可以拿到同一 cleanup lock。
   - 处理：持锁后用 `lockkeepalive.Start()` 派生 `runCtx`，清理步骤统一使用 `runCtx`；锁丢失时本轮清理停止，release 使用 `context.WithoutCancel(ctx)` 包短 timeout，避免正常 shutdown 时跳过释放。
   - Guardrail：`TestCleanerStopsRunningTaskWhenCleanupLockIsLost` 覆盖运行中 lock token 被替换后清理上下文会停止；`TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation` 覆盖 shutdown 取消 base context 后仍释放 cleanup lock。

3. `assessment_cleaner`
   - 结论：当前没有全局 cleaner owner；多副本会重复触发 `RebuildAllSkillProfiles()` 的学生列表扫描。但 `CalculateSkillProfile()` 对每个 user 使用 `AcquireFullProfileRebuildLock()`，同一用户画像写入仍是单 owner。
   - 影响：这是负载 / 噪音风险，不是当前 correctness blocker。若学生量增长或全量重建成本变高，建议后续新增全局 assessment cleaner lock；本轮不扩大配置和测试面。

## Senior Implementation Assessment

- 脚本拆分选择 `scripts/lib` 是当前最小正确落点：`tools/` 保留操作者入口，内部 helper 不进入 top-level script line guard，也不需要新增 `tools/` 子目录规则。
- `contest` 和 `runtime_cleaner` 统一复用 shared keepalive 后，lease refresh deadline、TTL 超时和 lock loss 语义不再分散在多套实现里。
- `assessment_cleaner` 没有直接代码变更是刻意选择：现有 per-user lock 已保护写入 correctness，新增全局 lock 会引入新配置和迁移面，超出本轮“复核 residual risk”的最小必要范围。

## Required Re-validation

- `bash scripts/check-script-guard.sh`
- `bash scripts/check-script-layer.sh`
- `go test ./internal/shared/lockkeepalive -count=1`
- `go test ./internal/module/contest/application/jobs -run 'RedisLockKeepalive|Lock|Scheduler' -count=1`
- `go test ./internal/module/runtime/infrastructure -run Cleaner -count=1`
- `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
- Independent reviewer also ran:
  - `bash -n tools/multi-instance-nginx-proxy-smoke.sh`
  - `bash -n scripts/lib/multi-instance-nginx-proxy-smoke/run.sh`
  - `bash scripts/check-script-guard.sh`
  - `bash scripts/check-script-layer.sh`
  - `go test ./internal/shared/lockkeepalive -count=1`
  - `go test ./internal/module/contest/application/jobs -run 'RedisLockKeepalive|Lock|Scheduler' -count=1`
  - `go test ./internal/module/runtime/infrastructure -run Cleaner -count=1`
  - `bash scripts/check-backend-architecture.sh --quick`

## Residual Risk

- 未执行真实 Docker 多实例 smoke；本轮只保证脚本入口拆分后治理检查通过。
- `assessment_cleaner` 全局单 owner 仍是后续容量优化候选，不作为本轮 blocker。
- `assessment_cleaner` 的 residual 分类假设多副本部署启用 Redis cache，使 `ProfileLockStore` 的 per-user lock 生效；如果存在 cache=nil 的部署形态，需要重新评估该 cleaner 的 correctness 风险。

## Touched Known-debt Status

- 本次触达此前 residual risk 中的 `contest` 旧 keepalive 与 `runtime_cleaner`，已在 touched surface 内收口。
- `assessment_cleaner` 已复核并分级为非 correctness blocker；没有把同面已知 correctness 债务留到后续。
