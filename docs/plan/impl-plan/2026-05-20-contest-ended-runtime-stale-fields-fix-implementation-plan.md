# contest ended runtime stale fields fix 实施计划

## Objective

修复实例进入 `stopped/expired` 终态后仍保留 `container_id / network_id / runtime_details / access_url / host_port` 的问题，避免管理员把数据库残留字段误判为仍有活跃容器。

## Non-goals

- 不在这刀处理历史脏数据批量回填
- 不在这刀处理 workspace volume 的删除
- 不重写比赛结束清理查询范围

## Inputs

- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/repository_destroyed_at_test.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go`

## Ownership evaluation

- `runtime repository` 是实例终态持久化的唯一 owner，负责把 `stopped/expired` 收口成一致的 runtime 字段清空语义。
- `instance maintenance` 负责识别过期实例并调用终态更新，不负责自行拼装字段清理逻辑。
- `contest ended runtime cleaner` 负责比赛结束时清理仍处于活跃态的实例；它依赖 repository 的终态语义，不额外复制字段清理规则。

## Task slices

1. 在 repository 回归测试中补上 `UpdateStatusAndReleasePort(..., stopped|expired)` 的 runtime 字段清空断言。
2. 修改 `UpdateStatusAndReleasePort` 终态更新逻辑，使其与 `FinalizeStoppedRuntime / ExpireInstanceRuntime` 对齐。
3. 运行 `runtime` 相关测试，确认终态更新、比赛结束清理与过期实例维护路径都保持通过。

## Data and compatibility impact

- 实例进入 `stopped/expired` 后，数据库中的运行时字段会立即清空。
- 管理端如果依赖这些字段判断容器是否存在，显示会从“历史残留值”变成空值。
- 不改变 `failed` 状态的现有诊断语义。

## Validation

- `go test ./internal/module/runtime/infrastructure -count=1`
- `go test ./internal/module/contest/infrastructure -count=1`
- `go test ./internal/module/instance/application/commands -count=1`

## Review focus

- `stopped/expired` 是否统一清空 runtime 字段与 allocation
- `failed` 状态是否仍保持现有行为
- 结束比赛清理是否继续依赖统一的 repository 终态语义，而不是复制新分支

## Rollback

如果发现终态字段清空影响现有诊断流程，可以回退 `UpdateStatusAndReleasePort` 的终态字段清理改动，仅保留测试文件并重新评估终态 owner。
