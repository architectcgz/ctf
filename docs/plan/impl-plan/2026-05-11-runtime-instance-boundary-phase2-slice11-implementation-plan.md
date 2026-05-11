# Runtime / Instance 边界 Phase 2 Slice 11 Implementation Plan

## Objective

继续收口 `runtime` 物理模块里残留的宽运行时依赖表达：

- 删除 `internal/module/runtime/runtime.Module` 与 `Deps` 上的宽 `Engine` 入口
- 改成显式的 container runtime capability 依赖
- 让 `InstanceModule`、practice runtime adapter、AWD defense SSH gateway 只拿各自需要的能力

## Non-goals

- 不迁移 `internal/module/runtime` 到新的物理目录
- 不改 `runtime/application/*` 的业务行为
- 不删除 `runtime_adapter_compat.go`
- 不调整 challenge / contest / ops / practice 对外 contract

## Inputs

- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice7-implementation-plan.md`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/runtime/runtime/{module.go,adapters.go,adapters_test.go}`
- `code/backend/internal/app/composition/{runtime_module.go,instance_module.go}`

## Current Baseline

- `runtime/application` 已经统一依赖 `runtime/ports/container_runtime.go`
- 但 `runtime/runtime.Module` 和 `Deps` 还保留一个宽 `Engine` 聚合入口
- `InstanceModule`、practice runtime adapter 仍通过这个宽入口去拿 cleanup / file / interactive / inspect / start 等不同能力

## Chosen Direction

1. `runtime/runtime` 不再暴露 `Engine`
   - `Deps` 改成显式 capability 字段
   - `Module` 只保留外部确实要复用的 capability 字段
2. composition 在边缘层拼 capability
   - `buildRuntimeEngine(...)` 继续只在 composition 本地作为 wiring helper 存在
   - `InstanceModule` 自己把 maintenance 所需的 inspect/start 能力拼成小 adapter
3. 文档同步改成“显式 capability fields”，不再写 `Module.Engine`

## Ownership Boundary

- `runtime/ports`
  - 负责：表达 container runtime capability
  - 不负责：变回宽 engine 聚合出口
- `runtime/runtime.Module`
  - 负责：按 capability 装配 challenge / contest / ops / practice 需要的运行时能力
  - 不负责：向上暴露一个全能 `Engine`
- composition
  - 负责：按 use case 组合 cleanup / file / interactive / inspect/start 等能力
  - 不负责：把 owner 逻辑塞回 runtime 物理模块

## Change Surface

- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/01-system-architecture.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice11-implementation-plan.md`

## Task Slices

### Slice 1: runtime module capability fields

目标：

- 移除 `Module.Engine` / `Deps.Engine`
- 改成显式 capability fields 和对应内部 wiring

Validation:

- `cd code/backend && go test ./internal/module/runtime/runtime -count=1`

### Slice 2: composition 与 practice adapter 收口

目标：

- `InstanceModule` 不再通过宽 `Engine` 装配 cleanup / workbench / ssh / maintenance
- practice runtime adapter 只依赖 inspect 所需能力

Validation:

- `cd code/backend && go test ./internal/app/composition ./internal/module/runtime/runtime -count=1`

### Slice 3: 文档对齐

目标：

- 当前事实不再把 capability 聚合字段写成 `Module.Engine`

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Risks

- 如果只是把 `Engine` 改名成别的聚合类型，这轮没有真正收口
- 如果 `InstanceModule` 继续直接吃 `runtimeinfra.Engine` concrete type，会把 edge wiring 和 runtime 物理边界重新混住
- 如果文档还保留 `Module.Engine` 说法，当前事实会和代码继续漂移

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/runtime -count=1`
2. `cd code/backend && go test ./internal/app/composition ./internal/module/runtime/runtime ./internal/module/instance/... -count=1`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`
