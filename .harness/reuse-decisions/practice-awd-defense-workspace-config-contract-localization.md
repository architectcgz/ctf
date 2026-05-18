# Reuse Decision

## Change type
ports / repository / application / test / contract localization

## Existing code searched
- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/model/contest_awd_service.go`

## Similar implementations found
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_runtime_subject.go`

## Decision
refactor_existing

## Reason
上一刀已经把 `practice` 启动实例的 runtime challenge/topology 读取从应用层 snapshot decode 中收掉，但 `awd_defense_workspace_support.go` 仍直接读取 `ContestAWDService.ServiceSnapshot`、解析 `defense_workspace` 配置并在应用层计算 seed signature。

这条依赖本质上仍然是 `practice` 应用层感知 contest 模块的 snapshot 存储形状，应该继续收进现有的 `ContestAWDServiceRuntimeSubject` contract，而不是再暴露一条新的 raw snapshot 读取路径。

## Files to modify
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`

## After implementation
- `prepareAWDDefenseWorkspacePlan` 不再直接 decode `ContestAWDServiceSnapshot`
- defense workspace config、checker token env、seed signature 通过 `practice` 本地 runtime subject contract 提供
- `PortAllocation` 和 runtime tx owner 仍留待后续独立切片
