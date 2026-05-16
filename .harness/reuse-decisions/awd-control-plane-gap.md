# Reuse Decision

## Change type
service / handler / repository / port / api / migration / schema

## Existing code searched
- `code/backend/internal/module/practice/application/commands`
- `code/backend/internal/module/practice/infrastructure`
- `code/backend/internal/module/runtime/infrastructure`
- `code/backend/internal/module/instance/application`
- `code/backend/internal/module/contest/application/commands`
- `code/backend/internal/model`
- `code/backend/migrations`

## Similar implementations found
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - 现有 automatic desired reconcile owner，可复用 team/service scope 推导与 desired state 清理逻辑
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
  - 现有 AWD scope 解析入口，可扩展显式启停 gate
- `code/backend/internal/module/runtime/infrastructure/{repository.go,awd_target_proxy_repository.go}`
  - 现有 runtime access / proxy scope 查询 owner，可在原查询里加 control join
- `code/backend/internal/model/awd.go`
  - 已有 AWD 相关持久化模型风格可参考，但 `AWDTeamService` 是按 round 的 service check 记录，不适合承载 runtime control
- `code/backend/internal/module/contest/application/commands/participation_{register,review}_commands.go`
  - 证明报名状态 owner 已存在，但它当前不能覆盖 runtime team-member 访问面，因此不适合作为本次 AWD runtime control owner

## Decision
extend_existing

## Reason
本次优先复用已有 AWD scope 解析、desired reconcile 和 runtime query owner；只在缺少持久化载体时新增一张专用 `awd_scope_controls` 表。没有复用 `contest_registrations`，因为那条 owner 无法直接收口 team-member 驱动的 runtime 访问与代理入口。

## Files to modify
- `code/backend/internal/module/practice/application/commands/*`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/*`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/model/*`
- `code/backend/migrations/*`
- `code/backend/migrations/000008_create_awd_scope_controls.up.sql`
- `code/backend/migrations/000008_create_awd_scope_controls.down.sql`

## After implementation
- 如果 `awd_scope_controls` 的 join / gate 方式后续仍会复用到更多 AWD 路径，再补 `harness/reuse/history.md`
