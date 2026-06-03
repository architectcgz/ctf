# Reuse Decision

## Change type
bugfix / composition / runtime routing

## Existing code searched
- code/backend/internal/app/composition
- code/backend/internal/module/runtime/infrastructure
- code/backend/internal/module/practice/application/commands
- code/backend/internal/module/contest/infrastructure

## Similar implementations found
- `code/backend/internal/app/composition/runtime_node_execution_router.go` 里的 `RemoveContainer`、`InspectManagedContainer`、`StartContainer`、`WriteFileToContainer` 已经统一通过 `clientForContainerID()` 做 `container_id -> node_id -> client` 路由。
- `code/backend/internal/module/runtime/infrastructure/repository.go` 已有 `FindRuntimeNodeIDByContainerID()`，能从 `instances.container_id`、`instances.runtime_details`、`awd_defense_workspaces.container_id join instances.node_id` 反查运行节点。
- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go` 和 `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go` 当前都会构造缺少 `NodeID` 的临时 cleanup payload，说明 authority 不能继续散落在调用方。

## Decision
extend_existing

## Reason
这次缺口不是少一处调用方补参，而是 `CleanupRuntime()` 的执行 authority 还没有和 router 内已有的容器路由 owner 对齐。

最小正确方案不是继续要求 practice / contest 各自记得补 `NodeID`，也不是在 application 层引入第二套 cleanup handle，而是沿用现有 router 模式：

- `CreateTopology` / `RunChecker` 走显式 `node_id`
- `RemoveContainer` / `WriteFileToContainer` 走 `container_id` 反查
- `CleanupRuntime` 在 `instance.NodeID` 缺失时，同样回退到容器维度反查节点

这样可以把 cleanup authority 收口到单点，并让 workspace failure cleanup 和 ended contest cleanup 共用同一条执行路径。

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-phase5-cleanup-router-authority.md
- code/backend/internal/app/composition/runtime_node_execution_router.go
- code/backend/internal/app/composition/runtime_node_execution_router_test.go
- code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner_test.go
- code/backend/internal/module/practice/application/commands/service_test.go

## After implementation
- cleanup 路由不再依赖调用方是否记得补 `NodeID`。
- runtime node authority 在 cleanup 副作用路径上与其他容器执行路径保持一致。
- practice / contest 两条 workspace cleanup 回归都能由测试直接约束。
