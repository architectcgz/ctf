# Reuse Decision

## Change type
ports / repository / infrastructure / test contract localization

## Existing code searched
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Similar implementations found
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/contest/entity`

## Decision
refactor_existing

## Reason
上一刀只是把 `practice` 对 contest-owned 持久化实体的引用，从 `internal/model` 兼容层切到了
`internal/module/contest/entity`。这比留在共享层更清楚，但 `practice/ports` 仍然直接暴露了 owner
模块的私有实体，导致 application 和测试桩也跟着认 `contest/entity` 的具体形状。

这刀继续把边界往下压：

- `practice/ports` 改成声明自己的 `Record`
- `practice/application`、stub 和 facade 只依赖这些本地 `Record`
- `practice/infrastructure` 仍可用 `contestentity` 做 ORM 查询，但返回前映射成本地 `Record`

这样保留 infra 对 owner 持久化实现的必要耦合，同时去掉上层 contract 对私有实体的外溢。

## Files to modify
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository_test.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## After implementation
- `practice/ports/ports.go` 不再 import `contest/entity`
- `practice/application` 与测试桩只看到 `practiceports.*Record`
- `practice` 对 `contest/entity` 的直接引用范围缩到 `infrastructure` 和必要的测试建模代码
