# Reuse Decision

## Change type
entity / repository / port / service / mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/contest_announcement.go`
- `code/backend/internal/module/contest/...`
- `code/backend/internal/app/full_router_integration_test.go`
- `.harness/reuse-decisions/*contest*announcement*`

## Similar implementations found
- `code/backend/internal/module/practice/entity/user_score.go`
- `code/backend/internal/module/challenge/entity/hint.go`
- `code/backend/internal/module/contest/infrastructure/participation_announcement_repository.go`

## Decision
refactor_existing

## Reason
`ContestAnnouncement` 只服务于 `contest` 模块的报名参与和公告发布读取链路，不满足继续留在全局 `internal/model` 共享桶的条件。最小正确方案是把 GORM 持久化实体收回 `internal/module/contest/entity`，保持表名、字段和行为不变。

非目标：本刀不处理 `Contest`、`ContestRegistration`、`Submission`。

## Files to modify
- `code/backend/internal/model/contest_announcement.go`
- `code/backend/internal/module/contest/entity/announcement.go`
- `code/backend/internal/module/contest/ports/participation.go`
- `code/backend/internal/module/contest/infrastructure/participation_announcement_repository.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository_test.go`
- `code/backend/internal/module/contest/application/commands/participation_announcement_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/participation_error_contract_test.go`
- `code/backend/internal/module/contest/application/queries/participation_error_contract_test.go`
- `code/backend/internal/module/contest/testsupport/db.go`
- `code/backend/internal/app/full_router_integration_test.go`

## After implementation
- 删除 `internal/model/contest_announcement.go`
- 后续再单独处理 `contest_status_transition`
