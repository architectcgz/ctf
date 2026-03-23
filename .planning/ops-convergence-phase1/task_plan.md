# Task Plan

## Goal

启动 `system -> ops` 收敛，让审计、通知、风控、仪表盘等运营能力从 `system` 过渡到 `ops` owner。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 `system` 与 `ops` 当前职责 | pending | 找出真正 owner |
| 2. 在 `ops` 下建立审计/通知/风控/仪表盘边界 | pending | 先统一 contracts |
| 3. 调整 composition 与 router 依赖 | pending | 路由路径不变 |
| 4. 删除 `system` 中的过渡性 owner 逻辑 | pending | 不保留双 owner |
| 5. 收紧架构与文档 | pending | checklist + architecture |
| 6. 定向验证 | pending | audit / notification / dashboard / risk |

## Key Files

- `code/backend/internal/app/composition/system_module.go`
- `code/backend/internal/module/system/*`
- `code/backend/internal/module/ops/*`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`

## Acceptance Checks

- composition 优先装配 `ops`，而不是把 `system` 当运营 owner
- 审计、通知、风控、仪表盘能力不再作为“杂项系统模块”存在
- `ops` 不直接持有 `runtime` persistence 实现
- 定向测试通过：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/system ./internal/module/ops -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouter|TestFullRouter_AdminSystemAndNotificationStateMatrix' -count=1`

## Constraints

- 保持 `/api/v1/admin/dashboard`、`/api/v1/admin/audit-logs`、`/api/v1/notifications` 等路径不变
- 不新增“system facade -> ops”兼容层
- 通知事件消费与 websocket 行为保持稳定

