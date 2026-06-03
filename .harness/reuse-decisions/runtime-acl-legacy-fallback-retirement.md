# Reuse Decision

## Change type
runtime ACL authority / startup migration / legacy compatibility retirement

## Existing code searched
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/runtime_node_execution_router.go
- code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go
- code/backend/internal/module/runtime/application/commands/provisioning_service.go
- code/backend/internal/module/runtime/contracts/runtime_details.go
- code/backend/internal/module/runtime/domain/resources.go
- code/backend/internal/module/runtime/infrastructure/repository.go
- code/backend/internal/module/runtime/infrastructure/acl.go
- docs/plan/impl-plan/2026-06-02-runtime-acl-authority-hardening-plan.md

## Similar implementations found
- runtime node / startup wiring 已经集中在 `runtime_module.go`，适合承接一次性启动迁移，不需要再新造独立 job。
- node 路由与 host execution owner 已收口到 `runtimeNodeExecutionRouter`，适合复用它去按 `node_id` 执行 ACL 迁移，而不是在迁移逻辑里再拼一套节点解析。
- 现有 ACL authority 已经从新实例的 provisioning 主路径切到 `runtime_details.acl`，说明剩余问题只在 legacy 实例的资源迁移与 cleanup fallback 退场。

## Decision
extend_existing

## Reason
这次不是新增另一套 ACL 模型，而是把已经存在的“实例级 chain handle”模型补齐到存量实例，然后删掉 cleanup fallback。

最小正确落点是：

- 在 runtime startup wiring 中增加一次性 legacy ACL 迁移
- 复用现有 `runtimeNodeExecutionRouter` 按节点执行 `ApplyACL` / `RemoveACLRules`
- 在 repository 中增加针对 legacy ACL 候选实例的读取和 runtime_details 回写
- 在 cleanup service 中移除对 `acl_rules` 的 authority fallback

这样能把 owner、迁移和回退窗口都收在同一条启动链路里，而不是继续把兼容判断散落在 cleanup 调用路径。

## Files to modify
- .harness/reuse-decisions/runtime-acl-legacy-fallback-retirement.md
- docs/plan/impl-plan/2026-06-03-runtime-acl-legacy-fallback-retirement-plan.md
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/runtime_node_execution_router.go
- code/backend/internal/app/composition/runtime_module_test.go
- code/backend/internal/module/runtime/infrastructure/repository.go
- code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go
- code/backend/internal/module/runtime/domain/resources.go
- code/backend/internal/module/runtime/service_acl_test.go
- docs/architecture/backend/03-container-architecture.md

## After implementation
- 启动时会把 legacy 平铺 ACL 规则迁成实例级 chain handle，并更新 `runtime_details.acl`。
- cleanup 不再依赖 `acl_rules` compatibility fallback。
- `acl_rules` 保留为调试快照，但不再承担 runtime cleanup authority。
