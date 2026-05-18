# runtime port allocation entity localization 实施计划

## Objective

把 `PortAllocation` 实体定义迁到 `runtime/entity`，并把当前运行中的引用切到 runtime owner 路径。

## Non-goals

- 本刀不删除 `internal/model/port_allocation.go`
- 不重构 `PortAllocation` 表结构
- 不改变端口预留、绑定、释放的业务语义

## Inputs

- `internal/model/port_allocation.go`
- `internal/module/runtime/infrastructure/repository.go`
- `.harness/reuse-decisions/runtime-port-allocation-entity-localization.md`
- `.harness/reuse-decisions/practice-runtime-port-owner-delegation.md`

## Ownership evaluation

- `PortAllocation` 已经由 runtime repository 拥有持久化语义，实体文件继续留在 `internal/model` 只会保留错误暗示。
- 由于删除文件需要单独确认，这刀先做“owner 路径切换 + 兼容别名”。

## Task slices

1. 新增 `runtime/entity.PortAllocation`
2. 把 runtime 生产代码和相关测试迁到新路径
3. 把 app / practice / contest 测试里的迁移和断言同步到新路径
4. 在 `internal/model` 保留兼容别名，避免一次性破坏删除

## Validation

- `go test ./internal/module/practice/... ./internal/module/runtime/... -count=1`
- `go test ./internal/module/contest/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run TestModuleArchitectureBoundaries -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 生产代码是否已经不再直接引用 `model.PortAllocation`
- 新旧路径是否只剩兼容别名，不留下第二份 owner
- 这刀是否保持在实体路径本地化范围，没有夹带删除和额外行为改动

## Rollback

如有回归，可暂时把引用切回 `model.PortAllocation`，保留 `runtime/entity` 文件待下一轮继续收口。
