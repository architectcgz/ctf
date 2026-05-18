# practice contest 测试实体本地化实现方案

## Objective

把 `practice` 模块测试和 testsupport 里剩余的 `contest/entity` 路径依赖切到模块内测试专用 shim，
避免 `practice` 测试继续直接引用 owner 模块私有持久化实体。

## Non-goals

- 不修改 `contest` owner 模块实体
- 不在这刀里重写所有测试夹具为 `practiceports.Record`
- 不处理 `practice` 之外其他模块的测试依赖
- 不调整业务断言或测试场景

## Inputs

- `.harness/reuse-decisions/practice-contest-test-entity-localization.md`
- `code/backend/internal/module/practice/testsupport/test_helper.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/application/commands/*_test.go`
- `code/backend/internal/module/contest/entity/*.go`

## Ownership Evaluation

- `contest/entity` 仍然是生产代码中的 owner 持久化模型
- `practice` 测试需要的是兼容表结构和 snapshot codec 的测试建模，不需要继续直接依赖 owner 路径
- `practice/testsupport/contestentity` 是这刀承接测试兼容形状的位置

## Task slices

1. 在 `practice/testsupport/contestentity` 新增测试专用本地 contest row 与 snapshot codec
2. 更新 `testsupport/test_helper.go` 与 `repository_test.go`，切到本地测试 shim
3. 更新 `application/commands` 测试和 `repository_stub_test.go` import 路径
4. 跑 `practice` 相关测试，确认 fixture 行为不变

## Validation

- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`

## Review focus

- 测试 shim 是否覆盖了当前测试实际需要的字段和 `TableName`
- snapshot encode / decode 是否与原测试期望保持一致
- import 路径是否已全部从 owner 切到 `practice/testsupport/contestentity`

## Rollback

本刀仅改测试代码。如有回归，可把测试 import 临时切回 `contest/entity`，再逐个核对 shim 字段缺口。
