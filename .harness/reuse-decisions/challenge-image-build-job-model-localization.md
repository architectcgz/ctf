# Reuse Decision

## Change type
entity / repository / port / service / runtime bridge / test / model localization

## Existing code searched
- `code/backend/internal/model/image_build_job.go`
- `code/backend/internal/module/challenge/...`
- `.harness/reuse-decisions/*image*build*job*`

## Similar implementations found
- `code/backend/internal/module/challenge/entity/publish_check_job.go`
- `code/backend/internal/module/challenge/entity/hint.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`

## Decision
refactor_existing

## Reason
`ImageBuildJob` 和对应状态常量只服务于 `challenge` 模块的平台镜像构建链路，不满足继续留在全局 `internal/model` 共享桶的条件。最小正确方案是把 GORM 持久化实体和强绑定状态常量收回 `internal/module/challenge/entity`，保持表名、字段和行为不变。

非目标：本刀不处理 `Image`，也不处理 `Challenge` / `ChallengePackageRevision`。

## Files to modify
- `code/backend/internal/model/image_build_job.go`
- `code/backend/internal/module/challenge/entity/image_build_job.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_build_repository.go`
- `code/backend/internal/module/challenge/runtime/import_tx_bridge.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`

## After implementation
- 删除 `internal/model/image_build_job.go`
- 后续再单独处理 `Image`
