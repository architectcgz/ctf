# runtime instance delete cleanup hardening 实施计划

## Objective

修复实例高并发删除时的同步尾延迟与状态分叉问题：主动删除在写入 `stopping` 后立即返回，由后台限并发 worker 完成 cleanup 与 finalize，避免删除请求超时并阻止 maintenance 把删除中的实例误恢复为 `pending`。

## Non-goals

- 不重做 practice scheduler 或 runtime cleanup 的整体架构
- 不把实例网络切换成共享大网段
- 不在这刀引入 Redis / MQ 分布式删除队列

## Inputs

- `docs/architecture/backend/05-key-flows.md`
- `.harness/reuse-decisions/runtime-instance-delete-cleanup-hardening.md`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/config/config.go`
- `code/backend/configs/config.yaml`
- `code/backend/internal/module/instance/application/commands/instance_service.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`

## Ownership evaluation

- `instance service` 负责主动删除入口、权限校验和 `running|creating|pending|failed -> stopping` 的原子状态推进；不再同步 cleanup。
- `runtime cleanup` 负责容器 / 网络 / ACL / 端口 / 子网的实际清理，以及网络删除重试。
- `runtime repository` 负责 `stopping` 状态持久化、最终 `stopped` 收尾、运行时字段清空和 allocation 删除。
- `instance maintenance` 负责 active runtime recovery 和 `stopping` cleanup loop，但不负责恢复 `stopping` 实例。
- `practice scheduler` 继续只处理 `pending -> creating -> running`，不直接接手删除流程。

## Task slices

1. 保留已有 `stopping` 状态与 repository owner，不再扩展新的删除状态。
2. 改实例删除入口：只 `mark stopping` 并立即返回，不再在请求内执行 cleanup / finalize。
3. 在 maintenance 中新增常驻 `stopping` cleanup loop，按固定并发消费 `stopping` 实例并调用 `CleanupRuntime -> FinalizeStoppedRuntime`。
4. 通过 `container.delete_poll_interval` 与 `container.delete_max_concurrent` 暴露删除 worker 的轮询频率和并发上限，并注册为后台任务。
5. 保持 active runtime recovery 跳过 `stopping`，补齐异步删除与 worker 并发控制测试。

## Data and compatibility impact

- `instances.status` 新增中间状态 `stopping`
- 删除请求在运行时资源尚未释放前就会返回；用户在查询接口里看到的是 `destroying`
- cleanup 失败时，实例不再保持 `running`，而是停留在 `stopping`
- `FinalizeStoppedRuntime` 会清空 `host_port / container_id / network_id / runtime_details / access_url`
- 新增删除 worker 配置：`container.delete_poll_interval`、`container.delete_max_concurrent`

## Validation

- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/module/instance/... -count=1`
- `go test ./internal/module/practice/... -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 删除失败后实例是否稳定停留在 `stopping`，不会被 maintenance 重新入队
- 删除请求是否在 `MarkStopping` 后快速返回，不再同步持有 Docker cleanup 尾延迟
- `instance_stopping_cleanup` 是否是 `stopping` cleanup 的唯一 owner，避免和旧 stale retry 双重消费
- `stopping` 实例是否仍被视为受管 runtime，避免 orphan cleanup 抢删
- `FinalizeStoppedRuntime` 是否统一清空 runtime 字段并释放 allocation
- 网络删除重试是否只覆盖删除侧超时 / active endpoints，不引入无限等待

## Rollback

如果这刀出现回归，可以先回退 `instance_stopping_cleanup` 后台任务和删除 worker 配置，把 `DestroyInstance` 恢复成同步 cleanup + finalize，再单独重新设计删除队列 owner。
