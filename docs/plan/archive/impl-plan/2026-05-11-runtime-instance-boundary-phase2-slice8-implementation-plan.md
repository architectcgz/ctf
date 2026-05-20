# Runtime / Instance 边界 Phase 2 Slice 8 Implementation Plan

## Objective

在用户已确认不再保留 compat import path 的前提下，删除 `runtime/application/*` 中剩余的 instance / proxy ticket / maintenance thin wrapper：

- 删除 `runtime/application/commands/instance_service.go`
- 删除 `runtime/application/commands/runtime_maintenance_service.go`
- 删除 `runtime/application/queries/instance_service.go`
- 删除 `runtime/application/queries/proxy_ticket_service.go`
- 同步移除仅用于这层 wrapper 的最小委托测试、allowlist 和当前架构事实里的“待确认删除”表述

## Non-goals

- 不删除 `runtime/application/commands/provisioning_service.go`
- 不删除 `runtime/application/commands/runtime_cleanup_service.go`
- 不迁移 `runtime/application` 目录下仍承载 container capability 的 service
- 不处理 `internal/app/composition/runtime_adapter_compat.go` 与 `internal/module/runtime/runtime/adapters.go` 的重复适配逻辑

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice7-implementation-plan.md`
- `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice7-review.md`
- `code/backend/internal/module/runtime/application/{instance_service_test.go,proxy_ticket_service_test.go,commands/runtime_maintenance_service_test.go}`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Current Baseline

- 仓库内非测试生产调用已经不再依赖这 4 个 compat wrapper
- 当前残留依赖只剩：
  - wrapper 本体文件
  - wrapper 自测中的最小委托测试
  - `architecture_allowlist_test.go` 里的路径白名单
  - 文档里“已迁空但待确认删除”的过渡表述
- 用户已经明确确认删除

## Chosen Direction

1. 直接删除 4 个 thin wrapper 文件
2. 从现有测试文件中移除 wrapper-specific delegate tests，保留真实 owner 行为测试
3. 收缩 `architecture_allowlist_test.go`，移除这 4 条路径
4. 把当前事实更新为：
   - compat wrapper 已删除
   - 若后续需要重新引入跨模块兼容面，只能重新评估，不默认恢复旧路径

## Ownership Boundary

- `instance/application/*`
  - 负责：实例命令、查询、proxy ticket、maintenance 的唯一 owner
- `runtime/application/*`
  - 负责：仅保留 container capability 相关 application service
  - 不负责：实例业务兼容入口

## Change Surface

- Delete: `code/backend/internal/module/runtime/application/commands/instance_service.go`
- Delete: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service.go`
- Delete: `code/backend/internal/module/runtime/application/queries/instance_service.go`
- Delete: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- Modify: `code/backend/internal/module/runtime/application/instance_service_test.go`
- Modify: `code/backend/internal/module/runtime/application/proxy_ticket_service_test.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice8-implementation-plan.md`
- Add: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice8-review.md`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 删除 wrapper 本体与最小委托测试

目标：

- 删除 4 个 compat wrapper 文件
- 移除对应 delegate test，保留真实 owner 行为测试

Validation:

- `cd code/backend && go test ./internal/module/runtime/application/...`

Review focus:

- 是否没有误删仍被生产代码使用的 runtime service
- 是否不存在对已删 symbol 的残留引用

### Slice 2: 收缩 guardrail 与文档事实

目标：

- allowlist 不再保留已删路径
- 当前架构与目标设计稿明确写成“compat wrapper 已删除”

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否区分了“已删除”和“仍保留 container capability service”
- 历史 plan / review 是否保持为历史记录，不被误改成当前事实

### Slice 3: 集成复验

目标：

- 确认 runtime / instance / app 组合面没有残留依赖

Validation:

- `cd code/backend && go test -timeout 3m ./internal/module/runtime/application/... ./internal/module/runtime/runtime`
- `cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/composition ./internal/app/...`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 删除是否真正让 current facts 与代码一致
- 是否还存在未清掉的 legacy path 断言或字符串引用

## Risks

- 如果遗漏测试或 allowlist 中的路径引用，会留下“代码已删但 guardrail 仍假设存在”的漂移
- 如果误删 `runtime/application` 中仍承载 container capability 的 service，会把本轮范围扩大到非预期回归

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/application/...`
2. `cd code/backend && go test -timeout 3m ./internal/module/runtime/application/... ./internal/module/runtime/runtime`
3. `cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/composition ./internal/app/...`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例业务 compat owner 被彻底删除，唯一 owner 固定为 `instance`
- landing zone explicit：runtime 只保留 container capability service 与 ports
- structure converges, not just behavior：删除的是实际残留文件，不是只改文档说法
- touched debt closure explicit：用户已确认后，本轮直接关闭 “compat wrapper 是否继续保留” 这笔债
