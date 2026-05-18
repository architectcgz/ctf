# Reuse Decision

## Change type
entity / ports / repository / app-test / model localization

## Existing code searched
- `code/backend/internal/model/skill_profile.go`
- `code/backend/internal/module/assessment/...`
- `code/backend/internal/module/teaching_query/...`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/*assessment*skill*profile*`

## Similar implementations found
- `code/backend/internal/module/assessment/entity/report.go`
- `code/backend/internal/module/challenge/entity/submission_writeup.go`
- `code/backend/internal/module/ops/entity/audit_log.go`

## Decision
refactor_existing

## Reason
`SkillProfile` 是 assessment 模块能力画像计算与推荐链路的持久化实体，owner 明确在 `assessment`。但 `skill_profile.go` 里的维度常量、`AllDimensions` 和 `IsValidDimension` 已经被多个模块作为共享题目维度语义使用，不能一起机械搬走。

最小正确方案是只把 `SkillProfile` 持久化实体收回 `internal/module/assessment/entity`，保留共享维度常量在 `internal/model`，并同步替换 assessment、teaching_query 与 app 测试里的实体引用。

非目标：本刀不迁移 `DimensionWeb`、`AllDimensions`、`IsValidDimension` 等共享语义；不处理 `Challenge` 分类字段。

## Files to modify
- `code/backend/internal/model/skill_profile.go`
- `code/backend/internal/module/assessment/entity/skill_profile.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/domain/profile.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- `code/backend/internal/module/assessment/application/queries/profile_service_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/infrastructure/repository_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## After implementation
- `internal/model/skill_profile.go` 只保留共享维度常量和校验函数
- `SkillProfile` 实体引用全部切到 `assessment/entity`
