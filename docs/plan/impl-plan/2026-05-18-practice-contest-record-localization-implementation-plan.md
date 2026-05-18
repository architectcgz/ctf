# practice contest record 本地化实现方案

## Objective

把 `practice/ports` 里仍直接暴露的 `contest/entity` alias 改成模块内 `Record`，同时让
`practice/infrastructure` 负责 `contestentity <-> practiceports.Record` 的映射。

## Non-goals

- 不改变 `contest` 模块的 owner
- 不新建跨模块 query service
- 不移除 `practice/infrastructure` 中所有 `contest/entity` 查询
- 不处理 `practice/testsupport` 之外的其他模块测试残留

## Inputs

- `.harness/reuse-decisions/practice-contest-record-localization.md`
- `.harness/reuse-decisions/practice-contest-entity-compat-cleanup.md`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`

## Ownership Evaluation

- `contest/entity` 仍然是 `contests / contest_challenges / contest_awd_services / teams / submissions`
  的持久化 owner
- `practice/ports` 是消费侧 contract，不应继续把 owner 私有实体直接暴露给 application
- `practice/infrastructure` 是本刀允许承接 owner 实体映射的位置

## Task slices

1. 在 `practice/ports` 定义本地 `Record` 和本地常量，去掉 `contest/entity` import
2. 在 `practice/infrastructure` 增加映射，仓储返回值改为 `practiceports.Record`
3. 更新 facade wrapper 与 stub / 单测签名，收口到本地 `Record`
4. 删除 `practice/ports/ports.go -> contest/entity` 的架构 allowlist
5. 执行 `go generate`、受影响模块测试和架构检查

## Validation

- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice/ports` 是否彻底去掉了 `contest/entity` 暴露
- `practice/infrastructure` 是否只在实现层保留 `contestentity` 查询和映射
- `application`、stub、wrapper 是否统一切到本地 `Record`
- allowlist 是否只保留仍然必要的 infra 例外

## Rollback

本刀无 schema 变更。如有回归，可先把 `practice/ports` 恢复为 alias，再按仓储返回值逐个复核映射缺口。
