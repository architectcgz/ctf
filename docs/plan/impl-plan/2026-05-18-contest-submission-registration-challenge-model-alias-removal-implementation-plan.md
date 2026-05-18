# contest submission / registration / challenge model alias removal 实施计划

## Objective

删除 `internal/model/submission.go`、`internal/model/contest_registration.go`、
`internal/model/contest_challenge.go` 三层兼容 alias，并由 `contest/contracts` 提供稳定契约入口，完成这组 contest owner 实体的调用面收敛。

## Non-goals

- 不处理 `Contest`、`Team`、`ContestAWDService`、`AWDRound` 等其他 contest alias
- 不重新设计 `assessment`、`teaching_query` 的查询返回结构
- 不改动表结构、字段语义和现有 SQL 行为
- 不处理仍然应该保留在 `internal/model` 的共享 owner 类型

## Inputs

- `.harness/reuse-decisions/contest-submission-registration-challenge-model-alias-removal.md`
- `internal/model/submission.go`
- `internal/model/contest_registration.go`
- `internal/model/contest_challenge.go`
- `internal/module/contest/contracts/*`
- `internal/module/assessment/infrastructure/*`
- `internal/module/assessment/application/queries/teacher_awd_review_service_test.go`
- `internal/module/teaching_query/infrastructure/*`
- `internal/module/runtime/infrastructure/*`
- `internal/module/challenge/domain/*`
- `internal/middleware/*`
- `internal/app/*integration_test.go`

## Ownership evaluation

- `Submission`、`ContestRegistration`、`ContestChallenge` 的 owner 仍然在 `contest/entity`
- 对外暴露路径改为 `contest/contracts`，避免外部模块直接握住 owner 实体包
- `AWDChecker*` 与 registration / submission 状态常量也跟随收口到 `contest/contracts`

## Task slices

1. 在 `contest/contracts` 增加稳定契约类型和状态常量
2. 替换 `assessment`、`teaching_query`、`runtime`、`challenge`、middleware 的生产代码引用，改为依赖 `contest/contracts`
3. 替换 app 集成测试和受影响模块测试中的实体、状态常量与 AutoMigrate 入口
4. 删除 `internal/model/submission.go`、`contest_registration.go`、`contest_challenge.go`
5. 更新架构 allowlist，运行受影响模块与架构验证

## Validation

- `go generate ./internal/module/contest/contracts`
- `go test ./internal/module/contest/... -count=1`
- `go test ./internal/module/assessment/... -count=1`
- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/module/runtime/... ./internal/module/challenge/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_ContestParticipationStateMatrix|TestFullRouter_AWDTrafficSummaryAndEventsStateMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- `go test ./internal/middleware -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 是否所有 `Submission / ContestRegistration / ContestChallenge` 调用都已切到 `contest/contracts`
- 是否只删除了 alias，没有误删仍应保留在 `internal/model` 的共享语义
- 新增的 `challenge -> contest`、`teaching_query -> contest` 依赖是否仅停留在 contract 层

## Rollback

本刀无 schema 变更。如有遗漏引用，可临时恢复三份 `internal/model` alias 文件，再按调用面重新补齐。
