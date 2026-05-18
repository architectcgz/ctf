# assessment skill profile model 模块内化实现方案

## Objective

把 `internal/model/skill_profile.go` 中的 `SkillProfile` 持久化实体收回 `internal/module/assessment/entity`，让 assessment 模块能力画像计算、推荐查询、teaching_query 测试和 app 集成测试直接依赖 assessment 模块内实体。

## Non-goals

- 不迁移 `DimensionWeb`、`AllDimensions`、`IsValidDimension`
- 不改变维度评分、推荐排序或画像更新逻辑
- 不处理 `Challenge` 的分类字段归属

## Inputs

- `internal/model/skill_profile.go`
- `internal/module/assessment/...`
- `internal/module/teaching_query/infrastructure/repository_test.go`
- `internal/app/full_router_integration_test.go`
- `internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/assessment-skill-profile-model-localization.md`

## Ownership Evaluation

- owner 明确：`SkillProfile` 的写路径和推荐消费 owner 在 `assessment`
- 共享语义已识别：维度常量与校验函数属于跨模块共享语义，继续留在 `internal/model`
- landing zone 明确：`internal/module/assessment/entity/skill_profile.go`
- 影响面可控：assessment 为主，外围只涉及 teaching_query 测试和 app migration schema

## Task slices

1. 新增 `assessment/entity/skill_profile.go`
2. 更新 assessment ports / repository / domain / services / tests
3. 更新 teaching_query 与 app 测试中的 `SkillProfile` 引用
4. 裁剪 `internal/model/skill_profile.go`，只保留共享维度语义

## Validation

- `go test ./internal/module/assessment/... -count=1`
- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `SkillProfile` 是否完全收口到 `assessment/entity`
- 共享维度常量是否仍留在稳定共享层
- teaching_query 是否只消费共享维度语义，不反向依赖 assessment 私有实现

## Rollback

本刀无 schema 变更，如有回归可把 `SkillProfile` 临时放回 `internal/model`，维度常量无需回滚。
