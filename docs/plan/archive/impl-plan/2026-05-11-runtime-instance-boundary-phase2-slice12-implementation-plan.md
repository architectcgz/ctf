# Runtime / Instance 边界 Phase 2 Slice 12 Implementation Plan

## Objective

继续收口 app composition 里残留的 compat 痕迹：

- 删除 `internal/app/composition/runtime_adapter_compat.go` 这个 compat 文件本体
- 保留仍然需要的 runtime HTTP facade，但改成普通 composition adapter
- 去掉 `runtime/api/http` 中已经没有路由入口的 AWD defense workbench 死接口

## Non-goals

- 不恢复 `defense/files`、`defense/directories`、`defense/commands` 路由
- 不改 `runtime/api/http` 现有实例访问、代理和 AWD defense SSH 的外部契约
- 不继续搬动 `runtime/application/*` 或 `runtime/ports/container_runtime.go` 的物理目录
- 不新增新的跨模块 landing zone

## Inputs

- `docs/todos/2026-05-11-runtime-container-ports-followup.md`
- `docs/architecture/backend/{01-system-architecture.md,03-container-architecture.md,07-modular-monolith-refactor.md}`
- `docs/architecture/features/AWD防守工作区与边界设计.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice10-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice11-implementation-plan.md`
- `code/backend/internal/app/composition/{instance_module.go,runtime_adapter_compat.go,runtime_module_test.go}`
- `code/backend/internal/module/runtime/api/http/{handler.go,handler_test.go}`

## Current Baseline

- `runtime/application/*` 中的 instance / proxy ticket / maintenance compat import path 已经删除
- 生产使用的 runtime HTTP facade 仍保留在 `internal/app/composition/runtime_adapter_compat.go`
- 该 facade 里还挂着 AWD defense workbench 的转发方法，但当前路由表已经明确没有 `defense/files`、`defense/directories`、`defense/commands` 入口
- `InstanceModule` 因此仍然在给一个不会被路由调用的 facade 注入 `AWDDefenseWorkbenchService`

## Chosen Direction

1. 删除 compat 文件名，但不删除活跃 adapter
   - 新增普通命名的 composition adapter 文件
   - `runtime_adapter_compat.go` 文件本体删除
2. 让 facade 只覆盖当前活跃路由所需能力
   - 保留实例访问、教师实例管理、proxy ticket、AWD target proxy、AWD defense SSH
   - 删除 dead workbench 方法和相应 service interface 依赖
3. 当前事实同步成“composition 本地薄 adapter”，不再把它写成 compat 残留

## Ownership Boundary

- `runtime/api/http`
  - 负责：HTTP 协议适配、当前开放路由的 service 依赖面
  - 不负责：保留已经下线路由的死接口形状
- `app/composition`
  - 负责：把实例 owner contract 适配给 runtime HTTP handler
  - 不负责：继续以 compat 文件名承载当前活跃 adapter
- `instance/application`
  - 负责：实例 owner 与 AWD defense workbench owner
  - 不负责：为未开放路由继续提供额外 facade 入口

## Change Surface

- Delete: `code/backend/internal/app/composition/runtime_adapter_compat.go`
- Add: `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/runtime_module_test.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler_test.go`
- Modify: `docs/todos/2026-05-11-runtime-container-ports-followup.md`
- Modify: `docs/architecture/backend/{01-system-architecture.md,03-container-architecture.md,07-modular-monolith-refactor.md}`
- Modify: `docs/architecture/features/AWD防守工作区与边界设计.md`

## Task Slices

### Slice 1: 删除 compat 文件名

目标：

- 用普通命名的 composition adapter 替代 `runtime_adapter_compat.go`
- `InstanceModule` 继续通过该 adapter 给 runtime handler 提供实例访问 facade

Validation:

- `cd code/backend && go test ./internal/app/composition -run RuntimeHTTPServiceAdapter -count=1`

### Slice 2: 裁掉 dead workbench 接口

目标：

- `runtime/api/http` 不再声明未开放路由对应的 4 个 workbench service 方法
- `InstanceModule` 不再给 runtime HTTP facade 注入无效的 `AWDDefenseWorkbenchService`

Validation:

- `cd code/backend && go test ./internal/module/runtime/api/http -count=1`

### Slice 3: 文档与 TODO 对齐

目标：

- TODO 标记前两刀已完成，保留后续 guardrail 和剩余边界事项
- 当前事实改成“composition 本地薄 adapter”，不再继续引用 compat 文件路径

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Risks

- 如果直接删除 adapter 而不补普通命名文件，会切断 runtime handler 的生产 wiring
- 如果只删文件名、不删 dead workbench 接口，`InstanceModule` 仍会继续注入失活依赖
- 如果更新了代码但没有同步当前事实文档，架构口径会继续把活跃 adapter 误写成 compat 残留

## Verification Plan

1. `cd code/backend && go test ./internal/app/composition -run RuntimeHTTPServiceAdapter -count=1`
2. `cd code/backend && go test ./internal/module/runtime/api/http -count=1`
3. `cd code/backend && go test ./internal/app/composition ./internal/module/runtime/api/http -count=1`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `bash scripts/check-workflow-complete.sh`
