# Reuse Decision

## Change type
entity / ports / repository / service / testsupport / compatibility shim

## Existing code searched
- `code/backend/internal/model/contest.go`
- `code/backend/internal/model/contest_registration.go`
- `code/backend/internal/model/team.go`
- `code/backend/internal/model/submission.go`
- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/model/contest_awd_service.go`
- `code/backend/internal/model/contest_awd_service_snapshot.go`
- `code/backend/internal/model/awd.go`
- `code/backend/internal/module/contest/...`
- `code/backend/internal/module/contest/application/commands/awd_service_upsert_transaction.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_native_test.go`
- `code/backend/internal/module/contest/application/commands/response_mappers_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_service_check_empty_result.go`
- `code/backend/internal/module/contest/application/jobs/awd_service_check_outcome_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_service_check_probe_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_service_definition_support.go`
- `code/backend/internal/module/contest/application/jobs/status_transition_service_test.go`
- `code/backend/internal/module/contest/application/queries/awd_summary_service_support.go`
- `code/backend/internal/module/contest/application/queries/challenge_service_test.go`
- `code/backend/internal/module/contest/application/queries/contest_awd_service_query_test.go`
- `code/backend/internal/module/contest/application/queries/contest_service_test.go`
- `code/backend/internal/module/contest/application/queries/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/queries/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/domain/awd_service_config.go`
- `code/backend/internal/module/contest/entity/contest_awd_service.go`
- `code/backend/internal/module/contest/entity/contest_awd_service_snapshot.go`
- `code/backend/internal/module/contest/infrastructure/awd_team_service_repository.go`

## Similar implementations found
- `code/backend/internal/module/contest/entity/announcement.go`
- `code/backend/internal/module/contest/entity/status_transition.go`
- `code/backend/internal/module/runtime/entity/port_allocation.go`

## Decision
refactor_existing

## Reason
`Contest`、`ContestRegistration`、`Team`、`Submission`、`ContestChallenge`、`ContestAWDService` 以及 contest AWD 轮次与流量相关实体都属于 `contest` 模块拥有的持久化模型，不适合继续把真实 owner 留在全局 `internal/model`。

本刀采用两段式收口：

1. 先把真实定义迁入 `internal/module/contest/entity`
2. `internal/model` 暂时降级为兼容 alias，避免一次把 `practice/runtime/assessment/app` 的外部引用全部打散

这样可以先完成 owner 收回，再继续分刀清理外部模块残留依赖。

非目标：
- 本刀不处理 `AWDScopeControl`
- 本刀不处理 `AWDServiceOperation`
- 本刀不删除 `internal/model` 下对应兼容文件

## Files to modify
- `code/backend/internal/module/contest/entity/*.go`
- `code/backend/internal/model/contest.go`
- `code/backend/internal/model/contest_registration.go`
- `code/backend/internal/model/team.go`
- `code/backend/internal/model/submission.go`
- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/model/contest_awd_service.go`
- `code/backend/internal/model/contest_awd_service_snapshot.go`
- `code/backend/internal/model/awd.go`
- `code/backend/internal/module/contest/...`

## After implementation
- contest 自有持久化实体真实 owner 迁到 `internal/module/contest/entity`
- `contest` 模块内部不再依赖这些 `internal/model` 定义
- 外部模块仍可通过兼容 alias 继续编译，后续再分刀清理
