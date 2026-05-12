# Reuse Decision

## Change type

composition / runtime / adapter / repository / port / architecture

## Existing code searched

- `code/backend/internal/app/composition/{instance_module.go,practice_module.go,runtime_module.go}`
- `code/backend/internal/module/runtime/runtime/{module.go,adapters.go,adapters_test.go}`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/instance/ports/ports.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/app/composition/challenge_module.go`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`

## Decision

refactor_existing

## Reason

当前 debt 不是“缺一套新 runtime service”，而是 `runtime` 物理模块在 phase 2 已经把实例 owner 移到 `instance` 之后，仍然残留了一层 practice-facing glue。最小合理改法不是新建第三个中间模块，而是复用现有 `InstanceModule` 作为 app 层组合落点，把 practice-specific repository / runtime adapter 收回 composition 边缘，让 `runtime` 回到 container-facing capability owner。

## Files to modify

- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/composition/instance_practice_runtime_adapter.go`
- `code/backend/internal/app/composition/instance_practice_runtime_adapter_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/runtime/adapters_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-12-runtime-instance-boundary-phase2-slice14-implementation-plan.md`

## After implementation

- 如果后续继续推进 `container_runtime` 物理模块落地，再把“container capability ports 的最终 landing zone”补进 `harness/reuse/history.md`
