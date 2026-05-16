# Reuse Decision

## Change type
service / repository / port / runtime

## Existing code searched
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `docs/todos/2026-05-11-runtime-container-ports-followup.md`

## Similar implementations found
- `practice/infrastructure/repository.go`
  - 已经把端口分配拆成 `reserve -> bind` 两阶段，说明“未绑定预留端口”和“已归属实例端口”本来就是两种不同语义。
- `practice/application/commands/runtime_container_create.go`
  - 已经在 published host port 冲突重绑时区分“重试前保留新端口”和“成功后释放旧端口”，适合作为收口 owner。
- `runtime/application/commands/provisioning_service.go`
  - 已经统一承接 runtime 创建失败回滚，说明未绑定预留端口的释放应该继续落在这里，而不是扩散到更上层。
- `runtime/application/commands/runtime_cleanup_service.go`
  - 已经是实例 runtime 资源清理 owner，适合在这里根据是否存在 `instance.ID` 区分 owner-aware release 和 reserved release。

## Decision
refactor_existing

## Reason
这次不是新增一套端口管理机制，而是把现有实现里语义过宽的裸 `ReleasePort` 收紧成两种明确释放路径：`ReleaseReservedPort` 只处理已预留未绑定的端口，`ReleasePortForInstance` 只处理确属某实例的端口。继续沿用现有 `practice/runtime repository + cleanup/provisioning` 的 owner 结构做收口，改动最小，也不会把 preview / startup recovery / published host port rebind 这些已存在路径拆成新的端口子系统。

## Files to modify
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/ports/instance_context_contract_test.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/service_test.go`

## After implementation
- 如果后续要继续演进端口分配模型，优先保持“当前占用锁表”和“历史/审计记录”分离，不把这次已经收紧的 release 语义重新放宽成通用删除入口。
- 如果 runtime preview 或 startup recovery 后续再新增无 `instance.ID` 的清理入口，继续复用 `ReleaseReservedPort`，不要重新绕回 owner 不明的裸释放。
