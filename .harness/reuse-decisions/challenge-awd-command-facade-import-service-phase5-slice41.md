# Reuse Decision

## Change type

service / composition / runtime wiring

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Similar implementations found

- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`

## Decision

refactor_existing

## Reason

这刀不处理 `AWDChallengeImportService` 内部仍依赖 `*gorm.DB` 的事务实现，只先把 facade 这层的 concrete 泄漏拿掉：`AWDChallengeCommandFacade` 改为消费预构建好的 import service，runtime 负责装配。这样可以先清掉最外层 `gorm` 依赖，不把 import tx surface 和 runtime 组装一次性混成更大 slice。

## Files to modify

- `.harness/reuse-decisions/challenge-awd-command-facade-import-service-phase5-slice41.md`
- `docs/plan/impl-plan/2026-05-14-challenge-awd-command-facade-import-service-phase5-slice41-implementation-plan.md`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## After implementation

- `challenge/application/commands/awd_challenge_command_facade.go -> gorm.io/gorm` 从 allowlist 移除
- runtime 明确承担 `AWDChallengeImportService` 的装配 owner
- `AWDChallengeImportService` 自己的 `*gorm.DB` 依赖留给后续 slice 单独处理
