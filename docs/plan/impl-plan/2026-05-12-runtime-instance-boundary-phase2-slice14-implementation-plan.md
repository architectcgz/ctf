# Runtime / Instance 边界 Phase 2 Slice 14 Implementation Plan

## Objective

继续完成 phase 2 的剩余收口，但这刀只处理一个最小、可审查的结构面：

- 把 `runtime` 物理模块里残留的 practice-facing glue 收回 `app/composition/instance`
- 让 `runtime/runtime.Module` 不再暴露 `PracticeInstanceRepository`、`PracticeRuntimeService`
- 删除 `runtime` 对 `practice/ports` 的直接 import，收掉 `runtime -> practice` 这条代码级依赖

## Non-goals

- 不进入 phase 5，不处理 Redis / GORM concrete allowlist 收窄
- 不新建完整的 `internal/module/container_runtime` 物理模块
- 不改 `practice` 现有实例启动、重启、清理、AWD 工作区行为
- 不处理 `practice -> runtime` 或 `contest -> runtime` 的全部剩余依赖
- 不改外部 HTTP 路由、响应结构或 proxy / SSH 访问语义

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/todos/2026-05-11-runtime-container-ports-followup.md`
- `code/backend/internal/app/composition/{instance_module.go,practice_module.go,runtime_module.go}`
- `code/backend/internal/module/runtime/runtime/{module.go,adapters.go,adapters_test.go}`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Current Baseline

- `BuildPracticeModule` 已经只从 `InstanceModule` 取 `PracticeInstanceRepository` 和 `PracticeRuntimeService`
- 但 `InstanceModule` 目前仍然把这两个依赖从 `runtime.runtime.Module` 透传出来
- `runtime/runtime.Module` 为此继续 import `practice/ports`，并在 `runtime/runtime/adapters.go` 中维护 practice-specific adapter
- `code/backend/internal/module/architecture_allowlist_test.go` 仍保留 `runtime -> practice`

## Chosen Direction

1. 让 `runtime` 只保留 container-facing 与 runtime-facing 能力
   - `runtime/runtime.Module` 继续暴露 `ProvisioningRuntime`、`CleanupRuntime`、`FileRuntime`、`ManagedContainerInventory`、`InteractiveExecutor`
   - 不再继续暴露 practice-specific repository / service 视图
2. 把 practice-specific glue 回收到 `app/composition/instance`
   - `InstanceModule` 直接基于 `runtimeinfra.NewRepository(...)` 构造 `PracticeInstanceRepository`
   - `InstanceModule` 本地持有 practice runtime adapter，把 topology/container/cleanup/inspect 的转换留在 composition 边缘
3. 同步更新 guardrail 和架构事实
   - 删除 `runtime -> practice` allowlist
   - 更新 phase 2 当前事实，明确 practice-facing glue 已迁到 `InstanceModule`

## Ownership Boundary

- `runtime/runtime.Module`
  - 负责：容器运行时 capability fields 与 runtime 自身 wiring
  - 不负责：对 `practice` 暴露 repository 或 runtime service 适配面
- `app/composition/instance_module.go`
  - 负责：把 `instance` owner 与 runtime capability 组合成 `practice` 可消费的窄依赖
  - 不负责：重新成为宽 runtime owner，或回退成直接承载 HTTP proxy owner
- `practice`
  - 负责：继续只消费 `InstanceModule` 暴露的 typed deps
  - 不负责：感知底层 runtime repo / provisioning service 的具体实现

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-runtime-instance-boundary-phase2-slice14-implementation-plan.md`
- Add: `.harness/reuse-decisions/runtime-instance-boundary-phase2-slice14.md`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Add: `code/backend/internal/app/composition/instance_practice_runtime_adapter.go`
- Add: `code/backend/internal/app/composition/instance_practice_runtime_adapter_test.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 收口 composition/instance glue

目标：

- `InstanceModule` 直接构造 `PracticeInstanceRepository`
- `InstanceModule` 本地持有 practice runtime adapter
- `runtime/runtime.Module` 不再暴露 practice-facing fields

Validation:

- `cd code/backend && go test ./internal/app/composition ./internal/app -run 'Practice|Router|Instance' -count=1 -timeout 120s`

### Slice 2: 删除 runtime -> practice 物理依赖

目标：

- `runtime/runtime/{module.go,adapters.go}` 不再 import `practice/ports`
- 删除 `runtime -> practice` allowlist
- 迁移或删掉对应 adapter 测试，保持行为覆盖

Validation:

- `cd code/backend && go test ./internal/module/runtime/... ./internal/module -run 'Runtime|ModuleDependencyAllowlist' -count=1 -timeout 120s`

### Slice 3: 文档回收 phase 2 当前事实

目标：

- 设计稿和当前架构事实源都写清这刀已经完成
- 不把 `container_runtime` 物理模块未落地的后续动作误写成已完成

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Risks

- 如果只删 `runtime -> practice` allowlist，但仍让 `InstanceModule` 间接依赖 runtime 里的 practice adapter，只是把 debt 改名，不算收口
- 如果把 practice runtime adapter 直接搬进 `practice` 模块，会让 app 层 owner 再次含混，不符合当前组合边界
- 如果文档只写“移除了依赖”而不说明新 landing zone，后续容易重新把 practice glue 放回 runtime

## Verification Plan

1. `cd code/backend && go test ./internal/app/composition ./internal/app -run 'Practice|Router|Instance' -count=1 -timeout 120s`
2. `cd code/backend && go test ./internal/module/runtime/... -count=1 -timeout 120s`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 60s`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`
