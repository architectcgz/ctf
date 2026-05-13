# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/commands/participation_register_commands.go`
- `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- `code/backend/internal/module/contest/application/commands/submission_validation.go`
- `code/backend/internal/module/contest/application/queries/participation_progress_query.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/team_query_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/participation.go`
- `code/backend/internal/module/contest/ports/team.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/flag_repository.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`

## Decision

refactor_existing

## Reason

`contest` 当前这组 allowlist 不是单一 `registration lookup` 泄漏，而是同一批 application use case 同时暴露了两类 concrete not-found 语义：

- `ParticipationRepository` / `SubmissionRepository` 的 registration lookup 直接返回 `gorm.ErrRecordNotFound`
- `TeamRepository.FindUserTeamInContest` 的 team finder lookup 也直接返回 `gorm.ErrRecordNotFound`

如果只给 registration 加 adapter，`participation_register_commands.go`、`submission_validation.go`、`participation_progress_query.go` 仍然需要直接 branch `gorm.ErrRecordNotFound`，对应 allowlist 无法删除。

因此这次最小正确切片是：

- 在 `contest/ports` 增加两条模块内 sentinel：
  - registration lookup not-found
  - user team in contest not-found
- 为 `participation`、`submission`、`team finder` 各加窄 adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 收口成模块内语义
- application service 只消费 sentinel，并分别映射为：
  - `ErrContestRegistrationNotFound`
  - `ErrNotRegistered`
  - `nil` / fallback
- `contest/runtime/module.go` 负责把 adapter 注入 participation / submission wiring

这仍然避开了更大的 `TeamRepository` 全量收口，不触碰 `FindByID`、`FindContestRegistration`、`CreateWithMember`、`AddMemberWithLock` 等更宽的 team surface。

## Files to modify

- `code/backend/internal/module/contest/ports/participation.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/application/commands/participation_register_commands.go`
- `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- `code/backend/internal/module/contest/application/commands/submission_validation.go`
- `code/backend/internal/module/contest/application/commands/participation_service.go`
- `code/backend/internal/module/contest/application/queries/participation_progress_query.go`
- `code/backend/internal/module/contest/application/queries/participation_service.go`
- `code/backend/internal/module/contest/application/commands/participation_service_test.go`
- `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-contest-participation-registration-not-found-contract-phase5-slice25-implementation-plan.md`

## After implementation

- `contest/application/commands/participation_register_commands.go -> gorm.io/gorm`
- `contest/application/commands/participation_review_commands.go -> gorm.io/gorm`
- `contest/application/commands/submission_validation.go -> gorm.io/gorm`
- `contest/application/queries/participation_progress_query.go -> gorm.io/gorm`

这四条例外应可一起删除。
