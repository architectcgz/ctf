# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/topology_service.go`
- `code/backend/internal/module/challenge/application/queries/topology_service.go`
- `code/backend/internal/module/challenge/application/commands/topology_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/topology_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`

## Decision

refactor_existing

## Reason

`challenge/application/commands/topology_service.go` 和 `challenge/application/queries/topology_service.go` 当前混合消费三类 raw lookup 的 `gorm.ErrRecordNotFound`：

- challenge lookup not-found -> `errcode.ErrChallengeNotFound`
- challenge topology / environment template not-found -> `errcode.ErrNotFound`
- package revision detail not-found -> query 层忽略并继续返回 topology

这些语义已经在 topology command/query 内部稳定存在，不需要重写 raw `Repository` 或 TemplateRepository，也不应该把 shared allowlist/docs 一起改进来。最小正确方案是：

- 在 `challenge/ports` 增加 topology/template/package-revision 相关 sentinel
- 新增窄 topology service adapter，把 raw repository / template repository / package revision repository 的 `gorm.ErrRecordNotFound` 收口成这些 sentinel
- topology command/query service 只消费 `challenge/ports` sentinel，不再直接 import GORM
- `challenge/runtime/module.go` 负责给 topology command/query 注入 adapter

## Files to modify

- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/application/commands/topology_service.go`
- `code/backend/internal/module/challenge/application/queries/topology_service.go`
- `code/backend/internal/module/challenge/application/commands/topology_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/topology_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `.harness/reuse-decisions/challenge-topology-not-found-contract-phase5-slice29.md`
- `docs/plan/impl-plan/2026-05-13-challenge-topology-not-found-contract-phase5-slice29-implementation-plan.md`

## After implementation

- `challenge/application/commands/topology_service.go -> gorm.io/gorm`
- `challenge/application/queries/topology_service.go -> gorm.io/gorm`

这两条例外应可由主线程在共享 allowlist 收口时删除。
