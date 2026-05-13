# Reuse Decision

## Change type

service / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/commands/team_support.go`
- `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- `code/backend/internal/module/contest/application/commands/team_leave_commands.go`
- `code/backend/internal/module/contest/application/commands/team_captain_manage_commands.go`
- `code/backend/internal/module/contest/application/commands/team_create_retry_support.go`
- `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- `code/backend/internal/module/contest/application/commands/team_service.go`
- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/ports/participation.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`

## Decision

refactor_existing

## Reason

这次不是新建 team command 仓储或改 raw `TeamRepository` 语义，而是沿用 slice25/26 已经采用的 onion error adapter pattern：

- raw `TeamRepository` 继续允许返回 `gorm.ErrRecordNotFound`
- command surface 新增一个更窄的 adapter，专门给 `TeamService` 使用
- adapter 负责把 lookup 与 registration binding 相关的 not-found 统一翻译为 `contest/ports` sentinel
- application command 只消费 `contest/ports` / domain sentinel，不再直接依赖 GORM concrete

现有 `team_query_adapter` 已经证明这种做法能把 query surface 和 raw GORM 解耦；本 slice 只在 command wiring 侧平移这个模式，不扩展到 AWD、challenge 或共享文档。

## Files to modify

- `code/backend/internal/module/contest/application/commands/team_support.go`
- `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- `code/backend/internal/module/contest/application/commands/team_leave_commands.go`
- `code/backend/internal/module/contest/application/commands/team_captain_manage_commands.go`
- `code/backend/internal/module/contest/application/commands/team_create_retry_support.go`
- `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- `code/backend/internal/module/contest/application/commands/team_error_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/team_command_adapter.go`
- `code/backend/internal/module/contest/infrastructure/team_command_adapter_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `docs/plan/impl-plan/2026-05-13-contest-team-command-not-found-contract-phase5-slice27-implementation-plan.md`

## After implementation

主线程可删除以下 allowlist：

- `contest/application/commands/team_support.go -> gorm.io/gorm`
- `contest/application/commands/team_join_commands.go -> gorm.io/gorm`
- `contest/application/commands/team_leave_commands.go -> gorm.io/gorm`
- `contest/application/commands/team_captain_manage_commands.go -> gorm.io/gorm`
- `contest/application/commands/team_create_retry_support.go -> gorm.io/gorm`
