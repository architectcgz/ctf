# practice runtime port owner delegation 实施计划

## Objective

把 `practice/infrastructure/repository.go` 对 `PortAllocation` 的直接持久化读写收回 runtime owner，同时保持 `practice` 现有事务入口和应用层接口不变。

## Non-goals

- 本刀不迁移 `PortAllocation` 实体文件路径
- 不重做 `practice` / `runtime` 的事务管理结构
- 不改 published host port rebind、runtime cleanup、provisioning 的行为语义

## Inputs

- `internal/module/practice/infrastructure/repository.go`
- `internal/module/runtime/infrastructure/repository.go`
- `.harness/reuse-decisions/practice-runtime-port-owner-delegation.md`
- `.harness/reuse-decisions/runtime-port-release-semantics.md`

## Ownership evaluation

- `PortAllocation` 是 runtime 运行时资源占用记录，owner 应在 runtime repository。
- `practice` 只拥有“什么时候需要端口”这一流程决策，不拥有表结构和持久化细节。
- 最小切片是保留 `practice` 本地 port contract，对内部实现改成委托 runtime owner。

## Task slices

1. 在 runtime repository 补齐 `practice` 需要的端口 owner 能力
2. 通过 `runtime/ports + runtime/contracts` 暴露 owner 外观
3. 让 `practice` repository 委托 runtime port source，而不是自己直接 CRUD `PortAllocation`
4. 保持 `ResetInstanceRuntimeForRestart` 行为不变，并补相关测试

## Validation

- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice` 是否已经退出对 `PortAllocation` 的直接持久化依赖
- runtime owner 是否完整承接 reserve / bind / release / restart sync 语义
- 本刀是否保持在 owner delegation 范围内，没有偷渡实体迁移和大范围重构

## Rollback

如果有回归，可临时把 `practice` repository 切回原有 `PortAllocation` 直接读写实现，再单独重做 owner delegation。
