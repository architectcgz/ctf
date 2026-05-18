# practice contest entity infra 依赖清理实现方案

## Objective

清掉 `practice` 非测试基础设施层里剩余的 `contest/entity` 直接依赖，让 `practice` 继续消费
`contest` 数据，但不再把 owner 模块的私有持久化类型当作自己的 ORM row 和 snapshot owner。

## Non-goals

- 不修改 `contest` 模块实体 owner
- 不处理 `practice` 测试代码里的 `contest/entity` 建模残留
- 不迁移 `contest` 表结构或字段语义
- 不重构 `practice` 上层 application / ports 已完成的 `Record` contract

## Inputs

- `.harness/reuse-decisions/practice-contest-entity-infra-elimination.md`
- `.harness/reuse-decisions/practice-contest-record-localization.md`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/contest/entity/*.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Ownership Evaluation

- `contest` 仍然是 `contests`、`contest_challenges`、`contest_awd_services`、`teams`、`contest_registrations`
  和 `submissions` 的持久化 owner。
- `practice` 需要的是消费这些表的稳定读取 / 写入形状，而不是继续依赖 owner 模块私有 row 类型。
- `practice/infrastructure` 是这一刀承接本地 row、record 映射和 snapshot 解码的位置。

## Task slices

1. 新增 `practice/infrastructure` 本地 contest row 与 AWD service snapshot decode helper
2. 更新 `repository.go`，把 contest 查询、锁和 submission 写入改到本地 row
3. 更新 `contest_awd_runtime_subject_mapper.go`，改为消费 `practiceports.ContestAWDServiceRecord`
4. 更新 `score_repository.go`，去掉 `contest/entity` 的 submission model 依赖
5. 删除 `architecture_allowlist_test.go` 中剩余 `practice -> contest/entity` allowlist
6. 运行 `go generate`、受影响模块测试、架构测试和一致性检查

## Validation

- `go generate ./internal/module/practice/...`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice/infrastructure` 是否彻底去掉了非测试 `contest/entity` import
- 本地 row 是否保留了必要的表名、字段和 soft delete 语义
- AWD service snapshot 本地 decode 后，runtime subject 语义是否与之前一致
- architecture allowlist 是否只删除了已经不再需要的例外

## Rollback

本刀无 schema 变更。如出现回归，可临时恢复 `practice/infrastructure` 对 `contestentity` 的 row 引用，
再逐个比对本地 row 字段和 snapshot 解码差异。
