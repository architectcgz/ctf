# Reuse Decision

## Change type
entity / repository / port / service / query mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/user_score.go`
- `code/backend/internal/module/practice/...`
- `code/backend/internal/app/*integration_test.go`
- `.harness/reuse-decisions/*practice*score*`

## Similar implementations found
- `code/backend/internal/module/challenge/entity/hint.go`
- `code/backend/internal/module/challenge/entity/image_build_job.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`

## Decision
refactor_existing

## Reason
`UserScore` 只服务于 `practice` 模块的积分写入、排行读取和对应 app 测试，不满足继续留在全局 `internal/model` 共享桶的条件。最小正确方案是把 GORM 持久化实体收回 `internal/module/practice/entity`，保持表名、字段和行为不变。

非目标：本刀不处理 `Submission`、`SkillProfile`、`PortAllocation`。

## Files to modify
- `code/backend/internal/model/user_score.go`
- `code/backend/internal/module/practice/entity/user_score.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository_test.go`
- `code/backend/internal/module/practice/application/commands/score_service.go`
- `code/backend/internal/module/practice/application/commands/score_service_test.go`
- `code/backend/internal/module/practice/application/queries/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/queries/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/application/queries/score_service_test.go`
- `code/backend/internal/module/practice/testsupport/test_helper.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## After implementation
- 删除 `internal/model/user_score.go`
- 后续再单独处理 `Submission` / `SkillProfile`
