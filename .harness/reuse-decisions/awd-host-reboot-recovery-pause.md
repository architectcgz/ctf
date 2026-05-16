# Reuse Decision

## Change type
service / repository / job / composition

## Existing code searched
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/composition/contest_module.go`
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`

## Similar implementations found
- `instance/application/commands/maintenance_service.go`
  - 已经收口 runtime 丢失恢复、容器重启和重新入队逻辑，适合作为宿主机恢复链路的实例 owner
- `runtime/infrastructure/repository.go`
  - 已经提供 `ListRecoverableActiveInstances`、`RequeueLostRuntime` 等运行态恢复所需数据口径，适合继续扩展 AWD 实例过期时间刷新
- `contest/application/jobs/status_update_runner.go`
  - 已经是赛事时间窗推进 owner，说明“比赛时间”不能散落到 runtime cleaner 里各自判断
- `contest/application/jobs/awd_round_scheduler_runtime.go`
  - 已经是 AWD round 推进 owner，说明恢复完成前必须阻止它按旧时间窗继续推进
- `app/composition/instance_module.go`
  - 已经集中装配 runtime cleaner 与 maintenance service，适合作为宿主机重启恢复入口

## Decision
extend_existing

## Reason
这次需求不是引入一套新的通用赛事暂停系统，而是先把“宿主机整机重启后，AWD 比赛时间应当暂停且实例恢复后再继续”收口成最小可交付链路。现有 `maintenance_service + runtime repository + contest scheduler` 已经覆盖实例恢复、比赛推进和后台任务装配边界，继续扩展这些 owner 的能力比新增并行 pause 模块更符合当前仓库结构，也能避免把前端状态枚举、比赛状态机和通用人工暂停语义一起拉进同一刀。

## Files to modify
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/instance/application/commands/` 下新增宿主机重启恢复 service
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/contest/ports/contest.go`
- `code/backend/internal/module/contest/infrastructure/contest_repository.go`
- `code/backend/internal/module/contest/application/jobs/awd_service_check_result.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/pkg/redis/keys.go`
- `code/backend/migrations/000007_add_contest_paused_seconds.up.sql`
- `code/backend/migrations/000007_add_contest_paused_seconds.down.sql`

## After implementation
- 如果“宿主机重启恢复会顺延 AWD 比赛时间”后续会反复复用，再追加到 `harness/reuse/history.md`
