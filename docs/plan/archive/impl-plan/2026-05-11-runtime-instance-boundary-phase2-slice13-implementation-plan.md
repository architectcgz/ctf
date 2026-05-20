# Runtime / Instance 边界 Phase 2 Slice 13 Implementation Plan

## Objective

补齐本轮边界收尾需要的 architecture guardrail：

- 防止 `composition/runtime_adapter_compat.go` 这类 compat facade 文件名回流
- 防止 `runtime/runtime.Module` 或 `Deps` 再暴露宽 `Engine` 结构面
- 防止 runtime HTTP handler 再为已下线的 AWD defense browser workbench 接口留 service 方法

## Non-goals

- 不恢复 `defense/files`、`defense/directories`、`defense/commands` 路由
- 不新增新的 container runtime port 或继续搬动物理目录
- 不修改当前实例访问、AWD target proxy、AWD defense SSH 的行为
- 不把架构 guardrail 扩成新的脚本检查；优先复用现有 Go test 落点

## Inputs

- `docs/todos/2026-05-11-runtime-container-ports-followup.md`
- `docs/architecture/backend/{03-container-architecture.md,04-api-design.md,07-modular-monolith-refactor.md}`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice11-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice12-implementation-plan.md`
- `code/backend/internal/app/composition/architecture_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/module/runtime/architecture_test.go`

## Current Baseline

- `runtime_adapter_compat.go` 已删除，但当前还没有机械化测试阻止同名 compat 文件回流
- `runtime/runtime.Module` 与 `Deps` 已改成显式 capability fields，但当前 guardrail 还没有直接验证 `Engine` 结构面不会重新出现
- 路由层已经明确缺失 `defense/files`、`defense/directories`、`defense/commands`，但 runtime handler interface 还没有单独的 architecture guardrail 来阻止死接口回流

## Chosen Direction

1. 复用现有 architecture test 落点
   - `internal/app/composition/architecture_test.go` 负责 compat facade 与失活注入约束
   - `internal/module/runtime/architecture_test.go` 负责宽 `Engine` 与 runtime HTTP service interface 约束
2. 继续让 `router_test.go` 承担外部路由事实
   - 本轮不另起一套路由 guardrail，只把内部 service interface 约束补齐
3. 把 guardrail 名单同步回 TODO 和架构文档
   - 让“当前事实”里能直接看到哪些测试在防回流

## Ownership Boundary

- `app/composition`
  - 负责：约束 app 层只保留当前活跃的 runtime HTTP adapter，不再恢复 compat facade 文件名或失活实例注入
  - 不负责：重新承载 AWD defense browser workbench owner
- `module/runtime`
  - 负责：维持 container-facing capability fields 和当前活跃的 runtime HTTP service interface
  - 不负责：重新暴露宽 `Engine`，或为已下线路由继续保留 dead service methods
- `app/router`
  - 负责：维持当前开放路由集合
  - 不负责：替代内部 architecture guardrail

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice13-implementation-plan.md`
- Modify: `code/backend/internal/app/composition/architecture_test.go`
- Modify: `code/backend/internal/module/runtime/architecture_test.go`
- Modify: `docs/todos/2026-05-11-runtime-container-ports-followup.md`
- Modify: `docs/architecture/backend/{03-container-architecture.md,04-api-design.md,07-modular-monolith-refactor.md}`

## Task Slices

### Slice 1: composition guardrail

目标：

- 阻止 `runtime_adapter_compat.go` 文件名回流
- 阻止 `InstanceModule` 再注入 `AWDDefenseWorkbenchService`

Validation:

- `cd code/backend && go test ./internal/app/composition -run 'TestCompositionDoesNotReintroduceRuntimeCompatFacade|TestInstanceModuleDoesNotInjectRetiredDefenseWorkbenchService' -count=1`

### Slice 2: runtime guardrail

目标：

- 阻止 `runtime/runtime.Module` 或 `Deps` 重新出现 `Engine` 结构面
- 阻止 `runtime/api/http` 再声明 retired defense workbench 方法

Validation:

- `cd code/backend && go test ./internal/module/runtime -run 'TestRuntimeModuleDoesNotExposeLegacyEngineSurface|TestAPIHTTPDoesNotDeclareRetiredDefenseWorkbenchMethods' -count=1`

### Slice 3: 文档与 TODO 对齐

目标：

- TODO 标记 architecture guardrail 已完成
- 架构事实源补上当前 guardrail 名单

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Risks

- 如果只在路由层保留“缺少 defense/files 路由”的断言，内部 service interface 仍可能悄悄回流
- 如果 guardrail 只做字符串搜索而不锁定具体结构面，后续容易通过改名绕回宽聚合
- 如果文档不记录新的 guardrail，后续 review 很难判断哪些边界已经机械化

## Verification Plan

1. `cd code/backend && go test ./internal/app/composition -count=1`
2. `cd code/backend && go test ./internal/module/runtime -count=1`
3. `cd code/backend && go test ./internal/app -run NewRouterUsesRuntimeHandlersForInstanceRoutes -count=1`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `bash scripts/check-workflow-complete.sh`
