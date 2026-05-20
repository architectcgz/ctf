# Runtime / Instance 边界 Phase 2 Slice 4 Implementation Plan

## Objective

在不触发物理删除的前提下，继续完成 `runtime -> instance` 的 owner 迁移收口：

- 把仍直接 new 旧 compat service 的存量调用点和测试改到 `internal/module/instance/*`
- 验证 `runtime/application/*` compat mirror 还能不能继续往“薄包装”推进
- 更新架构事实，明确当前残留的是 compat facade，而不是第二份实例业务实现

## Non-goals

- 不删除 `internal/module/runtime/application/{commands,queries}` 文件
- 不在这一轮搬动 `runtime/api/http`、`runtime/infrastructure/*`、`runtime/ports/*`
- 不在这一轮创建 `internal/module/instance/contracts`
- 不改实例、AWD、proxy 的外部 HTTP / WebSocket 契约

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice3-implementation-plan.md`
- `code/backend/internal/module/instance/application/{commands,queries}/*.go`
- `code/backend/internal/module/runtime/application/{commands,queries}/*.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/runtime/service_test.go`

## Current Baseline

- `internal/module/instance/*` 已经承接实例命令、查询、proxy ticket 和 maintenance 的真实 owner
- 但 `internal/module/runtime/application/{commands,queries}` 里还保留一份兼容实现，导致底层仍像是双 owner
- 部分测试和辅助构造仍直接依赖 `runtimecmd.NewInstanceService`、`runtimecmd.NewRuntimeMaintenanceService`、`runtimeqry.NewInstanceService`、`runtimeqry.NewProxyTicketService`
- 当前文档已经把这层认定为 compatibility mirror，但代码形态还不够薄

## Plan Revision After Guardrail Check

在开始实现后，额外验证发现这轮原定的“compat wrapper 直接转发到 `instance/application/*`”不可交付：

- `code/backend/internal/app/architecture_rules_test.go` 会阻止 `runtime -> instance application` 这类 concrete cross-module import
- 因此 `runtime/application/*` 现在还不能直接改成跨模块 wrapper
- 这轮实际收口范围回退为：
  - 保持 compat mirror 的现状不动
  - 继续减少外部直接 new compat service 的调用点
  - 在文档和 review 里明确记录 guardrail 阻塞，而不是把“薄包装”误记为已完成

## Chosen Direction

这轮按 guardrail 回环后的方向继续推进：

1. 保持 `runtime/application/*` compat mirror 的现状，不在这一轮强行做跨模块 wrapper
2. 把还直接 new compat service 的外部调用点先迁到 `instance/*`
   - 优先处理 `practice_flow_integration_test.go`
   - 其次处理 `runtime/service_test.go`
3. 更新架构事实与目标设计稿，明确：
   - compat mirror 仍存在
   - 它还不能直接转成 `runtime -> instance/application` wrapper
   - 下一刀需要在“删除 compat mirror”与“引入中性 landing zone”之间二选一

这样至少能先阻止 compat import path 继续向外扩散，同时不违反现有模块边界 guardrail。

## Ownership Boundary

- `instance/*`
  - 负责：实例业务规则、proxy ticket 规则、后台维护规则
  - 不负责：兼容旧 import path 的过渡入口
- `runtime/application/{commands,queries}`
  - 负责：短期兼容入口
  - 不负责：继续持有第二份实例业务实现
- `runtime/runtime` 与 `app/composition`
  - 负责：容器能力 builder 和实例访问装配
  - 不负责：回退为旧的 runtime owner 装配点

## Change Surface

- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/runtime/service_test.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice4-implementation-plan.md`
- Create: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice4-review.md`

## Task Slices

### Slice 1: guardrail 复核与方案回环

目标：

- 确认 `runtime/application/*` 能否安全转成跨模块 wrapper
- 如果不能，及时把本轮范围回退到可交付切片

Validation:

- `go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...`

Review focus:

- 当前 guardrail 是否允许 `runtime -> instance application` import
- 如果不允许，本轮是否已经及时回到更小的可审查范围

### Slice 2: 存量调用点迁移

目标：

- 直接 new compat service 的测试改到 `instance/*`
- compat 包只保留兼容用途，不再作为新增调用入口

Validation:

- `rg -n "runtimecmd\\.NewInstanceService|runtimecmd\\.NewRuntimeMaintenanceService|runtimeqry\\.NewInstanceService|runtimeqry\\.NewProxyTicketService" code/backend/internal/app code/backend/internal/module -g '*.go'`
- `go test ./internal/app/... ./internal/module/runtime/... ./internal/module/instance/...`

Review focus:

- 是否还有新的非 compat 场景继续依赖旧 import path
- `practice_flow_integration_test.go` 和 `runtime/service_test.go` 是否已经对齐新 owner

### Slice 3: 文档和 review 收口

目标：

- 当前事实明确为“compat facade 仍在，且这轮还不能直接压成跨模块 wrapper”
- review 存档写清 guardrail 阻塞、已完成项和残留项

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否把 compat facade 说成已经删除
- review 是否清楚区分“已瘦身”和“已彻底移除”

## Risks

- 如果文档把 compat mirror 说成“已经瘦身成 wrapper”，会和当前 guardrail 及代码事实冲突
- 如果只记录 guardrail 阻塞，不继续迁出外部调用点，后续仍会继续新增到旧 import path
- 这轮不触碰 compat mirror 本体，因此 duplicated implementation 仍是下一刀要收的结构债

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/...`
2. `cd code/backend && go test ./internal/app/... ./internal/module/runtime/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例 owner 继续固定在 `internal/module/instance/*`
- shared landing zone explicit：当前没有可直接复用的中性 landing zone，所以 compat mirror 还不能合法转发到 `instance/application/*`
- structure converges, not just behavior：这一刀先把新的外部调用点继续迁走，并把“wrapper 方案被 guardrail 拦下”显式写进事实源，避免后续再重复试错
- touched debt closure explicit：本轮在 guardrail 回环后，不再宣称已经收掉 compat mirror 的 duplicated implementation；剩余 debt 已作为下一阶段的明确任务继续跟踪
