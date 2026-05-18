# Reuse Decision

## Change type
testsupport / test fixture / compatibility shim / test contract localization

## Existing code searched
- `code/backend/internal/module/practice/testsupport/test_helper.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/application/commands/*.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/contest/entity/*.go`
- `.harness/reuse-decisions/practice-contest-entity-infra-elimination.md`

## Similar implementations found
- `practice/infrastructure/contest_persistence_rows.go`
  - 非测试代码已经把 owner 持久化依赖收到了模块内私有 row。
- `practice/testsupport/test_helper.go`
  - 已经是测试装配入口，适合承接模块内测试专用持久化 shim。
- `contest/entity/*.go`
  - 当前测试大量直接拿 owner 模块实体当 fixture 结构和 snapshot codec 使用。

## Decision
create_new_with_reason

## Reason
`practice` 测试里对 `contest/entity` 的依赖主要有三类：

- `AutoMigrate` / seed 时需要表结构兼容的 row
- `ContestAWDServiceSnapshot` 的 encode / decode helper
- stub 中把 contest row 转成本地 `practiceports.Record`

如果逐个测试直接改成散落的本地 struct，会引入大量重复 fixture 代码，也会把测试建模切得过碎。
更合适的是在 `practice/testsupport` 下补一个测试专用本地 `contestentity` shim：

- 路径属于 `practice` 自己
- 对外保留现有测试最常用的类型名和 helper 形状
- 内部只复制本模块测试实际需要的最小 owner 结构

这样能先把 `practice` 测试对 owner 模块的路径依赖切断，后续再按需要继续收细。

## Files to modify
- `code/backend/internal/module/practice/testsupport/contestentity/models.go`
- `code/backend/internal/module/practice/testsupport/test_helper.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/application/commands/*_test.go`
- `docs/plan/impl-plan/2026-05-18-practice-contest-test-entity-localization-implementation-plan.md`

## After implementation
- `practice` 测试与 testsupport 不再 import `ctf-platform/internal/module/contest/entity`
- 测试继续沿用熟悉的 fixture 形状，但 owner 路径改到 `practice/testsupport/contestentity`
- 非测试代码与测试代码的本地化边界保持一致
