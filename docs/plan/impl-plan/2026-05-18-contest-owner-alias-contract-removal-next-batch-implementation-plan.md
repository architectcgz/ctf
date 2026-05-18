# contest owner alias contract removal next batch 实施计划

## Objective

删除 `internal/model/contest.go`、`team.go`、`awd.go`、
`contest_awd_service.go`、`contest_awd_service_snapshot.go` 这些 contest owner
兼容 alias，并由 `contest/contracts` 统一承接跨模块类型、常量和 snapshot 编解码入口。

## Non-goals

- 不处理 `User`、`Challenge`、`Image`、`Instance` 等非 contest owner 类型
- 不处理 `AWDServiceOperation`、`AWDScopeControl`
- 不改动数据库 schema、表名、字段名与 SQL 语义
- 不重写已有业务逻辑，只收口类型归属和依赖入口

## Inputs

- `.harness/reuse-decisions/contest-owner-alias-contract-removal-next-batch.md`
- `internal/model/contest.go`
- `internal/model/team.go`
- `internal/model/awd.go`
- `internal/model/contest_awd_service.go`
- `internal/model/contest_awd_service_snapshot.go`
- `internal/module/contest/contracts/persistence.go`
- `internal/module/contest/entity/*.go`
- `internal/app/*integration_test.go`
- `internal/middleware/*`
- `internal/module/runtime/*`
- `internal/module/practice/*`

## Ownership evaluation

- `Contest`、`Team`、`ContestAWDService`、`ContestAWDServiceSnapshot`、`AWDRound`、
  `AWDTeamService`、`AWDAttackLog`、`AWDTrafficEvent` 的 owner 都在
  `contest/entity`
- 对外稳定入口统一收口到 `contest/contracts`
- `internal/model` 不再承担 contest owner 的兼容转发职责

## Task slices

1. 扩展 `contest/contracts`：补齐 contest owner 类型 alias、状态常量、snapshot 编解码入口
2. 替换外部模块与测试：把 app、middleware、runtime、practice、ops、assessment、teaching_query 中对上述 contest owner 的 `internal/model` 引用切到 `contest/contracts`
3. 删除 `internal/model/contest.go`、`team.go`、`awd.go`、`contest_awd_service.go`、`contest_awd_service_snapshot.go`
4. 运行生成、受影响 Go 测试、架构边界与一致性检查

## Validation

- `go generate ./internal/module/contest/contracts`
- `go test ./internal/app ./internal/middleware ./internal/module/runtime/... ./internal/module/practice/... ./internal/module/ops/... ./internal/module/assessment/... ./internal/module/teaching_query/... ./internal/module/contest/... -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 外部模块是否都通过 `contest/contracts` 访问 contest owner 类型与常量
- 是否误动了仍应保留在 `internal/model` 的非 contest owner 类型
- snapshot 编解码入口是否保持兼容

## Rollback

本刀无 schema 变更。如有漏网引用，可临时恢复对应 `internal/model` alias 文件，再按编译错误逐个补齐调用面。
