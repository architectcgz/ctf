# Runtime / Instance 边界 Phase 2 Slice 9 Implementation Plan

## Objective

删除 `internal/module/runtime/runtime/adapters.go` 中已经没有生产 wiring 的 `runtimeHTTPServiceAdapter`，收掉与 `internal/app/composition/runtime_adapter_compat.go` 的重复实现，并同步清理 runtime 包里只为这份 dead adapter 服务的重复测试与旧文档路径。

## Non-goals

- 不改 `internal/app/composition/runtime_adapter_compat.go` 的对外行为
- 不改 `runtimePracticeServiceAdapter`、`runtimeChallengeServiceAdapter`、`runtimeOpsStatsProviderAdapter`
- 不调整 runtime HTTP handler 的路由、cookie、proxy ticket 或 AWD defense workbench 契约
- 不处理前端或其他 harness 在途改动

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docs/architecture/features/AWD防守工作区与边界设计.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice7-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice8-implementation-plan.md`
- `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice7-review.md`
- `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice8-review.md`
- `code/backend/internal/app/composition/runtime_adapter_compat.go`
- `code/backend/internal/app/composition/runtime_module_test.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/runtime/adapters_test.go`

## Current Baseline

- 生产 wiring 里的 runtime HTTP service adapter 已经只在 `internal/app/composition/runtime_adapter_compat.go` 构造，入口是 `BuildInstanceModule(...)`
- `internal/module/runtime/runtime/adapters.go` 里仍保留一整份同名 `runtimeHTTPServiceAdapter` 和同一组 AWD defense helper
- 这份 runtime 包内 HTTP adapter 已经没有生产引用，只剩 `internal/module/runtime/runtime/adapters_test.go` 在 new 它
- `docs/architecture/backend/03-container-architecture.md` 仍引用已删除的 `runtime/application/queries/proxy_ticket_service.go`、`runtime/application/commands/runtime_maintenance_service.go`

## Chosen Direction

1. 删除 `runtime/runtime/adapters.go` 里的 dead `runtimeHTTPServiceAdapter` 及其专属 helper / error constructor
2. 从 `runtime/runtime/adapters_test.go` 删除只覆盖 dead HTTP adapter 的重复测试，保留 practice / challenge / ops 仍在使用的 adapter 测试
3. 把 runtime 包测试里仍有价值但 composition 侧尚未覆盖的行为测试补到 `internal/app/composition/runtime_module_test.go`
4. 更新当前事实文档，明确：
   - runtime HTTP adapter 的唯一生产 owner 在 `internal/app/composition/runtime_adapter_compat.go`
   - runtime 物理模块只保留 practice / challenge / ops 还在使用的底层 adapter

## Ownership Boundary

- `internal/app/composition/runtime_adapter_compat.go`
  - 负责：runtime HTTP handler 需要的实例命令 / 查询 / proxy ticket / AWD defense workbench 适配
- `internal/module/runtime/runtime/adapters.go`
  - 负责：practice / challenge / ops 仍在使用的 runtime adapter
  - 不负责：重复承载一份 runtime HTTP adapter

## Change Surface

- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`
- Modify: `code/backend/internal/app/composition/runtime_module_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/architecture/backend/03-container-architecture.md`
- Modify: `docs/architecture/features/AWD防守工作区与边界设计.md`
- Add: `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice9-implementation-plan.md`
- Add: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice9-review.md`

## Task Slices

### Slice 1: 删除 dead HTTP adapter 与重复测试

目标：

- 删掉 runtime 模块里没有生产引用的 `runtimeHTTPServiceAdapter`
- 删除 runtime 包里只服务这份 dead code 的重复测试
- 在 composition 测试里补齐缺失覆盖

Validation:

- `cd code/backend && go test -timeout 3m ./internal/module/runtime/runtime ./internal/app/composition`

Review focus:

- 是否没有误删 practice / challenge / ops 仍依赖的 adapter
- composition 侧是否补上删除 dead test 后缺失的行为覆盖

### Slice 2: 对齐当前事实文档

目标：

- 文档不再引用已删除的 runtime application owner 路径
- 文档明确 runtime HTTP adapter 的唯一生产 owner

Validation:

- `python3 scripts/check-docs-consistency.py`

Review focus:

- 当前事实是否反映真实生产 wiring，而不是历史过渡结构

### Slice 3: 集成复验

目标：

- 确认 runtime / composition / docs 没有残留引用或回归

Validation:

- `cd code/backend && go test -timeout 3m ./internal/module/runtime/runtime ./internal/app/composition`
- `cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/...`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- dead runtime HTTP adapter 是否已经彻底退出当前代码和事实源
- 是否还存在 runtime 模块内的重复 HTTP 适配逻辑

## Risks

- 如果 runtime 包测试里仍混着 practice adapter 相关断言，删除 dead HTTP adapter 时容易误删仍有价值的测试
- 如果只删代码不补文档，当前事实会继续把旧 owner 路径写成活跃结构

## Verification Plan

1. `cd code/backend && go test -timeout 3m ./internal/module/runtime/runtime ./internal/app/composition`
2. `cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：runtime HTTP adapter 的唯一生产 owner 固定在 composition，runtime 物理模块不再留平行副本
- landing zone explicit：仍需保留的 practice / challenge / ops adapter 继续留在 `runtime/runtime/adapters.go`
- structure converges, not just behavior：本轮删除的是失活实现和重复测试，不是只在文档里宣称收口
- touched debt closure explicit：slice7/8 review 里记录的“重复 runtime HTTP adapter”债务，本轮直接在 touched surface 内关闭
