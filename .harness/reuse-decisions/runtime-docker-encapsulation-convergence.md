# Reuse Decision

## Change type

infrastructure / application wiring / tests / docs

## Existing code searched

- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/runtime_metrics.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/runtime/ports/errors.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/runtime/infrastructure/engine_error_test.go`

## Similar implementations found

- `runtime/ports/errors.go`
  - 已经存在 `ErrPublishedHostPortConflict` 这一层 typed runtime error，说明底层 Docker 语义应继续在 runtime port 层收口，而不是让应用层继续解析 daemon message
- `runtime/infrastructure/runtime_metrics.go`
  - 运行时适配器已经按 stats 能力拆到独立文件，`engine.go` 继续按能力拆分可以复用同样的包内组织方式，不需要新建新的 adapter package
- `runtime/runtime/module.go`
  - runtime module 已经统一构建 `ProvisioningService` / `RuntimeCleanupService`，说明实例组合层重复 new 同一批服务属于应当回收的装配漂移，而不是新需求

## Decision

extend_existing

## Reason

这次不是新增一套 runtime provider，也不是替换 Docker SDK，而是把已经正确建立起来的 runtime 分层继续收口到更稳定的形态。最小正确改法不是再加一层 façade，而是：

- 继续沿用 `runtime/ports` 作为 Docker 语义的唯一 typed error owner
- 在 `runtime/infrastructure` 内部拆分 `Engine` 文件职责，而不改变对外接口
- 让 `runtime.Module` 真正成为 `ProvisioningService` / `RuntimeCleanupService` 的 owner，避免 composition 重复装配
- 去掉 adapter 对调用方 `ContainerConfig` 的原地修改，保持输入对象无副作用

这样可以在不扩散改动面的前提下，把 review 里指出的分层和维护性问题一起收掉。

## Files to modify

- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/engine_*.go`
- `code/backend/internal/module/runtime/infrastructure/engine_test.go`
- `code/backend/internal/module/runtime/infrastructure/engine_error_test.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/runtime/adapters_test.go`
- `code/backend/internal/module/runtime/runtime/module_test.go`
- `code/backend/internal/module/runtime/ports/errors.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/app/composition/instance_module.go`
- `docs/plan/impl-plan/2026-05-16-runtime-docker-encapsulation-convergence-implementation-plan.md`

## After implementation

- 如果后续真的引入第二种 runtime provider，再重新评估是否需要把 `Engine` 继续拆成更细的 concrete adapter struct，而不是只保持多文件组织
