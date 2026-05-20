# Reuse Decision

## Change type

service / observability / test

## Existing code searched

- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/runtime/infrastructure/cleaner.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `/tmp/ctf-backend.log`

## Similar implementations found

- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/cleaner.go`
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`

## Decision

extend_existing

## Reason

这次问题不是缺一套新的 provisioning service，也不是要把 Docker engine 层再包一层 telemetry adapter。根因已经明确落在现有 `ProvisioningService.CreateTopology()` 的阶段内，只是缺少可以直接判责的观测字段。

最小正确方案是：

- 继续复用现有 `ProvisioningService` 作为拓扑创建的唯一 owner
- 在现有创建顺序上补阶段耗时日志，而不是改调用方超时逻辑或新增旁路采集器
- 复用已有 `zap.Duration("duration", ...)` 字段风格和现有 `runtime/service_test.go` 的 fake engine
- 保持错误传播与 cleanup 语义不变，让日志只承担观测，不承担控制

这样能最快把 50/100 并发失败从“总超时”拆解成“网络创建 / 镜像探测 / 容器创建 / 容器启动”四类可操作信号。

## Files to modify

- `.harness/reuse-decisions/runtime-provisioning-stage-timing.md`
- `docs/plan/impl-plan/2026-05-20-runtime-provisioning-stage-timing-implementation-plan.md`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/service_test.go`

## After implementation

- 创建链路失败时，日志能直接标出具体阶段与耗时
- 压测失败归因不再只剩 `context deadline exceeded`
- 如果后续还要改超时预算或并发控制，可以直接基于这些阶段日志做定量调整，而不是再盲调线程池
