# Reuse Decision

## Change type
entity / repository / port / service / runtime bridge / mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/challenge_hint.go`
- `code/backend/internal/module/challenge/...`
- `code/backend/internal/app/*integration_test.go`
- `.harness/reuse-decisions/*challenge*hint*`

## Similar implementations found
- `code/backend/internal/module/challenge/entity/publish_check_job.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/domain/mappers.go`

## Decision
refactor_existing

## Reason
`ChallengeHint` 只服务于 `challenge` 模块的题目提示读写、题包导入导出和对应 app 测试，不满足继续留在全局 `internal/model` 共享桶的条件。最小正确方案不是把它塞进 `infrastructure`，而是沿用当前模块边界，把 GORM 持久化实体收回 `internal/module/challenge/entity`，保持表名、字段和行为不变。

非目标：本刀不处理 `Challenge`、`ChallengePackageRevision`、`Image`、`ImageBuildJob`。

## Files to modify
- `code/backend/internal/model/challenge_hint.go`
- `code/backend/internal/module/challenge/entity/hint.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/domain/mappers.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go`
- `code/backend/internal/module/challenge/application/queries/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/queries/response_mapper_goverter_gen.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
- `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository_test.go`
- `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- `code/backend/internal/module/challenge/runtime/import_tx_bridge.go`
- `code/backend/internal/module/challenge/runtime/package_export_tx_bridge.go`
- `code/backend/internal/module/challenge/ports/challenge_command_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/challenge_query_context_contract_test.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## After implementation
- 删除 `internal/model/challenge_hint.go`
- 后续再单独处理 `skill_profile` / `image` / `image_build_job`
