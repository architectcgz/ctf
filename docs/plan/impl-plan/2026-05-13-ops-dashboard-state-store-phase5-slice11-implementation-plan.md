# Ops Dashboard State Store Phase 5 Slice 11 Implementation Plan

## Objective

继续 phase 5 收窄 `ops` application concrete allowlist，把 dashboard 的 Redis 状态面下沉到模块 port / infrastructure adapter：

- 去掉 `ops/application/queries/dashboard_service.go -> github.com/redis/go-redis/v9`
- 保持 dashboard cache 命中优先、在线用户去重统计和缓存回填的现有行为不变

## Non-goals

- 不处理 `assessment/application/commands/profile_service.go` 或 `assessment/application/queries/recommendation_service.go` 的 Redis concrete allowlist
- 不改 dashboard HTTP contract、runtime stats provider 契约或告警阈值计算逻辑
- 不把 auth session payload 提升成新的跨模块稳定 contract

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- `code/backend/internal/module/ops/application/queries/dashboard_service_test.go`
- `code/backend/internal/module/ops/ports/dashboard.go`
- `code/backend/internal/module/ops/runtime/module.go`

## Current Baseline

- `DashboardService` 当前直接持有 `*redis.Client`
- dashboard cache 的 key、JSON 序列化和 `Get/Set` 细节散落在 query 层
- 在线用户统计通过 query 层直接 `Scan` + `MGet` auth session key，并在同一处解析 session payload
- allowlist 当前保留：
  - `ops/application/queries/dashboard_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把 dashboard 的 Redis 状态面表达成 `ops` 自己的 dashboard state store：

1. 在 `ops/ports` 新增 `DashboardStateStore` 和 cache snapshot value object
2. `DashboardService` 只依赖该 store，保留“何时优先命中缓存、何时降级实时查询、何时计算平均值和告警”的查询编排语义
3. `ops/infrastructure` 提供 Redis adapter，内部封装 dashboard cache key、auth session scan、JSON 编解码和 distinct online user count 细节
4. `ops/runtime/module.go` 统一构建该 store，并注入 dashboard query service

## Ownership Boundary

- `ops/application/queries/dashboard_service.go`
  - 负责：dashboard 查询编排、runtime stats 聚合、平均值与告警计算、缓存失败时降级到实时查询
  - 不负责：知道 Redis client、cache key、auth session 扫描、JSON 编解码或 `Scan/MGet/Get/Set` 细节
- `ops/infrastructure/dashboard_state_store.go`
  - 负责：实现 dashboard cache 和在线会话计数的 Redis 状态读写
  - 不负责：决定何时命中缓存、何时回填缓存或如何计算资源告警
- `ops/runtime/module.go`
  - 负责：装配 dashboard state store 并注入 dashboard query service
  - 不负责：把 Redis client 继续暴露回 `ops` application query surface

## Change Surface

- Add: `.harness/reuse-decisions/ops-dashboard-state-store-phase5-slice11.md`
- Add: `docs/plan/impl-plan/2026-05-13-ops-dashboard-state-store-phase5-slice11-implementation-plan.md`
- Add: `code/backend/internal/module/ops/infrastructure/dashboard_state_store.go`
- Add: `code/backend/internal/module/ops/ports/dashboard_state_context_contract_test.go`
- Modify: `code/backend/internal/module/ops/ports/dashboard.go`
- Modify: `code/backend/internal/module/ops/application/queries/dashboard_service.go`
- Modify: `code/backend/internal/module/ops/application/queries/dashboard_service_test.go`
- Modify: `code/backend/internal/module/ops/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 提取 dashboard state store port 与 Redis adapter
  - 目标：`DashboardService` 不再持有 `*redis.Client`，dashboard cache 与在线会话计数行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/ops/application/queries -run 'DashboardService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/ops/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：dashboard query 是否已经不再知道 Redis concrete；cache miss / cache write / 在线用户去重统计是否保持不变

- [x] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除 `ops/application/queries/dashboard_service.go` 的 Redis allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除本次 dashboard 实际收口的 allowlist；文档准确描述 ops 剩余 concrete 依赖范围

## Risks

- 如果 cache snapshot 结构和当前 `dto.DashboardStats` 序列化不兼容，现有缓存命中会回退成 miss
- 如果在线会话计数没有保持按 `user_id` 去重，dashboard 在线人数会漂移
- 如果 runtime 继续把旧 Redis client 注回 `DashboardService`，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/ops/application/queries -run 'DashboardService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/ops/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：`DashboardService` 继续持有 dashboard 查询编排语义，Redis 状态细节落回 infrastructure
- reuse point 明确：复用 phase5 已验证的“模块内窄 state store + runtime wiring”模式，不引入新的宽缓存抽象
- 这刀同时解决行为与结构：保留 dashboard cache / online session 现有行为，同时删掉 `ops` query surface 的 concrete Redis import
