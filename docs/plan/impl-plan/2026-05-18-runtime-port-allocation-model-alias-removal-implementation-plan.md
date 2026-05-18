# runtime port allocation model alias removal 实施计划

## Objective

删除 `internal/model/port_allocation.go` 这一层兼容别名，完成 `PortAllocation` 到 runtime 模块内部的路径收敛。

## Non-goals

- 不再改动 `PortAllocation` 的表结构和业务语义
- 不再改动上一刀已经完成的 runtime owner 边界

## Inputs

- `internal/model/port_allocation.go`
- `internal/module/runtime/entity/port_allocation.go`
- `.harness/reuse-decisions/runtime-port-allocation-model-alias-removal.md`

## Ownership evaluation

- `PortAllocation` 的实体 owner 已经是 `runtime/entity`
- `internal/model/port_allocation.go` 仅剩历史兼容作用，删除后不会改变 owner，只会移除错误入口

## Task slices

1. 确认代码中已经不存在 `model.PortAllocation` 的实际引用
2. 删除 `internal/model/port_allocation.go`
3. 运行最小验证确认删除后编译和边界检查仍通过

## Validation

- `go test ./internal/module/practice/... ./internal/module/runtime/... ./internal/module/contest/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run TestModuleArchitectureBoundaries -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 是否仍有任何代码路径依赖 `model.PortAllocation`
- 删除兼容别名后，runtime entity 是否成为唯一入口

## Rollback

如有遗漏引用，可临时恢复 `internal/model/port_allocation.go` 兼容别名，再补齐剩余迁移。
