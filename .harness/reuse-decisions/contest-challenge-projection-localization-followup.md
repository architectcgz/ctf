# Reuse Decision

## Change type
refactor_existing

## Existing code searched
- code/backend/internal/module
- code/backend/internal/app/composition
- code/backend/internal/model
- code/backend/internal/module/contest/ports
- code/backend/internal/module/contest/infrastructure
- code/backend/internal/module/contest/application

## Similar implementations found
- code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter.go
- code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go
- code/backend/internal/module/contest/entity/contest_awd_service.go
- code/backend/internal/module/challenge/contracts/contracts.go

## Decision
refactor_existing

## Reason
本次目标是继续收口 `contest` 模块内对 `internal/model` 的直接依赖。已有可复用路径是：保留跨模块 contract 接口不变，在 `contest` 内部通过 adapter 和本地 projection entity 做语义映射。相比新增并行仓储或跨模块直连，改造现有 `ports + infrastructure adapter + runtime composition` 的最小改动更符合当前 Onion 边界约束，也能保持现有用例行为和测试组织方式。

## Files to modify
- code/backend/internal/module/contest/entity/challenge.go
- code/backend/internal/module/contest/application/commands/awd_error_contract_test.go
- code/backend/internal/module/contest/application/commands/awd_resource_validation_support.go
- code/backend/internal/module/contest/application/commands/awd_service_test.go
- code/backend/internal/module/contest/application/commands/challenge_add_commands.go
- code/backend/internal/module/contest/application/commands/challenge_service.go
- code/backend/internal/module/contest/application/commands/contest_awd_service_service.go
- code/backend/internal/module/contest/application/commands/contest_challenge_error_contract_test.go
- code/backend/internal/module/contest/application/commands/participation_error_contract_test.go
- code/backend/internal/module/contest/application/commands/response_mappers.go
- code/backend/internal/module/contest/application/commands/submission_score_support.go
- code/backend/internal/module/contest/application/commands/submission_score_transaction.go
- code/backend/internal/module/contest/application/commands/submission_submit_validation.go
- code/backend/internal/module/contest/application/queries/challenge_service.go
- code/backend/internal/module/contest/application/queries/challenge_service_test.go
- code/backend/internal/module/contest/infrastructure/awd_command_repository.go
- code/backend/internal/module/contest/infrastructure/awd_command_repository_test.go
- code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go
- code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter.go
- code/backend/internal/module/contest/infrastructure/submission_lookup_repository.go
- code/backend/internal/module/contest/infrastructure/submission_registration_repository.go
- code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go
- code/backend/internal/module/contest/ports/awd.go
- code/backend/internal/module/contest/ports/challenge.go
- code/backend/internal/module/contest/ports/submission.go
- code/backend/internal/module/contest/runtime/module.go

## After implementation
- 已完成 `contest` 生产代码层面对 `internal/model` 的直接依赖收口，改为模块内 `contestentity.Challenge` 投影。
- 本次保持 contract 端口语义稳定，由 `contest` 侧 adapter 做 legacy challenge contract 到模块内实体的字段映射。
