# Task Plan

## Goal

启动 `system -> ops` 收敛，让审计、通知、风控、仪表盘等运营能力从 `system` 过渡到 `ops` owner。

当前 Phase 1 范围只覆盖：

- `audit`
- `dashboard`
- `risk`

`notification` 因为仍耦合 websocket / 事件消费，本轮继续留在 `system`。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 `system` 与 `ops` 当前职责 | completed | 已确认 Phase 1 先迁 `audit / dashboard / risk`，`notification` 后置 |
| 2. 在 `ops` 下建立审计/通知/风控/仪表盘边界 | in_progress | `audit / dashboard / risk` 已迁入 `ops`，`notification` 待下一阶段 |
| 3. 调整 composition 与 router 依赖 | completed | admin 审计/仪表盘/风控路由已改由 `ops` handler 装配，路径不变 |
| 4. 删除 `system` 中的过渡性 owner 逻辑 | in_progress | `audit / dashboard / risk` 已从 `system` 删除，`notification` 仍保留 |
| 5. 收紧架构与文档 | completed | phase1 planning、roadmap 与架构文档已同步 |
| 6. 定向验证 | completed | 已完成 `ops / system / app / auth` 限核定向测试 |

## Key Files

- `code/backend/internal/app/composition/system_module.go`
- `code/backend/internal/module/system/*`
- `code/backend/internal/module/ops/*`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`

## Acceptance Checks

- composition 优先装配 `ops` 的审计/仪表盘/风控能力，而不是把 `system` 当运营 owner
- 审计、风控、仪表盘不再作为 `system` 的 owner 逻辑存在
- `notification` 路由与 websocket 行为保持在 `system`，直到后续单独迁移
- `ops` 不直接持有 `runtime` persistence 实现
- 定向测试通过：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/system ./internal/module/ops -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouter|TestFullRouter_AdminSystemAndNotificationStateMatrix' -count=1`

## Constraints

- 保持 `/api/v1/admin/dashboard`、`/api/v1/admin/audit-logs`、`/api/v1/notifications` 等路径不变
- 不新增“system facade -> ops”兼容层
- 通知事件消费与 websocket 行为保持稳定
