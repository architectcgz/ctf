# Reuse Decision

## Change type

port / repository adapter / command service / runtime wiring

## Existing code searched

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge`

## Similar implementations found

- `challenge_query_repository.go` 已使用 `ports 投影 + infrastructure 映射` 收口查询侧 `model` 暴露。
- `flag/writeup/topology` 相关仓储 adapter 已用“raw repository + 端口投影”的方式隔离 `internal/model`。
- `challenge` 命令服务已按仓储接口注入，适合只在仓储边界补映射，不改业务行为。

## Decision

refactor_existing

## Reason

目标是收口 `ChallengeWriteRepository` 对 `*model.Challenge` 的直接暴露，保持现有事务/错误语义和业务逻辑不变。沿用模块内既有 adapter 模式，在 ports 定义命令侧投影，在 infrastructure 做双向映射，在 command service 做最小转换衔接。

## Files to modify

- `.harness/reuse-decisions/challenge-command-write-port-projection.md`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- `code/backend/internal/module/challenge/ports/challenge_command_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/runtime/module.go`
