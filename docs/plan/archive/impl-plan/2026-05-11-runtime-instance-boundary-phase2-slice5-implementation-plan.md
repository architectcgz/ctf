# Runtime / Instance 边界 Phase 2 Slice 5 Implementation Plan

## Objective

在不触碰 `runtime/application/*` compat mirror 本体的前提下，先补出一个合法的跨模块 landing zone：

- 新增 `internal/module/instance/contracts`，收口实例 owner 对外暴露的 command / query / proxy ticket / maintenance service 接口
- 把 `runtime/runtime` 这侧对实例 owner 的依赖，从 runtime 模块本地临时接口切到 `instance/contracts`
- 更新架构事实和目标设计稿，明确 compat mirror 的下一刀要基于这个 landing zone 继续收口

## Non-goals

- 不删除 `internal/module/runtime/application/{commands,queries}` compat mirror
- 不在这一轮修改 `runtime/api/http` 的外部 HTTP 契约
- 不在这一轮搬动 `runtime/infrastructure/*`、`runtime/ports/*` 中的底层容器适配
- 不在这一轮把 runtime application 中的 duplicated implementation 直接改成 wrapper

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice4-implementation-plan.md`
- `code/backend/internal/module/instance/application/{commands,queries}/*.go`
- `code/backend/internal/module/runtime/runtime/{module,adapters}.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/runtime/api/http/handler.go`

## Current Baseline

- `instance/application/*` 已经是实例 owner 的真实实现位置
- `runtime/runtime/adapters.go` 这侧已经不直接 new compat instance service，但仍通过 runtime 模块内部临时接口接住实例命令、查询和 proxy ticket service
- `runtime/application/*` compat mirror 还不能直接 import `instance/application/*`，因为会被 `code/backend/internal/app/architecture_rules_test.go` 拦下
- 因此下一刀要先给 `runtime -> instance` 之间补一个允许依赖、又不会把 owner 逻辑重新塞回 runtime 的稳定落点

## Chosen Direction

这轮先补 contract，不碰 compat mirror 本体：

1. 新增 `code/backend/internal/module/instance/contracts/services.go`
   - 定义 `InstanceCommandService`
   - 定义 `InstanceQueryService`
   - 定义 `ProxyTicketService`
   - 定义 `MaintenanceService`
   - 复用 `ProxyTicketClaims`
2. 把 `code/backend/internal/module/runtime/runtime/adapters.go` 改成直接依赖 `instance/contracts`
   - 去掉 runtime 模块内部那组临时 `runtimeHTTP*Service` 接口
   - 保持 `runtimeHTTPServiceAdapter` 行为不变，只换依赖落点
3. 更新文档
   - 当前事实：`runtime` 侧已经开始通过 `instance/contracts` 依赖实例 owner
   - 目标设计：compat mirror 下一刀应基于 `instance/contracts` 或同等级中性落点继续收口

## Ownership Boundary

- `instance/application/*`
  - 负责：实例业务规则和真实 owner 行为
  - 不负责：为 `runtime` 侧保留临时本地接口壳
- `instance/contracts`
  - 负责：实例 owner 暴露给外部模块的稳定 service contract
  - 不负责：承载业务规则实现或替代 application owner
- `runtime/runtime`
  - 负责：容器能力 builder、adapter 组合和实例访问 adapter
  - 不负责：继续声明一套 runtime 私有的 instance owner service 接口

## Change Surface

- Add: `code/backend/internal/module/instance/contracts/services.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice5-implementation-plan.md`
- Add: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice5-review.md`

## Task Slices

### Slice 1: 补 instance contract 落点

目标：

- 给 runtime 侧依赖实例 owner 补一个合法、显式、可复用的跨模块 contract 包

Validation:

- `go test ./internal/module/runtime/runtime ./internal/app/composition ./internal/module/runtime/api/http`

Review focus:

- contract 是否只暴露 owner service 能力，而不是把 runtime adapter 便利方法塞回 instance contract
- `contracts` 是否保持“接口定义 + 稳定结构”职责，没有重新长出业务实现

### Slice 2: runtime adapter 改依赖

目标：

- `runtime/runtime/adapters.go` 改成依赖 `instance/contracts`
- 去掉 runtime 模块内部那组临时 instance owner service 接口

Validation:

- `go test ./internal/module/runtime/runtime ./internal/app/composition ./internal/module/runtime/api/http`
- `go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...`

Review focus:

- 这轮是否只改变依赖落点，不改变 handler / adapter 的实际行为
- `ResolveTicket` / `ResolveAWDTargetAccessURL` 这类 claim 类型是否仍保持兼容

### Slice 3: 文档与 review 收口

目标：

- 当前事实明确为“instance/contracts 已落地，但 compat mirror 仍未删除”
- review 档清楚区分“补 landing zone”与“compat mirror 真正瘦身”

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否把 contract 落地误写成 compat mirror 已经完成收口
- review 是否清楚指出下一刀仍要处理 duplicated implementation

## Risks

- 如果 `instance/contracts` 暴露的是 runtime adapter 便利方法，而不是 owner service 本身，会把 contract 再次做歪
- 如果这轮把 “已有 landing zone” 说成 “compat mirror 已可直接改 wrapper”，会和当前代码事实冲突
- compat mirror 还在，下一刀如果不基于这个 contract 继续推进，runtime 底层仍会停留在双份实现

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/runtime ./internal/app/composition ./internal/module/runtime/api/http`
2. `cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例 owner 仍然固定在 `internal/module/instance/application/*`
- shared landing zone explicit：`internal/module/instance/contracts` 现在是 runtime 侧依赖实例 owner 的合法入口
- structure converges, not just behavior：这一刀先把 “合法依赖落点” 补齐，避免后续继续靠 runtime 私有临时接口或非法 application import 过渡
- touched debt closure explicit：本轮只收口 landing zone，不宣称 compat mirror 已完成瘦身；duplicated implementation 仍保留为下一刀的明确 debt
