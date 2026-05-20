# Runtime / Instance 边界 Phase 2 Slice 10 Implementation Plan

## Objective

把 `internal/app/composition/runtime_adapter_compat.go` 中 AWD defense workbench 的文件 / 命令访问逻辑迁回 `instance` owner：

- `instance/contracts` 暴露稳定的 AWD defense workbench service contract
- `instance/application` 持有 scope、路径、backup、命令执行等业务规则
- composition 只负责把实例 owner contract 和 container file runtime capability 接起来

## Non-goals

- 不恢复 `runtime/api/http` 中当前禁用的 AWD defense file / command 路由
- 不删除 `runtime_adapter_compat.go` 文件本体
- 不改 AWD defense SSH ticket 的外部 HTTP 契约
- 不在这一轮搬动 `internal/module/runtime` 到新的物理目录

## Inputs

- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice7-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice9-implementation-plan.md`
- `code/backend/internal/app/composition/runtime_adapter_compat.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/instance/contracts/services.go`
- `code/backend/internal/module/runtime/ports/container_runtime.go`

## Current Baseline

- `runtime_adapter_compat.go` 仍同时承载两类职责：
  - runtime HTTP facade 的实例命令 / 查询 / proxy ticket 转发
  - AWD defense workbench 的 scope、路径、敏感文件、backup、命令超时判断
- 这些 AWD defense workbench 判断依赖实例 owner 语义，而不是单纯的容器 provider 语义：
  - `FindAWDDefenseSSHScope(...)`
  - `EditablePaths`
  - contest / service / user scope
- `runtime` 侧已经有 container file runtime capability，但这组业务判断还没有 owner-side landing zone

## Chosen Direction

1. 在 `instance/contracts` 新增 `AWDDefenseWorkbenchService`
2. 在 `instance/application` 新增 owner 应用服务
   - 负责读取 / 列目录 / 保存 / 执行命令
   - 直接承接当前 compat 层里的路径与权限判断
3. composition 新增 provider adapter
   - 只负责把 runtime container file capability 映射给 instance owner 使用
4. `runtime_adapter_compat.go` 改成纯转发
   - 保留 runtime HTTP handler 需要的 facade 形状
   - 不再保留 AWD defense workbench 业务 helper

## Ownership Boundary

- `instance/contracts`
  - 负责：实例 owner 对 AWD defense workbench 暴露的稳定 service contract
  - 不负责：表达 Docker / engine 细节
- `instance/application`
  - 负责：AWD defense workbench 的 scope、路径、敏感文件、backup、命令执行规则
  - 不负责：直接依赖 runtime 模块 concrete implementation
- composition
  - 负责：owner contract 的 wiring 与 provider capability mapping
  - 不负责：继续承载 owner 业务判断

## Change Surface

- Add: `code/backend/internal/module/instance/application/awd_defense_workbench_service.go`
- Modify: `code/backend/internal/module/instance/contracts/services.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/runtime_adapter_compat.go`
- Modify: `code/backend/internal/app/composition/runtime_module_test.go`
- Add: `code/backend/internal/module/instance/application/awd_defense_workbench_service_test.go`
- Add: `docs/todos/2026-05-11-runtime-container-ports-followup.md`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice10-implementation-plan.md`

## Task Slices

### Slice 1: owner contract 与应用服务落点

目标：

- 在 `instance` owner 侧定义 AWD defense workbench contract 与实现
- 把 compat 层现有规则迁过去，不改行为

Validation:

- `cd code/backend && go test ./internal/module/instance/application -run AWDDefenseWorkbench -count=1`

Review focus:

- 规则 owner 是否已经回到 `instance`
- `instance/application` 是否没有直接依赖 runtime concrete type

### Slice 2: composition 改成纯 wiring

目标：

- `runtime_adapter_compat.go` 只保留 facade 转发
- `instance_module.go` 负责 new owner service 与 provider adapter

Validation:

- `cd code/backend && go test ./internal/app/composition -run RuntimeHTTPServiceAdapter -count=1`

Review focus:

- composition 是否不再保留路径 / 敏感文件 / backup 等业务 helper
- runtime HTTP facade 行为是否保持兼容

### Slice 3: 集成复验

目标：

- owner / composition 两层都能通过最小相关测试

Validation:

- `cd code/backend && go test ./internal/module/instance/... ./internal/app/composition -count=1`
- `bash scripts/check-consistency.sh`

Review focus:

- 是否还残留 compat 业务判断
- 是否引入新的跨模块 concrete import

## Risks

- 如果直接让 `instance/application` 依赖 `runtime/ports`，会形成新的 owner 反向耦合
- 如果只搬公开方法、不搬 helper 规则，compat 层仍然会留下同样的业务债
- 如果 composition adapter 过厚，会把新的 provider mapping 重新长成第二份业务逻辑

## Verification Plan

1. `cd code/backend && go test ./internal/module/instance/application -run AWDDefenseWorkbench -count=1`
2. `cd code/backend && go test ./internal/app/composition -run RuntimeHTTPServiceAdapter -count=1`
3. `cd code/backend && go test ./internal/module/instance/... ./internal/app/composition -count=1`
4. `bash scripts/check-consistency.sh`

## Architecture-Fit Evaluation

- target ownership explicit：AWD defense workbench 的业务规则回到 `instance` owner
- landing zone explicit：composition 只接 contract 与 container capability，不再藏业务 helper
- structure converges, not just behavior：不是只把方法换个地方，而是连 helper owner 一起迁走
- touched debt closure explicit：本轮直接收掉 `runtime_adapter_compat.go` 上这块已知 owner-mixed debt，而不是继续记 follow-up
