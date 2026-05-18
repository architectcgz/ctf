# practice user score model 模块内化实现方案

## Objective

把 `internal/model/user_score.go` 中的 `UserScore` 持久化实体收回 `internal/module/practice/entity`，让 practice 计分、排行查询和 app 测试直接依赖模块内实体。

## Non-goals

- 不处理 `Submission`
- 不处理 `SkillProfile`
- 不改变积分计算、排行查询或表结构

## Inputs

- `internal/model/user_score.go`
- `internal/module/practice/...`
- `internal/app/full_router_integration_test.go`
- `internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/practice-user-score-model-localization.md`

## Ownership Evaluation

- owner 明确：`UserScore` 只由 `practice` 模块消费
- landing zone 明确：`internal/module/practice/entity/user_score.go`
- 非目标明确：不扩到 `Submission`、`SkillProfile`
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `practice/entity/user_score.go`
2. 更新 practice ports / repository / service / query mapper / tests
3. 更新 app 集成测试 schema 与 seed 引用
4. 删除 `internal/model/user_score.go`

## Validation

- `go generate ./internal/module/practice/application/queries`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_StartAndStopInstanceWithFlagSubmission|TestFullRouter_RankingListsIncludeSolvedUsers' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- user score 实体是否完全收敛到 `practice/entity`
- repository / service / query mapper 是否没有残留全局依赖
- app / testsupport schema 与 seed 是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/user_score.go`。
