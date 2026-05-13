# Reuse Decision

## Change type

query service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/queries/team_info_query.go`
- `code/backend/internal/module/contest/application/queries/team_list_query.go`
- `code/backend/internal/module/contest/application/queries/team_service.go`
- `code/backend/internal/module/contest/application/queries/team_service_test.go`
- `code/backend/internal/module/contest/infrastructure/team_query_repository.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/team.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- `code/backend/internal/module/challenge/infrastructure/flag_repository.go`

## Decision

refactor_existing

## Reason

当前剩余的 `contest/application/queries/team_info_query.go` 与 `contest/application/queries/team_list_query.go` 都还在直接 branch `gorm.ErrRecordNotFound`，但它们只消费 query 能力：

- `FindByID` 的 team detail not-found -> `errcode.ErrTeamNotFound`
- `FindUserTeamInContest` 的 user team lookup not-found -> `nil` fallback

这两处不需要把整组 team command / membership / registration binding 一起收口。最小正确方案是：

- 在 `contest/ports` 补一条 `team detail not found` sentinel
- 新增一个 query-only team adapter，包住现有 raw `TeamRepository`
- adapter 只翻译 `FindByID` / `FindUserTeamInContest` 的 not-found 语义，其余 query 方法 passthrough
- `contest/runtime/module.go` 只把 adapter 注给 query `TeamService`，command `TeamService` 继续保留 raw repo

这样能先删掉 team query 这两条 GORM concrete allowlist，而不把 `FindContestRegistration`、`CreateWithMember`、`AddMemberWithLock` 等更复杂的 command surface 拉进同一刀。

## Files to modify

- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/application/queries/team_info_query.go`
- `code/backend/internal/module/contest/application/queries/team_list_query.go`
- `code/backend/internal/module/contest/application/queries/team_service_test.go`
- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/team_query_adapter_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-contest-team-query-not-found-contract-phase5-slice26-implementation-plan.md`

## After implementation

- `contest/application/queries/team_info_query.go -> gorm.io/gorm`
- `contest/application/queries/team_list_query.go -> gorm.io/gorm`

这两条例外应可一起删除。
