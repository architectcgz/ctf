# Reuse Decision

## Change type
entity / contracts / ports / app / tests / compatibility shim removal

## Existing code searched
- `code/backend/internal/model/contest.go`
- `code/backend/internal/model/team.go`
- `code/backend/internal/model/awd.go`
- `code/backend/internal/model/contest_awd_service.go`
- `code/backend/internal/model/contest_awd_service_snapshot.go`
- `code/backend/internal/module/contest/contracts/persistence.go`
- `code/backend/internal/module/contest/entity/*.go`
- `code/backend/internal/app/*integration_test.go`
- `code/backend/internal/module/runtime/...`
- `code/backend/internal/module/practice/...`
- `code/backend/internal/module/ops/...`
- `code/backend/internal/middleware/...`

## Similar implementations found
- `code/backend/internal/module/contest/contracts/persistence.go`
- `code/backend/internal/module/challenge/contracts/persistence.go`
- `.harness/reuse-decisions/contest-submission-registration-challenge-model-alias-removal.md`

## Decision
refactor_existing

## Reason
`Contest`、`Team`、`AWDRound`、`AWDTeamService`、`AWDAttackLog`、`AWDTrafficEvent`、
`ContestAWDService`、`ContestAWDServiceSnapshot` 及其相关状态常量都属于 `contest`
模块 owner，不应继续通过全局 `internal/model` 暴露。

上一刀已经删除了 `Submission / ContestRegistration / ContestChallenge / AWDChallenge`
的开放 alias。本刀延续同一模式：

1. 由 `contest/contracts` 提供稳定跨模块入口
2. 外部模块与测试切到 `contest/contracts`
3. 删除 `internal/model` 中对应 contest owner alias 文件

非目标：
- 不处理 `User`、`Challenge`、`Image`、`Instance` 等其他模块 owner
- 不处理 `AWDServiceOperation`、`AWDScopeControl`
- 不改动表结构、字段语义和现有 SQL 行为

## Files to modify
- `code/backend/internal/module/contest/contracts/persistence.go`
- `code/backend/internal/model/contest.go`
- `code/backend/internal/model/team.go`
- `code/backend/internal/model/awd.go`
- `code/backend/internal/model/contest_awd_service.go`
- `code/backend/internal/model/contest_awd_service_snapshot.go`
- `code/backend/internal/app/...`
- `code/backend/internal/middleware/...`
- `code/backend/internal/module/runtime/...`
- `code/backend/internal/module/practice/...`
- `code/backend/internal/module/ops/...`
- `code/backend/internal/module/assessment/...`
- `code/backend/internal/module/teaching_query/...`

## After implementation
- `contest` owner 实体与状态常量的跨模块入口统一落在 `contest/contracts`
- app、runtime、practice、middleware 以及相关测试不再通过 `internal/model`
  访问这些 contest owner 类型
- `internal/model` 删除上述 contest owner alias 文件
