# Reuse Decision

## Change type
bugfix / backend runtime orchestration

## Existing code searched
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/migrations/000002_create_awd_service_operations.up.sql`

## Similar implementations found
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
  - 现有 provisioning 完成路径会按 `instance_id` 收口活跃 AWD operation，但不能处理“同 scope 已切换到新实例”的遗留记录。
- `code/backend/migrations/000002_create_awd_service_operations.up.sql`
  - 数据库已经把 AWD operation 的唯一活跃约束定义为 `(contest_id, team_id, service_id)`，说明 scope 才是正确 owner，而不是单个 instance。

## Decision
extend_existing

## Reason
当前 desired reconcile 的失败不是新的 runtime/migration 问题，而是 `CreateAWDServiceOperation` 与数据库 scope 级唯一约束不一致：历史遗留的活跃 operation 没有在新一轮 scope 重建前收口，导致后续插入直接撞上 `uk_awd_service_operations_active`。最小正确修复是在现有 repository 写入口里补 scope 级遗留活跃 operation 收口，并用现有 repository 测试面覆盖回归，而不是把清理逻辑散落到多个 command/service 调用点。

## Files to modify
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
