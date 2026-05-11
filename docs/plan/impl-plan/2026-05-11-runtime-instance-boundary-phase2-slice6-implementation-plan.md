# Runtime / Instance 边界 Phase 2 Slice 6 Implementation Plan

## Objective

把 `runtime/application/*` 中仍承载实例业务的 compat mirror 收口成真正的薄层：

- `commands/instance_service.go` 改成委托 `instance/contracts.InstanceCommandService`
- `commands/runtime_maintenance_service.go` 改成委托 `instance/contracts.MaintenanceService`
- `queries/instance_service.go` 改成委托 `instance/contracts.InstanceQueryService`
- `queries/proxy_ticket_service.go` 改成委托 `instance/contracts.ProxyTicketService`
- 原本在 `runtime/application` 目录里覆盖实例 owner 行为的测试，改为直接验证 `instance/application/*` 的真实 owner；compat 包只保留最小 wrapper 测试

## Non-goals

- 不删除 `runtime/application/*` compat import path
- 不改 `runtime/application/commands/{runtime_cleanup_service,provisioning_service}.go`
- 不改 `runtime/runtime`、`runtime/infrastructure`、`runtime/api/http` 的生产 wiring
- 不在这一轮删除 `runtime/application` 目录内的历史测试文件

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice5-implementation-plan.md`
- `code/backend/internal/module/instance/contracts/services.go`
- `code/backend/internal/module/instance/application/{commands,queries}/*.go`
- `code/backend/internal/module/runtime/application/commands/instance_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service.go`
- `code/backend/internal/module/runtime/application/queries/instance_service.go`
- `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- `code/backend/internal/module/runtime/application/*_test.go`

## Current Baseline

- 生产 wiring 已经直接装配 `instance/application/*`，`runtime/application/*` 只剩兼容 import path 和同目录行为测试仍在使用
- `runtime/application/*` 不能直接 import `instance/application/*`，否则会被 `code/backend/internal/app/architecture_rules_test.go` 判成 concrete cross-module import 违规
- `instance/contracts` 已经作为合法 landing zone 落地，允许 compat 包基于 contract 做委托

## Chosen Direction

1. compat 包改成 contract-backed wrapper
   - `runtime/application/commands.InstanceService` 持有 `instancecontracts.InstanceCommandService`
   - `runtime/application/commands.RuntimeMaintenanceService` 持有 `instancecontracts.MaintenanceService`
   - `runtime/application/queries.InstanceService` 持有 `instancecontracts.InstanceQueryService`
   - `runtime/application/queries.ProxyTicketService` 持有 `instancecontracts.ProxyTicketService`
2. 保留 compat import path，但不再保留 duplicated implementation
3. 现有大体量行为测试改为直接调用 `instancecmd` / `instanceqry`
4. compat 包只补最小 wrapper 测试，验证委托路径和兼容方法（例如 `MaxAge()`）仍成立

## Ownership Boundary

- `instance/application/*`
  - 负责：实例命令、查询、proxy ticket、后台维护的真实业务实现
  - 不负责：兼容 import path 的历史命名
- `instance/contracts`
  - 负责：owner 对外暴露的稳定 service contract
  - 不负责：提供额外 convenience API 迁就 compat 层
- `runtime/application/*`
  - 负责：兼容 import path、转发到 owner contract
  - 不负责：继续承载实例业务规则或维护第二份实现

## Change Surface

- Modify: `code/backend/internal/module/runtime/application/commands/instance_service.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service.go`
- Modify: `code/backend/internal/module/runtime/application/queries/instance_service.go`
- Modify: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- Modify: `code/backend/internal/module/runtime/application/instance_service_test.go`
- Modify: `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice6-implementation-plan.md`
- Add: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice6-review.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: compat wrapper 化

目标：

- 把 runtime compat service 改成 `instance/contracts` 委托壳，不再保留 duplicated implementation

Validation:

- `cd code/backend && go test ./internal/module/runtime/application/...`

Review focus:

- compat 包是否只剩转发逻辑
- 是否没有重新引入对 `instance/application/*` 的非法 import

### Slice 2: 行为测试切回真实 owner

目标：

- 原本覆盖实例 owner 行为的 runtime/application 测试，改为直接验证 `instance/application/*`
- compat 包补最小 wrapper 测试，覆盖 delegate wiring 和 `MaxAge()` 兼容方法

Validation:

- `cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/...`

Review focus:

- 行为测试是否真的不再绑定 compat 实现
- wrapper 测试是否足够说明 compat import path 仍可用

### Slice 3: 文档收口

目标：

- 当前事实明确为：compat mirror 已压成 contract-backed wrapper，但 import path 仍保留
- 设计稿中剩余工作收敛为“迁空剩余调用方 / 删除 compat 包”

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否把 “wrapper 化完成” 误写成 “compat 已删除”
- 剩余 debt 是否写成可执行的下一步，而不是含糊 residual risk

## Risks

- 如果 compat wrapper 继续保留 repo/config/engine 构造逻辑，本轮就只是换壳名，没有真正去掉 duplicated implementation
- 如果只改代码不迁测试，业务行为覆盖会继续挂在 runtime compat 目录上，owner 迁移仍不彻底
- `ProxyTicketService.MaxAge()` 不是 owner contract 一部分，wrapper 必须自己保存 TTL 秒数，不能再把 convenience 方法反向塞回 contract

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/application/...`
2. `cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例业务 owner 继续固定在 `internal/module/instance/application/*`
- landing zone explicit：compat 包统一通过 `internal/module/instance/contracts` 委托 owner
- structure converges, not just behavior：本轮会直接删掉 runtime compat 的 duplicated implementation，而不是只改生产 wiring
- touched debt closure explicit：已知 debt 从“runtime/application 仍有双份实例实现”收敛为“compat import path 尚未删除，但已经是薄壳”
