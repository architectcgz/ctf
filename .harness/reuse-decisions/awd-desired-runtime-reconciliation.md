# Reuse Decision

## Change type
service / repository / job / composition / config

## Existing code searched
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/composition/practice_module.go`
- `code/backend/internal/config/config.go`

## Similar implementations found
- `instance/application/commands/maintenance_service.go`
  - 已经提供 active runtime recovery，适合作为启动恢复里的第一层实例修复 owner
- `practice/application/commands/instance_start_service.go`
  - 已经拥有 AWD scope、expires_at、flag / nonce、restartable instance 复用语义，是期望态补齐的正确落点
- `practice/application/commands/instance_provisioning_scheduler.go`
  - 已经是 pending instance -> creating/running 的统一 provisioning loop，适合作为期望调和后的承接者
- `app/composition/instance_module.go` + `practice_module.go`
  - 已经拥有实例 owner 与 practice owner 的装配边界，适合做启动恢复 -> desired reconciler 的窄接口注入

## Decision
extend_existing

## Reason
这次不是要引入第二套 AWD 实例控制面，而是要把“应该活着的队伍服务”收口到现有 `practice` 编排 owner，并让启动恢复与周期 loop 都复用它。继续扩展 `maintenance_service + practice start/restart + scheduler + composition` 能保留现有实例生命周期、nonce/flag 复用和 provisioning 限流语义；新造一套 AWD runtime manager 只会复制已有 owner，并把配置、job、调度和审计做散。

## Files to modify
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/composition/practice_module.go`
- `code/backend/internal/config/config.go`

## After implementation
- 如果“startup recovery 触发 desired reconciliation + practice scheduler 周期收敛”后续在别的运行态任务中复用，再追加到 `harness/reuse/history.md`
