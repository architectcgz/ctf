# Reuse Decision

## Change type

port / repository / service / composition

## Existing code searched

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/challenge/ports/challenge_query_context_contract_test.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge`

## Similar implementations found

- `challenge` 模块已在 `flag/writeup/topology` 使用“ports 投影 + infra 映射”收口模式。
- `challengeinfra.NewChallengeQueryRepository(...)` 已作为查询侧仓储 adapter，适合继续承接 `model -> ports` 映射。
- `runtime/module.go` 已在其他依赖注入路径使用 adapter 注入而非直接注入 raw repository。

## Decision

refactor_existing

## Reason

目标是继续消除 `challenge` 查询接口对 `internal/model` 的暴露，不新增新模块或新层级。沿用现有 adapter 模式，在 `ChallengeReadRepository / ChallengePublishedRepository` 引入投影，保留查询行为和错误语义不变。

## Files to modify

- `.harness/reuse-decisions/challenge-query-read-published-port-projections.md`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/ports/challenge_query_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/runtime/module.go`
