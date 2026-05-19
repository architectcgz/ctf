# Reuse Decision

## Change type

service / repository / port / composition

## Existing code searched

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/{flag_repository.go,topology_service_repository.go,writeup_service_repository.go}`
- `code/backend/internal/module/challenge/application/{commands,queries}/*service*.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module`

## Similar implementations found

- `challenge` 事务边界已采用投影对象模式（`ChallengePackageCore`、`ImportedChallenge`）。
- `challengeinfra.NewFlagRepository(...)` 已作为 `FlagService` 的稳定装配入口，适合复用到跨模块测试与集成测试。
- `runtime/module.go` 已存在“先 adapter 收口再注入 application”模式。

## Decision

refactor_existing

## Reason

本次目标是继续移除 `challenge` 模块 `ports` 对 `internal/model` 的暴露，不新增新层级与新组件。沿用现有 adapter + 投影收口路径，最小化改动应用层与跨模块调用点，保持错误语义与行为不变。

## Files to modify

- `.harness/reuse-decisions/challenge-flag-writeup-topology-port-projections.md`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/challenge/application/commands/flag_service.go`
- `code/backend/internal/module/challenge/application/commands/flag_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/flag_service_test.go`
- `code/backend/internal/module/challenge/application/commands/topology_service.go`
- `code/backend/internal/module/challenge/application/commands/topology_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/flag_service.go`
- `code/backend/internal/module/challenge/application/queries/flag_service_test.go`
- `code/backend/internal/module/challenge/application/queries/topology_service_test.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service_test.go`
- `code/backend/internal/module/challenge/contracts/persistence.go`
- `code/backend/internal/module/challenge/infrastructure/flag_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/ports/challenge_topology_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/challenge_writeup_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/flag_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/submission_service_test.go`
