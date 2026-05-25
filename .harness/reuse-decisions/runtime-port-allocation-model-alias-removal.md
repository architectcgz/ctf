# Reuse Decision

## Change type
cleanup / compatibility shim removal / runtime entity convergence

## Existing code searched
- `code/backend/internal/model/port_allocation.go`
- `code/backend/internal/module/runtime/entity/port_allocation.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `.harness/reuse-decisions/runtime-port-allocation-entity-localization.md`

## Similar implementations found
- `code/backend/internal/module/runtime/entity/port_allocation.go`
  - 上一刀已经提供了 runtime owner 实体定义。
- `code/backend/internal/module/runtime/infrastructure/repository.go`
  - 运行中代码已经不再依赖 `model.PortAllocation`。

## Decision
refactor_existing

## Reason
上一刀保留 `internal/model/port_allocation.go` 只是为了避免在未确认前直接删文件。现在用户已经明确确认删除，而且代码中已经没有 `model.PortAllocation` 的实际使用，继续保留这个兼容别名只会给后续改动留下错误暗示。

## Files to modify
- `code/backend/internal/model/port_allocation.go`
- `docs/plan/archive/impl-plan/2026-05-18-runtime-port-allocation-model-alias-removal-implementation-plan.md`

## After implementation
- `internal/model` 不再保留 `PortAllocation` 兼容别名
- `PortAllocation` 的唯一 owner 路径为 `runtime/entity`
