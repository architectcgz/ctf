# Reuse Decision

## Change type
ports / repository / application / test / contract localization

## Existing code searched
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_runtime_subject.go`
- `code/backend/internal/model/contest_awd_service.go`

## Similar implementations found
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_runtime_subject.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`

## Decision
refactor_existing

## Reason
`practice` 当前直接读取 `ContestAWDService` 持久化实体并在应用层解析 `ServiceSnapshot`，这让 contest 模块内部的快照存储细节直接泄漏到了 `practice` 应用层。

这刀先不迁移 `ContestAWDService` owner，也不碰 admin 列表 / 预热里仍然需要的服务实体读取；只为 `practice` 启动实例和加载运行题面新增一条模块内只读 subject contract，把真正消费的字段收成受控视图。

这样可以先把 `practice -> contest snapshot storage shape` 这条依赖切断，为后续继续收 defense workspace 配置、再讨论 owner 内收留出边界。

## Files to modify
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_runtime_subject.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`

## After implementation
- `practice` 应用层不再直接 `DecodeContestAWDServiceSnapshot`
- `practice` 通过模块内 runtime subject contract 读取 contest AWD 服务运行题面
- `ContestAWDService` 实体 owner 和 defense workspace 配置读取继续留待后续独立切片
