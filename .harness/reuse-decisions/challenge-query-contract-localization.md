# Reuse Decision

## Change type
contracts / ports / repository / query service / tests

## Existing code searched
- code/backend/internal/module/challenge/ports/ports.go
- code/backend/internal/module/challenge/infrastructure/repository.go
- code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go
- code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go
- code/backend/internal/module/challenge/application/queries/challenge_service.go
- code/backend/internal/module/challenge/application/queries/awd_challenge_service.go

## Similar implementations found
- code/backend/internal/module/challenge/contracts/challenge_core.go
- code/backend/internal/module/challenge/contracts/awd_challenge.go
- code/backend/internal/module/challenge/application/commands/topology_command_input.go

## Decision
refactor_existing

## Reason
challenge 模块的 handler/commands/domain 已完成 contracts 收口，query 链路仍通过全局 dto 传递查询参数，导致 contracts 边界不闭环。本次在模块内复用既有 contracts 模式，将 query 参数迁移到 `challenge/contracts`，并同步 ports/repository/query service 与测试桩。

## Files to modify
- code/backend/internal/module/challenge/contracts/awd_challenge.go
- code/backend/internal/module/challenge/ports/ports.go
- code/backend/internal/module/challenge/ports/challenge_query_context_contract_test.go
- code/backend/internal/module/challenge/ports/awd_challenge_query_context_contract_test.go
- code/backend/internal/module/challenge/infrastructure/repository.go
- code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go
- code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go
- code/backend/internal/module/challenge/infrastructure/repository_test.go
- code/backend/internal/module/challenge/infrastructure/challenge_query_repository_test.go
- code/backend/internal/module/challenge/infrastructure/awd_challenge_repository_test.go
- code/backend/internal/module/challenge/application/queries/challenge_service.go
- code/backend/internal/module/challenge/application/queries/awd_challenge_service.go
- code/backend/internal/module/challenge/application/queries/challenge_service_test.go
- code/backend/internal/module/challenge/application/queries/awd_challenge_service_context_test.go
