# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/queries/awd_support.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- `code/backend/internal/module/contest/application/queries/awd_service.go`
- `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_team_relation_repository.go`
- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/ports/team.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`

## Decision

refactor_existing

## Reason

这次问题不是 AWD query 缺少新的读模型能力，而是 query application 直接知道 `gorm.ErrRecordNotFound`。phase5 里已经有两类可复用模式：

- team query 侧：把 concrete not-found 映射成 `contest/ports.ErrContestUserTeamNotFound`
- preview runtime lookup 侧：新增一层很窄的 contest adapter，只做 `gorm.ErrRecordNotFound -> contest sentinel`

因此 slice34 继续沿用这个模式：

- `contest/ports/awd.go` 新增 round lookup 专用 sentinel
- `contest/infrastructure` 新增 query-only AWD adapter，专门给 query `AWDService` / readiness query 注入
- adapter 只转换 `FindRoundByContestAndID`、`FindRunningRound`、`FindContestTeamByMember` 的 not-found；其他错误和其他查询方法全部透传
- `contest/application/queries` 只消费模块语义错误，不再 import `gorm.io/gorm`

这样可以把范围收在 AWD query path，不碰 AWD command/jobs，也不需要改 shared allowlist 和事实源文档。

## Files to modify

- `code/backend/internal/module/contest/application/queries/awd_support.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- `code/backend/internal/module/contest/application/queries/awd_query_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `docs/plan/impl-plan/2026-05-13-contest-awd-query-round-team-lookup-contract-phase5-slice34-implementation-plan.md`

## After implementation

- `contest/application/queries/awd_support.go` 和 `awd_workspace_query.go` 不再直接依赖 `gorm.io/gorm`
- round lookup not-found 收口到 `contest/ports` sentinel，再由 query application 决定 `errcode.ErrNotFound` 或空返回
- team lookup 缺失继续复用 `contest/ports.ErrContestUserTeamNotFound`
- 后续如果 AWD query path 还有其他 repository concrete 泄漏，优先复用这个 query-only adapter，而不是把 `gorm` 再带回 application query
