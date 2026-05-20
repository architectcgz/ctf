# Reuse Decision

## Change type
service / repository / port / job / state-machine / docs

## Existing code searched
- `code/backend/internal/module/instance/application/commands/instance_service.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/instance/entity/instance.go`
- `code/backend/internal/module/instance/contracts/persistence.go`
- `code/backend/internal/module/instance/ports/ports.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `docs/architecture/backend/05-key-flows.md`

## Similar implementations found
- `instance/application/commands/instance_service.go`
  - 已经是学生 / 教师主动销毁实例的唯一应用层入口，适合把删除请求收口成“只 mark stopping 并返回”。
- `instance/application/commands/maintenance_service.go`
  - 已经是运行时丢失恢复与孤儿清理的唯一后台维护 owner，适合继续承接 `stopping` cleanup loop，而不是再引入新的删除子系统。
- `runtime/application/commands/runtime_cleanup_service.go`
  - 已经拥有容器、网络、ACL、端口和子网释放链路，适合直接作为后台删除 worker 的实际 cleanup owner。
- `runtime/infrastructure/repository.go`
  - 已经拥有实例状态更新、运行时字段清空和端口 / 子网 allocation 清理能力，适合补 `MarkStopping` 与 `FinalizeStoppedRuntime`。

## Decision
extend_existing

## Reason
这次问题不只是缺少中间状态，而是删除链路仍然把 Docker cleanup 放在 HTTP 请求内执行。高并发删除时，容器 / 网络回收尾延迟直接拖慢请求，30 秒超时后还会留下 `stopping` 残留。

最小正确方案是在现有 owner 内补齐删除状态机，并把 cleanup 迁到 maintenance 的后台 loop：

- 主动删除先原子推进到 `stopping`
- 删除请求在 `stopping` 后立即返回，不再同步调用 runtime cleanup
- `instance maintenance` 后台 loop 按固定并发消费 `stopping`
- `stopping` 不再参与 active runtime recovery / requeue
- cleanup 完成后再由 repository 统一 `finalize stopped`
- 网络删除只做重试增强，不改变 cleanup owner

这样既能修掉“500 后误恢复”的根因，也能把删除并发收口到受控后台 worker，同时保持 practice scheduler、runtime cleanup 和 allocation 清理的既有边界不变。

## Files to modify
- `docs/architecture/backend/05-key-flows.md`
- `docs/plan/impl-plan/2026-05-20-runtime-instance-delete-cleanup-hardening-implementation-plan.md`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/config/config.go`
- `code/backend/configs/config.yaml`
- `code/backend/internal/module/instance/application/commands/instance_service.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/runtime/application/instance_service_test.go`
- `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
- `code/backend/internal/module/runtime/service_test.go`

## After implementation
- 如果这次 `stopping -> cleanup worker -> finalize` 链路在后续其他 runtime 资源删除场景中复用，再把这条模式沉淀到 `harness/reuse/history.md`
