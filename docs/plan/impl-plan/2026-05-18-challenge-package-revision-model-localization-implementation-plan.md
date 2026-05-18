# challenge package revision model 模块内化实现方案

## Objective

把 `internal/model/challenge_package_revision.go` 中的 `ChallengePackageRevision` 持久化实体和 source type 常量收回 `internal/module/challenge/entity`，让 challenge 题包导入、导出、拓扑版本查询和 app 测试直接依赖模块内实体。

## Non-goals

- 不处理 `Challenge`
- 不处理 `ChallengeTopology`
- 不改变 revision 编号、导入导出流程或表结构

## Inputs

- `internal/model/challenge_package_revision.go`
- `internal/module/challenge/...`
- `internal/app/full_router_integration_test.go`
- `internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/challenge-package-revision-model-localization.md`

## Ownership Evaluation

- owner 明确：`ChallengePackageRevision` 只由 `challenge` 模块拥有
- landing zone 明确：`internal/module/challenge/entity/package_revision.go`
- 影响面可控：主要在 challenge 模块内部，外围只涉及 app integration schema
- 结构收敛目标明确：删除旧全局模型文件并同步 mapper 生成代码

## Task slices

1. 新增 `challenge/entity/package_revision.go`
2. 更新 challenge ports / runtime / repository / service / tests
3. 更新 challenge mapper 声明并重新生成 goverter 代码
4. 更新 app integration schema 与 seed 引用
5. 删除 `internal/model/challenge_package_revision.go`

## Validation

- `go generate ./internal/module/challenge/application/commands`
- `go generate ./internal/module/challenge/domain`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `ChallengePackageRevision` 是否完全收敛到 `challenge/entity`
- ports / runtime / repository 是否没有残留全局依赖
- goverter 生成代码是否和新类型路径一致

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/challenge_package_revision.go`。
