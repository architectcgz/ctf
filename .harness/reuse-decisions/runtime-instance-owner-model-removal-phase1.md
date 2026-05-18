# Reuse Decision

## Change type
entity / contracts / ports / repository / application / tests / compatibility shim removal

## Existing code searched
- `code/backend/internal/model/instance.go`
- `code/backend/internal/model/awd_defense_workspace.go`
- `code/backend/internal/model/awd_service_operation.go`
- `code/backend/internal/module/instance/contracts/*`
- `code/backend/internal/module/instance/contracts/persistence.go`
- `code/backend/internal/module/instance/entity/instance.go`
- `code/backend/internal/module/instance/ports/*`
- `code/backend/internal/module/instance/application/queries/proxy_ticket_service.go`
- `code/backend/internal/module/runtime/contracts/*`
- `code/backend/internal/module/runtime/contracts/persistence.go`
- `code/backend/internal/module/runtime/entity/awd_defense_workspace.go`
- `code/backend/internal/module/runtime/entity/awd_service_operation.go`
- `code/backend/internal/module/runtime/ports/*`
- `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_service_operation_repository.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/testutil/runtimeadapters/practice_runtime_service.go`

## Similar implementations found
- `code/backend/internal/module/contest/contracts/persistence.go`
- `code/backend/internal/module/challenge/contracts/persistence.go`
- `code/backend/internal/module/runtime/entity/port_allocation.go`

## Decision
refactor_existing

## Reason
`Instance` 的真实 owner 在 `instance` 模块，`AWDDefenseWorkspace` 和
`AWDServiceOperation` 的真实 owner 在 `runtime` 模块。这三类持久化实体继续挂在
`internal/model` 会让 `instance / runtime / practice / app` 的边界一直停留在共享层。

这刀先处理持久化 owner：

1. `instance/contracts` 暴露 `ShareScope`、`Instance` 和实例状态常量
2. `runtime/contracts` 暴露 `AWDDefenseWorkspace`、`AWDServiceOperation` 及其状态常量
3. `instance / runtime / practice / app` 改走模块 contracts
4. 删除 `internal/model/instance.go`、`awd_defense_workspace.go`、`awd_service_operation.go`

非目标：
- 不处理 `ContainerConfig / ResourceLimits / SecurityConfig`
- 不处理 `InstanceRuntimeDetails / ACL / access URL helper`
- 不处理 `User / Role / Challenge / Image / Topology / AWDScopeControl`

## Files to modify
- `code/backend/internal/module/instance/contracts/*`
- `code/backend/internal/module/instance/contracts/persistence.go`
- `code/backend/internal/module/instance/entity/*`
- `code/backend/internal/module/instance/entity/instance.go`
- `code/backend/internal/module/instance/ports/*`
- `code/backend/internal/module/runtime/contracts/*`
- `code/backend/internal/module/runtime/contracts/persistence.go`
- `code/backend/internal/module/runtime/entity/*`
- `code/backend/internal/module/runtime/entity/awd_defense_workspace.go`
- `code/backend/internal/module/runtime/entity/awd_service_operation.go`
- `code/backend/internal/module/runtime/ports/*`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/instance/application/*`
- `code/backend/internal/module/instance/application/queries/proxy_ticket_service.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/app/composition/*`
- `code/backend/internal/module/contest/infrastructure/awd_service_operation_repository.go`
- `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
- `code/backend/internal/testutil/runtimeadapters/practice_runtime_service.go`
- `code/backend/internal/model/instance.go`
- `code/backend/internal/model/awd_defense_workspace.go`
- `code/backend/internal/model/awd_service_operation.go`

## After implementation
- `instance / runtime` 的持久化 owner 不再通过 `internal/model` 暴露
- `practice` 与 `app` 通过 `instance/contracts`、`runtime/contracts` 访问这三类 owner 实体
- `internal/model` 仅保留这刀未覆盖的其他模块 owner 与共享类型
