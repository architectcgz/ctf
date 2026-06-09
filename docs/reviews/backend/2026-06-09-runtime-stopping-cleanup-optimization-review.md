# Runtime Stopping Cleanup Optimization Review

- Review target:
  - Repository: `/home/azhi/workspace/projects/ctf`
  - Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-runtime-stopping-cleanup-optimization`
  - Task slug: `2026-06-09-runtime-stopping-cleanup-optimization`
  - Diff source: 当前 task worktree 未提交改动
  - Files reviewed:
    - `code/backend/internal/app/composition/instance_module.go`
    - `code/backend/internal/module/instance/application/commands/instance_service.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
    - `code/backend/internal/module/instance/contracts/events.go`
    - `code/backend/internal/module/runtime/infrastructure/repository.go`
    - `code/backend/internal/module/runtime/infrastructure/stopping_cleanup_lock_store.go`
    - `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
    - `code/backend/migrations/000014_instances_stopping_cleanup_index.up.sql`
    - `code/backend/migrations/000014_instances_stopping_cleanup_index.down.sql`
    - `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
    - `code/backend/internal/module/runtime/application/instance_service_test.go`
    - `code/backend/internal/module/runtime/infrastructure/cleaner_test.go`
    - `code/backend/internal/module/runtime/service_repository_test.go`
    - `code/backend/internal/module/runtime/service_topology_test.go`
    - `docs/plan/impl-plan/2026-06-09-runtime-stopping-cleanup-optimization-implementation-plan.md`
    - `feedback/2026-06-09-event-wakeup-keeps-durable-state-owner.md`

- Classification check:
  - Agree with `非琐碎任务`.

- Gate verdict:
  - `pass`

## Findings

### Resolved minor

1. `docs/plan/impl-plan/2026-06-09-runtime-stopping-cleanup-optimization-implementation-plan.md:58-60`
   - Severity: Minor
   - Issue: plan 里的 migration 文件名仍写成 `000002_instances_stopping_cleanup_index.*`，而实际提交物是 `000014_instances_stopping_cleanup_index.*`。
   - Why it matters: 当前代码和 migration 排序本身没有问题，但 plan 作为本次 task 的审计入口，文件名漂移会降低后续追溯和 review 复盘的准确性。
   - Resolution: 已在 plan 的 `Files` / `Task 1` 中同步为 `000014_*`。

## Material Findings

- 无。当前未发现会阻塞 completion gate 的 correctness、并发安全、架构边界或回滚可信度问题。

## Senior Implementation Assessment

- `ListStoppingInstances` 的新签名和查询形状保持在 repository owner 内，`status = stopping` + `ORDER BY updated_at, id` + `LIMIT n` 与新增复合索引一致，属于当前需求下最小且清晰的实现。
- stopping cleanup 的多节点去重没有把“锁 owner”扩散到 HTTP 请求或 runtime compat wrapper，而是收口在 `InstanceMaintenanceService` + `StoppingCleanupLockStore`；锁持有窗口覆盖了查询、分发和 `wg.Wait()`，因此不会出现“拿到同一批行但只锁了查询阶段”的空窗。
- wake-up 事件只负责向 maintenance loop 投递本地唤醒信号，handler 不直接做清理，也没有把 cleanup 完成绑定到 `DestroyInstance` 返回路径；这和架构文档里 `platform/events` 仅做进程内非关键路径联动的定位一致。
- 复用 `redislock` + `lockkeepalive` 而不是另写一套 stopping cleanup 续租逻辑，降低了锁丢失、释放和 shutdown 语义再次分叉的风险。

## Validation Evidence

- Implementation context provided:
  - `go test ./internal/module/runtime -run TestRepositoryListStoppingInstances -count=1`
  - `go test ./internal/module/runtime/application/commands -run 'TestRuntimeMaintenanceServiceRunStoppingCleanupLoop(SkipsWhenLockHeldByAnotherNode|PassesConfiguredBatchLimit|HonorsConcurrencyLimit|FinalizesStoppingInstances)|TestRuntimeMaintenanceServiceStoppingCleanupWakeupTriggersDispatchBeforeNextTick' -count=1`
  - `go test ./internal/module/runtime/application -run 'TestInstanceServiceDestroyInstance(PublishesStoppingCleanupWakeupAfterMarkStopping|SkipsWakeupWhenMarkStoppingDoesNotChangeState)' -count=1`
  - `go test ./internal/module/runtime/infrastructure -run 'TestStoppingCleanupLockStore' -count=1`
  - `go test ./internal/module/instance/...`
  - `go test ./internal/app/composition/...`
  - `go test ./internal/module/runtime/...`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Independent reviewer reran:
  - `timeout 240s go test ./internal/module/runtime -run TestRepositoryListStoppingInstances -count=1` -> pass
  - `timeout 240s go test ./internal/module/runtime/application/commands -run 'TestRuntimeMaintenanceServiceRunStoppingCleanupLoop(SkipsWhenLockHeldByAnotherNode|PassesConfiguredBatchLimit|HonorsConcurrencyLimit|FinalizesStoppingInstances)|TestRuntimeMaintenanceServiceStoppingCleanupWakeupTriggersDispatchBeforeNextTick' -count=1` -> pass
  - `timeout 240s go test ./internal/module/runtime/application -run 'TestInstanceServiceDestroyInstance(PublishesStoppingCleanupWakeupAfterMarkStopping|SkipsWakeupWhenMarkStoppingDoesNotChangeState)' -count=1` -> pass
  - `timeout 240s go test ./internal/module/runtime/infrastructure -run 'TestStoppingCleanupLockStore|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1` -> pass
  - `timeout 240s go test ./internal/app/composition -count=1` -> pass
  - `timeout 300s bash scripts/check-backend-architecture.sh --full` -> pass

## Residual Risk

- 没有 reviewer 侧的 PostgreSQL `migrate up/down` 实跑证据；`000014_instances_stopping_cleanup_index.*` 的 rollback 可信度目前基于 SQL 静态检查。由于这是单条 add/drop index migration，风险可接受，但仍不是执行级证据。
- 没有 reviewer 侧 `EXPLAIN` / `EXPLAIN ANALYZE` 证据证明 PostgreSQL 一定会选中新索引；我对“索引与查询形状匹配”的判断来自 `status` 等值过滤、`updated_at,id` 排序和 `LIMIT` 的静态分析。
- stopping cleanup 去重依赖共享 Redis 可用；当 cache client 为 `nil` 时会退回单节点/重复可容忍语义。这与现有 runtime cleaner 的 fallback 一致，本轮未扩大配置面。

## Open Questions Or Assumptions

- `service_topology_test.go:446,502` 与 `runtime/application/instance_service_test.go:104,812` 的 UTC 修正只触达测试 fixture，没有混入 production behavior 变更。我按“runtime 包稳定性修正，且符合仓库 UTC 规则”处理，未将其视为 blocker。
- `feedback/2026-06-09-event-wakeup-keeps-durable-state-owner.md` 满足 `feedback/AGENTS.md` 约定的结构与 `## 沉淀状态` 要求；我没有看到它越界记录业务待办或替代事实源。

## Touched Known-debt Status

- 本次触达的 `instance` owner、`runtime` repository 与 composition surface 没有命中当前 fact source 中仍未收口的已知 blocker debt。
- 没有发现这轮改动把 runtime cleanup owner 再次塞回 `runtime/application` compat wrapper，也没有新增 `context.Background()` / 跨模块 concrete import 这类现行架构守卫禁止的问题。
