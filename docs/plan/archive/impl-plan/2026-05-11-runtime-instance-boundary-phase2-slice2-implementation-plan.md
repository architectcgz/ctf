# Runtime / Instance 边界阶段 2 Slice 2 Implementation Plan

## Objective

在不搬动 `internal/module/runtime/*` 物理包的前提下，继续把 app/composition 层的模块边界收口到目标设计：

- 引入明确的 `ContainerRuntimeModule` 组合视图
- 让 `challenge`、`contest`、`ops`、`instance` 的装配签名依赖 `ContainerRuntimeModule`
- 让 app 层同时具备 `InstanceModule` 与 `ContainerRuntimeModule` 两个清晰视图
- 保持外部 HTTP 路由、runtime 底层实现和运行时行为不变

## Non-goals

- 不在这一轮创建 `internal/module/instance`
- 不移动 `internal/module/runtime/api/http`、`application/*`、`infrastructure/*` 到新目录
- 不改实例、AWD、proxy 的外部路径和响应契约
- 不改 `practice` 已经完成的 `InstanceModule` 依赖收口
- 不更新论文正文

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/backend/01-system-architecture.md`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/composition/challenge_module.go`
- `code/backend/internal/app/composition/contest_module.go`
- `code/backend/internal/app/composition/ops_module.go`
- `code/backend/internal/app/router.go`

## Current Baseline

- 上一刀已经引入 `InstanceModule`，`practice` 和实例路由不再直接依赖宽泛 `RuntimeModule`
- 但 app/composition 层里，`challenge`、`contest`、`ops` 仍然声明依赖 `RuntimeModule`
- `RuntimeModule` 这个名字在 app 层同时承载了两层含义：
  - 底层 `runtimemodule.Build(...)` 的唯一构建入口
  - challenge / contest / ops 看到的 container-facing 能力集合
- 这会让“实例 owner 视图”和“容器运行时视图”在 app 层仍然没有完全说清

## Chosen Direction

这轮只做 composition 命名和签名收口，不触碰底层物理包：

1. 在 `runtime_module.go` 中把当前 app 层视图明确命名为 `ContainerRuntimeModule`
2. 保留 `BuildRuntimeModule` / `RuntimeModule` 兼容别名，仅作为过渡 facade
3. 新增 `BuildContainerRuntimeModule(root)`，作为 router 使用的主 builder
4. `BuildInstanceModule`、`BuildChallengeModule`、`BuildContestModule`、`BuildOpsModule` 改为依赖 `*ContainerRuntimeModule`
5. `router.go` 和对应测试改用 `containerRuntime` 命名，体现当前 app 层的真实边界

这样可以把 app/composition 层先对齐到 `instance + container_runtime` 目标，再进入后续真正的底层拆包。

## Ownership Boundary

- `ContainerRuntimeModule`
  - 负责：镜像运行时、runtime probe、容器文件读写、运行时查询/统计，以及底层 `runtimemodule.Module` 的唯一构建入口
  - 不负责：实例 owner 的 HTTP handler、practice instance repo、practice runtime service
- `InstanceModule`
  - 负责：实例 handler、AWD target/defense SSH 访问入口、practice instance repo、practice runtime service
  - 不负责：镜像探针、容器文件写入、runtime 统计

## Change Surface

- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/challenge_module.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Modify: `code/backend/internal/app/composition/ops_module.go`
- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: `docs/architecture/backend/01-system-architecture.md`
- Modify: `docs/architecture/backend/04-api-design.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: 引入 ContainerRuntimeModule 组合视图

目标：

- 在 app/composition 层显式声明 `ContainerRuntimeModule`
- 保留 `RuntimeModule` 兼容别名，避免底层代码和历史引用一次性断裂

Validation:

- `rg -n "type ContainerRuntimeModule|type RuntimeModule =|BuildContainerRuntimeModule|BuildRuntimeModule" code/backend/internal/app/composition -g '*.go'`
- `go test ./internal/app/... -run 'Runtime|Container|Composition'`

Review focus:

- 是否仍只有一个底层 `runtimemodule.Build(...)` 入口
- 兼容别名是否只留在 app/composition 过渡层

### Slice 2: 把 challenge / contest / ops / instance 的装配签名切到 ContainerRuntimeModule

目标：

- `BuildChallengeModule`、`BuildContestModule`、`BuildOpsModule`、`BuildInstanceModule` 改接收 `*ContainerRuntimeModule`
- `router.go` 改用 `containerRuntimeModule` 命名
- builder 顺序和外部行为保持不变

Validation:

- `go test ./internal/app/... -run 'RouterBuild|CompositionBuilders|Challenge|Contest|Ops|Instance'`

Review focus:

- 是否只是命名/边界收口，而没有把实例入口重新混回 container 视图
- router 和测试中是否还残留“app 层只有一个 runtime 视图”的旧表达

### Slice 3: 文档回收和最终验证

目标：

- 更新当前架构事实，写清 app 层现在是 `InstanceModule + ContainerRuntimeModule`
- 设计稿标记阶段 2 的 slice 2 已落地，但底层物理包仍未拆分

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否把 `ContainerRuntimeModule` 说成已经等价于 `internal/module/container_runtime`
- API 文档是否明确本轮没有外部路径变化

## Risks

- 如果 `ContainerRuntimeModule` 和 `RuntimeModule` 同时各自 new 底层 runtime，会重复注册 cleaner 和其他 background job
- 如果只改了类型名，没改 router / 测试中的事实表达，后续维护仍会误把 app 层看成单一 runtime 视图
- 如果删掉 `RuntimeModule` 兼容别名过早，会扩大本轮改动面

## Verification Plan

1. `cd code/backend && go test ./internal/app/...`
2. `cd code/backend && go test ./internal/module/runtime/... ./internal/module/challenge/... ./internal/module/contest/... ./internal/module/ops/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：app 层显式分成 `instance` 与 `container_runtime` 两个组合视图
- shared builder landing zone explicit：底层构建仍统一落在 `BuildContainerRuntimeModule -> runtimemodule.Build(...)`
- not output-only：这轮不只是改文档名字，而是把 builder 签名和 router 组合表达都改到目标边界
- hidden redesign risk reduced：后续真实拆出 `internal/module/instance` 与 `internal/module/container_runtime` 时，app 层签名不需要再大改一轮
