# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_active_support.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- `code/backend/internal/module/contest/application/commands/awd_resource_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_team_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/infrastructure/team_command_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/ports/participation.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/team_command_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 AWD command 缺少新的仓储能力，而是 `AWDService` command application 仍直接知道 `gorm.ErrRecordNotFound`。phase5 里已经验证过的最小正确模式是：

- raw repository 保持现有 GORM 语义
- `contest/ports` 只补当前 use case 需要的模块 sentinel
- `contest/infrastructure` 新增窄 adapter，把 command surface 上的 concrete not-found 翻译成模块 sentinel
- runtime 只给 command service 注入 adapter，不改 jobs / http / query 其他路径

本 slice 继续沿用这个模式，分成两条适配链：

- `AWDCommandRepository`：包住 raw `AWDRepository`，只收口 round / registration / team / challenge / contest awd service lookup，以及 attack-log transaction 回传的 not-found
- `AWDRoundManagerAdapter`：包住现有 round manager，只把 `EnsureActiveRoundMaterialized` 的 concrete not-found 收口成 contest sentinel

这样可以把这 7 条 AWDService command allowlist 收在 command-side lookup contract 内，不改 raw repository 全局语义，不碰 `challenge_add_commands.go`、`contest_awd_service_service.go`、jobs/http updater/runtime 那 8 条 surface，也不碰 teaching 脏改和 shared cleanup 文件。

## Files to modify

- `.harness/reuse-decisions/contest-awd-command-lookup-contract-phase5-slice37.md`
- `docs/plan/impl-plan/2026-05-13-contest-awd-command-lookup-contract-phase5-slice37-implementation-plan.md`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_command_repository_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_manager_adapter.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_active_support.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- `code/backend/internal/module/contest/application/commands/awd_resource_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_team_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_error_contract_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`

## After implementation

- 上述 7 个 `contest/application/commands/awd_* -> gorm.io/gorm` allowlist 对应代码面已收口到 `contest/ports + contest/infrastructure`
- AWD command path 只消费模块 sentinel，不再直接 import GORM concrete
- `contest/application/jobs/*` updater/runtime surface 保持不变，留给下一刀
- `architecture_allowlist_test.go`、`docs/architecture/backend/07-modular-monolith-refactor.md`、`docs/design/backend-module-boundary-target.md` 保持不变，交由 leader 统一收口
