# runtime / instance owner model removal phase1 实施计划

## Objective

删除 `internal/model/instance.go`、`internal/model/awd_defense_workspace.go`、
`internal/model/awd_service_operation.go` 三个 owner model 入口，并由
`instance/contracts` 与 `runtime/contracts` 提供稳定跨模块契约。

## Non-goals

- 不处理 `ContainerConfig / ResourceLimits / SecurityConfig`
- 不处理 `InstanceRuntimeDetails / InstanceRuntimeACLRule / runtime access helper`
- 不处理 `User / Role / Challenge / Image / Topology / AWDScopeControl`
- 不改数据库 schema、表结构与 SQL 语义

## Inputs

- `.harness/reuse-decisions/runtime-instance-owner-model-removal-phase1.md`
- `internal/model/instance.go`
- `internal/model/awd_defense_workspace.go`
- `internal/model/awd_service_operation.go`
- `internal/module/instance/contracts/*`
- `internal/module/instance/ports/*`
- `internal/module/runtime/contracts/*`
- `internal/module/runtime/ports/*`
- `internal/module/runtime/infrastructure/repository.go`
- `internal/module/practice/ports/ports.go`
- `internal/module/practice/infrastructure/repository.go`
- `internal/app/composition/instance_module.go`

## Ownership evaluation

- `Instance` owner：`instance`
- `AWDDefenseWorkspace` owner：`runtime`
- `AWDServiceOperation` owner：`runtime`
- `ShareScope` 作为实例共享语义，跟随 `instance/contracts` 暴露

## Task slices

1. 新增 `instance/entity` 与 `instance/contracts`：落位 `Instance`、`ShareScope`、实例状态常量
2. 新增 `runtime/contracts`：暴露 `AWDDefenseWorkspace`、`AWDServiceOperation` 及其状态常量
3. 替换 `instance / runtime / practice / app` 生产代码与受影响测试的引用
4. 删除 `internal/model/instance.go`、`awd_defense_workspace.go`、`awd_service_operation.go`
5. 运行受影响模块测试、架构边界和一致性检查

## Validation

- `go test ./internal/app ./internal/module/instance/... ./internal/module/runtime/... ./internal/module/practice/... -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `Instance / AWDDefenseWorkspace / AWDServiceOperation` 是否都改为 contracts 入口
- `practice` 与 `app` 是否仍然直接抓取 `internal/model` 的这三类 owner
- `ShareScope` 是否只作为实例共享语义存在，不再依附在共享层入口

## Rollback

本刀无 schema 变更。若有漏网引用，可临时恢复三份 `internal/model` 文件，再按编译错误补齐迁移。
