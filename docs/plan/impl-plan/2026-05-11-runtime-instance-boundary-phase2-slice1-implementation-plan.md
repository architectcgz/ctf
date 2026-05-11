# Runtime / Instance 边界阶段 2 Slice 1 Implementation Plan

## Objective

按 `docs/design/backend-module-boundary-target.md` 的阶段 2 先落一段 app 层迁移：

- 在 `internal/app/composition` 层引入独立 `InstanceModule`
- 把实例 handler、practice instance repository、practice runtime service 从 `RuntimeModule` 中迁出
- 让 `practice` 组合和实例相关路由先依赖 `InstanceModule`
- 保持 `internal/module/runtime/*` 现有底层实现和外部 HTTP 行为不变

## Non-goals

- 不在这一轮直接重命名 `internal/module/runtime` 为 `instance`
- 不搬动 `runtime/api/http`、`runtime/application/*`、`runtime/infrastructure/*` 到新目录
- 不拆 challenge / contest / ops 当前仍依赖的 runtime probe、container files、runtime stats 能力
- 不改实例 API 路径、proxy ticket 契约、AWD defense SSH 行为或论文内容

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/backend/01-system-architecture.md`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/module/practice/ports/ports.go`

## Current Baseline

- `RuntimeModule` 同时暴露：
  - 实例 handler
  - practice instance repository
  - practice runtime service
  - challenge image/runtime probe
  - ops runtime query/stats
  - contest container file writer
- `BuildRuntimeModule` 还负责 SSH gateway 注册和 runtime handler 装配
- `BuildPracticeModule`、实例路由和 AWD target / defense SSH 路由都直接依赖 `RuntimeModule`
- 这让 app 层看不出“实例 owner”和“容器/运行时适配能力”的边界

## Chosen Direction

这轮先在 composition 边界上收口 owner 视图，不重复构建底层 runtime：

1. 保留 `BuildRuntimeModule(root)` 作为唯一底层 `runtimemodule.Build(...)` 调用点
2. 缩窄 `RuntimeModule` 公开面，只保留 challenge / contest / ops 仍需要的运行时与容器能力
3. 新增 `BuildInstanceModule(root, runtime)`，从同一个底层 `runtimemodule.Module` 导出：
   - 实例 handler
   - practice instance repository
   - practice runtime service
   - SSH gateway 注册和实例 handler 装配
4. `BuildPracticeModule` 和 `registerUserRoutes` 改依赖 `InstanceModule`

这样能先让下游依赖方向收口，而不在本轮把 runtime 底层实现大搬家。

## Ownership Boundary

- `RuntimeModule`
  - 负责：challenge runtime probe、image runtime、contest container files、ops runtime query/stats，以及底层 `runtimemodule.Module` 的唯一构建入口
  - 不负责：继续对外充当实例 owner 的统一 facade
- `InstanceModule`
  - 负责：实例 handler、practice instance repository、practice runtime service、实例 proxy/SSH access 的 app 层装配
  - 不负责：镜像探测、容器文件写入、runtime 统计等容器/平台适配能力
- `BuildPracticeModule`
  - 负责：只依赖实例 owner 视图和 challenge / assessment contract
  - 不负责：继续透传整个 `RuntimeModule`

## Change Surface

- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Create: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/practice_module.go`
- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `docs/architecture/backend/01-system-architecture.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: 在 composition 层拆出 InstanceModule

目标：

- 新增 `InstanceModule`
- `RuntimeModule` 移除实例 handler、practice runtime/repository 暴露
- 保证底层 `runtimemodule.Module` 只构建一次

Validation:

- `rg -n "PracticeRuntimeService|PracticeInstanceRepository|BuildHandler\\(|ProxyTicketService|SSHExecutor" code/backend/internal/app/composition code/backend/internal/app/router.go -g '*.go'`
- `go test ./internal/app/... -run 'Runtime|Instance|Composition|RouterBuild'`

Review focus:

- 是否有重复 runtime 构建或重复 background job / SSH gateway 注册
- RuntimeModule 公开面是否真的变窄

### Slice 2: 把 practice 和实例路由切到 InstanceModule

目标：

- `BuildPracticeModule` 改接收 `*InstanceModule`
- 用户实例路由、教师实例路由、AWD target proxy / defense SSH 路由改用 `InstanceModule.Handler`

Validation:

- `go test ./internal/app/... -run 'Practice|Router|Route|CompositionBuilders'`

Review focus:

- 下游依赖方向是否从 `RuntimeModule` 收口到 `InstanceModule`
- 路由行为是否保持原样

### Slice 3: 文档回收和最终验证

目标：

- 更新当前架构事实，写清 app 层已经引入 `InstanceModule` 视图
- 在设计稿里标记阶段 2 的首个 slice 已落地

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否写当前事实，而不是把 `internal/module/runtime` 已经完全拆完写成既成事实
- 设计稿是否明确这只是阶段 2 的第一刀

## Risks

- 如果 `BuildInstanceModule` 重新 new 一套底层 `runtimemodule.Module`，会重复注册 cleaner / SSH gateway，并造成双份运行时状态
- `router_routes.go` 中实例相关入口和 AWD target proxy / defense SSH 入口都走 `runtime.Handler`，漏改会让 owner 边界只改一半
- `practice_flow_integration_test.go` 直接拿 `RuntimeModule` 的 practice 字段，漏改会造成编译失败

## Verification Plan

1. `cd code/backend && go test ./internal/app/...`
2. `cd code/backend && go test ./internal/module/runtime/... ./internal/module/practice/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例 owner 先在 app/composition 层落到 `InstanceModule`
- shared builder landing zone explicit：底层仍统一落在 `BuildRuntimeModule -> runtimemodule.Build(...)`
- this slice is not output-only：它不是单纯换名，而是先把 practice 和实例路由的依赖入口改到新的 owner 视图
- hidden redesign risk reduced：后续底层把 `runtime` 真正拆成 `instance + container_runtime` 时，不需要再次牵动 practice 和路由组合边界
