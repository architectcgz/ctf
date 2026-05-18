# Reuse Decision

## Change type
ports / repository / application / test / contract localization

## Existing code searched
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/model/contest_registration.go`

## Similar implementations found
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`

## Decision
refactor_existing

## Reason
`ContestRegistration` 的 owner 更接近 `contest`，但当前 `practice` 直接依赖全局 GORM 实体，只为了读取注册状态和队伍 ID。继续把 `ContestRegistration` 当跨模块共享实体，会阻塞后续 owner 收口。

最小正确方案不是现在就搬 `ContestRegistration` 持久化实体，而是先把 `practice` 依赖收成模块内只读视图，只暴露 `status` 和 `team_id` 这两个当前真实需要的字段。这样先收掉跨模块 ORM 泄漏，再为后续把实体收到 `contest/entity` 留出边界。

非目标：本刀不迁移 `ContestRegistration` 持久化实体；不改 contest 模块自己的注册写路径；不改 app schema。

## Files to modify
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`

## After implementation
- `practice` 不再通过 ports 直接暴露 `model.ContestRegistration`
- 后续可以单独把 `ContestRegistration` 收到 `contest` owner 内部
