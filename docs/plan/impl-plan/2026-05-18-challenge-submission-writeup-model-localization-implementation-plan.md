# challenge submission writeup model 模块内化实现方案

## Objective

把 `internal/model/submission_writeup.go` 中的 `SubmissionWriteup` 持久化实体和状态常量收回 `internal/module/challenge/entity`，让 challenge 模块的题解投稿、教师审核、社区题解查询和 app 集成测试直接依赖模块内实体。

## Non-goals

- 不处理 `Submission`
- 不处理 `ChallengeWriteup`
- 不改变题解投稿、发布、可见性和推荐逻辑

## Inputs

- `internal/model/submission_writeup.go`
- `internal/module/challenge/...`
- `internal/app/full_router_integration_test.go`
- `internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/challenge-submission-writeup-model-localization.md`

## Ownership Evaluation

- owner 明确：`SubmissionWriteup` 的写路径和审核逻辑都由 `challenge` 模块拥有
- landing zone 明确：`internal/module/challenge/entity/submission_writeup.go`
- 影响面可控：主要在 challenge 模块，外围只涉及 app schema / integration 断言
- 结构收敛目标明确：删除旧全局模型文件并同步 mapper 生成代码

## Task slices

1. 新增 `challenge/entity/submission_writeup.go`
2. 更新 challenge ports / repository / service / tests
3. 更新 challenge mapper 声明并重新生成 goverter 代码
4. 更新 app integration schema 与状态断言引用
5. 删除 `internal/model/submission_writeup.go`

## Validation

- `go generate ./internal/module/challenge/domain`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestFullRouter_AccessControlMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `SubmissionWriteup` 是否完全收敛到 `challenge/entity`
- challenge ports / repository / mapper 是否没有残留全局依赖
- app 测试与状态常量是否都切到模块内实体

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/submission_writeup.go`。
