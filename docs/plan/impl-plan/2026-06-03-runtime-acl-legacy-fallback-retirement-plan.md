# 2026-06-03 runtime ACL legacy fallback 退场计划

## Objective

- 在不破坏存量实例 cleanup 的前提下，移除 `runtime_cleanup_service` 对 `runtime_details.acl_rules` 的 legacy compatibility fallback。
- 先把旧实例宿主机上的平铺 ACL 规则迁成实例级 chain，再把 cleanup authority 完全收口到 `runtime_details.acl`。

## Inputs

- `docs/plan/impl-plan/2026-06-02-runtime-acl-authority-hardening-plan.md`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/runtime_node_execution_router.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/acl.go`

## Key constraint

- 旧实例不是“只缺少 handle”，而是历史上直接把 ACL 规则平铺写进 `DOCKER-USER`。
- 仅回填数据库里的 `runtime_details.acl` 不够；必须先在真实宿主机把旧规则迁成实例级 chain。

## Task slices

### Slice 1：启动时迁移 legacy ACL 资源

- 在 runtime module startup 中扫描候选实例：
  - `runtime_details.acl_rules` 非空
  - `runtime_details.acl` 为空
  - `destroyed_at IS NULL`
- 对每个候选实例：
  - 按 `node_id` 路由到正确宿主机
  - `ApplyACL(handle, rules)` 创建实例级 chain
  - 尝试 `RemoveACLRules(rules)` 删除旧平铺规则
  - 成功后回写 `runtime_details.acl`
- 对“旧规则已不存在”的幂等场景允许忽略删除错误；其它迁移失败应阻断启动。

### Slice 2：删除 cleanup fallback

- `runtime_cleanup_service` 只按 `ACLHandle` 删除 ACL 资源。
- `ManagedResources` 不再把 `ACLRules` 暴露给 cleanup authority 路径。
- 保留 `runtime_details.acl_rules` JSON 字段作为调试快照，不继续承接 cleanup。

### Slice 3：补测试与事实源

- 增加 startup migration 测试，证明候选实例会被回填 `acl` handle。
- 更新 cleanup 测试，去掉“legacy fallback 仍可 cleanup”的预期。
- 更新 `03-container-architecture.md`，删掉“旧实例兼容回退”的现状描述。

## Validation

- `cd code/backend && go test ./internal/app/composition -count=1`
- `cd code/backend && go test ./internal/module/runtime/... -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 迁移是否真的发生在实例真实宿主机，而不是默认节点。
- startup migration 是否具备幂等性，不会因为旧规则已删就卡死。
- cleanup authority 是否已经完全从 `acl_rules` 退场。
