# Reuse Decision

## Change type

backend / contract localization

## Existing code searched

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/contracts/awd_challenge.go`
- `code/backend/internal/module/contest/infrastructure/*awd*lookup*`
- `code/backend/internal/module/contest/application/commands/*contract_test.go`

## Similar implementations found

- `challenge/ports.AWDChallengeQueryRepository` 已切换到 `*challengecontracts.AWDChallengeQuery`
- `challenge` 模块内 `AWDChallengeQuery` owner 已在 `challenge/contracts`

## Decision

refactor_existing

## Reason

`contest` 的 AWD challenge adapter 与测试桩仍使用 `dto.AWDChallengeQuery`，和 `challenge/ports` 当前契约不一致。先把这条 query contract 收口到 `challenge/contracts`，保持行为不变，仅消除跨模块 DTO 漂移。

## Files to modify

- `.harness/reuse-decisions/contest-awd-challenge-query-contract-localization.md`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository_test.go`
- `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter_test.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support_contract_test.go`
- `code/backend/internal/module/contest/application/commands/contest_challenge_error_contract_test.go`
