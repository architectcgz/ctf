# challenge hint model 模块内化实现方案

## Objective

把 `internal/model/challenge_hint.go` 中的 `ChallengeHint` 持久化实体收回 `internal/module/challenge/entity`，让 challenge 提示读写、导入导出链路和 app 测试直接依赖模块内实体。

## Non-goals

- 不处理 `Challenge`
- 不处理 `ChallengePackageRevision`
- 不处理 `Image`、`ImageBuildJob`
- 不改变题目提示接口、排序语义或表结构

## Inputs

- `internal/model/challenge_hint.go`
- `internal/module/challenge/...`
- `internal/app/full_router_integration_test.go`
- `internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/challenge-hint-model-localization.md`

## Ownership Evaluation

- owner 明确：`ChallengeHint` 只由 `challenge` 模块消费
- landing zone 明确：`internal/module/challenge/entity/hint.go`
- 非目标明确：不扩到 challenge 核心聚合和 image 相关实体
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `challenge/entity/hint.go`
2. 更新 challenge ports / domain / repository / runtime / tests
3. 更新 app 集成测试 schema 与 seed 引用
4. 删除 `internal/model/challenge_hint.go`

## Validation

- `go generate ./internal/module/challenge/domain ./internal/module/challenge/application/queries`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- hint 实体是否完全收敛到 `challenge/entity`
- domain / repository / runtime bridge / mapper 是否没有残留全局依赖
- app / testsupport schema 与导入导出链路是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/challenge_hint.go`。
