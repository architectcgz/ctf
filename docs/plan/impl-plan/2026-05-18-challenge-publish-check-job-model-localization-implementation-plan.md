# challenge publish check job model 模块内化实现方案

## Objective

把 `internal/model/challenge_publish_check_job.go` 中的 `ChallengePublishCheckJob` 持久化实体和状态常量收回 `internal/module/challenge/entity`，让 challenge 发布自检链路和 app 测试直接依赖模块内实体。

## Non-goals

- 不处理 `ChallengeHint`
- 不处理 `ImageBuildJob`、`Image`
- 不改变发布自检接口、状态语义或表结构

## Inputs

- `internal/model/challenge_publish_check_job.go`
- `internal/module/challenge/...`
- `internal/app/full_router_integration_test.go`
- `internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/challenge-publish-check-job-model-localization.md`

## Ownership Evaluation

- owner 明确：`ChallengePublishCheckJob` 只由 `challenge` 发布自检链路消费
- landing zone 明确：`internal/module/challenge/entity/publish_check_job.go`
- 非目标明确：不扩到 hint / image 相关实体
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `challenge/entity/publish_check_job.go`
2. 更新 challenge ports / repository / service / runtime / tests / mapper
3. 更新 app 集成测试 schema 引用
4. 删除 `internal/model/challenge_publish_check_job.go`

## Validation

- `go generate ./internal/module/challenge/application/commands`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AdminChallengePublishCheckStateFlow|TestPracticeFlow_StartAndStopInstanceWithFlagSubmission' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- publish-check 实体和状态常量是否完全收敛到 `challenge/entity`
- repository / runtime bridge / mapper 是否没有残留全局依赖
- app / testsupport schema 是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/challenge_publish_check_job.go`。
