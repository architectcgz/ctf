# Runtime / Instance 边界 Phase 2 Slice 7 Implementation Plan

## Objective

在不删除 `runtime/application/*` compat wrapper 的前提下，继续把剩余的容器适配边界收口清楚：

- 明确 `runtime/application/*` 这组 compat import path 在仓库内已经不再承担生产调用职责
- 把 `runtime/application` 和 `runtime/runtime.Module` 里仍散落的 Docker / ACL / 文件 / 镜像 / 容器指标能力收口成显式 container runtime ports
- 保持 compat wrapper 继续是薄层，不再允许 repo / config / engine 级构造逻辑回流

## Non-goals

- 不删除 `code/backend/internal/module/runtime/application/{commands,queries}` 下的 compat wrapper 文件
- 不迁移 `internal/module/runtime` 到新的 `container_runtime` 物理目录
- 不处理 `internal/app/composition/runtime_adapter_compat.go` 与 `internal/module/runtime/runtime/adapters.go` 的重复适配逻辑
- 不调整 `instance` owner 的业务行为或对外 API 契约

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice6-implementation-plan.md`
- `code/backend/internal/module/runtime/application/{container_file_service.go,image_runtime_service.go,container_stats_service.go}`
- `code/backend/internal/module/runtime/application/commands/{provisioning_service.go,runtime_cleanup_service.go}`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/module/runtime/ports/*.go`

## Current Baseline

- `runtime/application/*` 的 instance / proxy ticket / maintenance compat mirror 已经压成 `instance/contracts` thin wrapper
- 仓库内非测试生产调用已经不再直接 new 这些 compat wrapper；当前剩余调用主要是 wrapper 自测和兼容面保留
- `runtime/application` 里的 provisioning / cleanup / file / image / stats service 仍各自在本地声明一组 engine-ish 接口，container runtime ports 还没有成为统一落点
- `runtime/runtime.Module` 仍通过一个宽 `Engine` 接口把 Docker、ACL、文件、镜像、容器状态能力一次性带进来，边界意图不够显式

## Chosen Direction

1. `runtime/application/*` compat path 本轮继续保留，但只作为删除前的过渡层
   - 结论：仓库内生产调用面已经迁空，不再需要它承担业务 owner 或构造职责
   - 动作：不在这一轮删文件；删除动作单独等待用户确认
2. 在 `internal/module/runtime/ports` 新增 container runtime capability ports
   - 为 provisioning、cleanup、container file、image、managed container inventory、managed container stats 定义显式 port
3. `runtime/application/*` 与 `runtime/runtime.Module` 改为依赖这些 ports
   - 移除本地重复声明的 engine / runtime 接口，改成直接依赖 `runtime/ports`
   - 保持行为不变，只收口 owner 与 adapter 之间的边界表达
4. 文档同步明确当前事实
   - compat wrapper 仍在，但仓库内生产调用已不再依赖
   - 下一刀如果没有外部兼容需求，就是删除 wrapper，而不是继续补回新的 compat 构造逻辑

## Ownership Boundary

- `instance/application/*`
  - 负责：实例生命周期、访问 ticket、后台维护等真实业务实现
  - 不负责：兼容 import path 或底层 Docker 适配
- `runtime/application/*`
  - 负责：runtime 物理模块下仍保留的容器能力编排服务，以及已存在的 compat wrapper
  - 不负责：重新成为实例业务 owner
- `runtime/ports`
  - 负责：表达 container runtime capability ports，作为 application / runtime wiring 的统一消费边界
  - 不负责：承载 Docker SDK 细节实现
- `runtime/infrastructure`
  - 负责：实现上述 container runtime ports
  - 不负责：承接实例业务规则或跨模块 compat 逻辑

## Change Surface

- Add: `code/backend/internal/module/runtime/ports/container_runtime.go`
- Modify: `code/backend/internal/module/runtime/application/contracts.go`
- Modify: `code/backend/internal/module/runtime/application/container_file_service.go`
- Modify: `code/backend/internal/module/runtime/application/image_runtime_service.go`
- Modify: `code/backend/internal/module/runtime/application/container_stats_service.go`
- Modify: `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice7-implementation-plan.md`
- Add: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice7-review.md`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: container runtime ports 落地

目标：

- 在 `runtime/ports` 定义 provisioning / cleanup / file / image / managed container stats 等能力 port
- `runtime/application` 改为依赖这些 port，而不是在各文件里各自声明一组本地 runtime interface

Validation:

- `cd code/backend && go test ./internal/module/runtime/application/...`

Review focus:

- 新 port 是否仍是 consumer-side 最小能力，而不是重新造一个更宽的“总 engine”
- application 层是否没有引入 Docker SDK 或新的 concrete cross-layer import

### Slice 2: runtime module wiring 收口

目标：

- `runtime/runtime.Module.Engine` 改成由 container runtime ports 组合出来的依赖边界
- 生产 wiring 行为保持不变，但 capability owner 更显式

Validation:

- `cd code/backend && go test ./internal/module/runtime/runtime ./internal/app/composition`

Review focus:

- runtime module 是否只是改依赖表达，没有引入行为回归
- `runtime/application/*` compat wrapper 是否仍然保持薄层，不带回 repo / config / engine 构造逻辑

### Slice 3: 文档与 compat 决策同步

目标：

- 把“仓库内生产调用已迁空，compat wrapper 只剩删除决策”写成当前事实
- 下一步动作明确为“确认后删除 wrapper”，而不是继续保留含混状态

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否把“仓库内不再需要”误写成“已经删除”
- residual risk 是否只剩真正需要用户确认的删除动作

## Risks

- 如果 port 设计仍然把过多能力捆成一个宽接口，这轮只是在换名字，不算真正收口
- 如果为了保留 compat path 又重新把 repo / config / engine 级逻辑塞回 wrapper，本轮目标就会失效
- 如果文档没有明确区分“仓库内生产调用已迁空”和“wrapper 文件尚未删除”，后续很容易重复判断同一件事

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/application/...`
2. `cd code/backend && go test ./internal/module/runtime/runtime ./internal/module/runtime/... ./internal/module/instance/... ./internal/app/composition ./internal/app/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例业务 owner 继续固定在 `instance`；runtime 只保留 container-facing capability
- landing zone explicit：Docker / ACL / 文件 / 镜像 / 容器状态能力统一通过 `runtime/ports` 暴露
- structure converges, not just behavior：application 层不再各自维持一份本地 engine-ish interface
- touched debt closure explicit：compat wrapper 的 owner 债已经结束，本轮把剩余 container runtime capability 的边界表达收口；唯一保留的后续动作是用户确认后删除 wrapper 文件
