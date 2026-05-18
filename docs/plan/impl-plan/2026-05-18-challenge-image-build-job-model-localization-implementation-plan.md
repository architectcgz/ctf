# challenge image build job model 模块内化实现方案

## Objective

把 `internal/model/image_build_job.go` 中的 `ImageBuildJob` 持久化实体和构建状态常量收回 `internal/module/challenge/entity`，让 challenge 平台镜像构建链路直接依赖模块内实体。

## Non-goals

- 不处理 `Image`
- 不处理 `Challenge`、`ChallengePackageRevision`
- 不改变镜像构建接口、状态语义或表结构

## Inputs

- `internal/model/image_build_job.go`
- `internal/module/challenge/...`
- `.harness/reuse-decisions/challenge-image-build-job-model-localization.md`

## Ownership Evaluation

- owner 明确：`ImageBuildJob` 只由 `challenge` 模块的平台镜像构建链路消费
- landing zone 明确：`internal/module/challenge/entity/image_build_job.go`
- 非目标明确：不扩到 `Image`
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `challenge/entity/image_build_job.go`
2. 更新 challenge ports / repository / runtime / service / tests
3. 删除 `internal/model/image_build_job.go`

## Validation

- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- image build job 实体和状态常量是否完全收敛到 `challenge/entity`
- repository / runtime / service 是否没有残留全局依赖
- 相关测试和 schema 是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/image_build_job.go`。
