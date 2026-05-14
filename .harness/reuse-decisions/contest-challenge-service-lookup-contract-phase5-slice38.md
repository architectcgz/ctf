# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/challenge_service.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/challenge.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/challenge/contracts/contracts.go`
- `code/backend/internal/module/challenge/runtime/module.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 contest command 缺少新的业务能力，而是 `ChallengeService` 与 `ContestAWDServiceService` 仍直接知道 cross-module challenge lookup / awd challenge lookup / awd service lookup 的 concrete not-found 语义。延续 phase5 已验证模式，最小正确做法是：

- `contest/ports` 补当前 use case 需要的 challenge entity sentinel
- `contest/infrastructure` 对 cross-module challenge contract、AWD challenge query contract 做窄 adapter
- `contest/application/commands` 只消费 contest sentinel
- `contest/runtime/module.go` 只给 command 路径注入 adapter，query、jobs 和其它 contest surfaces 保持不变

其中：

- `ContestAWDServiceService.repo` 直接复用现有 `AWDCommandRepository` 的 `FindContestAWDServiceByContestAndID -> ErrContestAWDServiceNotFound`
- `challengeRepo` 与 `awdChallengeRepo` 不并入 `AWDCommandRepository`，单独落在更窄的 challenge lookup adapter，避免把 `challenge` cross-module 依赖继续堆进 AWD store adapter

## Files to modify

- `.harness/reuse-decisions/contest-challenge-service-lookup-contract-phase5-slice38.md`
- `docs/plan/impl-plan/2026-05-13-contest-challenge-service-lookup-contract-phase5-slice38-implementation-plan.md`
- `code/backend/internal/module/contest/ports/challenge.go`
- `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter.go`
- `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter_test.go`
- `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter.go`
- `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter_test.go`
- `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/contest_challenge_error_contract_test.go`
- `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`

## After implementation

- `challenge_add_commands.go -> gorm.io/gorm` 收口到 contest challenge lookup adapter
- `contest_awd_service_service.go -> gorm.io/gorm` 收口到 contest challenge lookup adapter、contest AWD challenge lookup adapter 与现有 `AWDCommandRepository`
- `contest/runtime/module.go` 只在这两个 command service 的 wiring 上引入 adapter
- `architecture_allowlist_test.go`、长期设计文档与剩余 jobs / challenge image build / updater-runtime surface 继续由 leader 在对应 slice 落地后统一收口
