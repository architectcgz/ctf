# ops 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/ops/`、`code/backend/internal/app/composition/ops_module.go`
> 替代：无

## 定位

`ops` 是运营支撑 owner，负责审计日志、站内通知、通知 WebSocket、竞赛实时 relay、风险查询和运行时 dashboard。

## 事实来源

- HTTP 入口：`code/backend/internal/module/ops/api/http/`
- 命令用例：`code/backend/internal/module/ops/application/commands/`
- 查询用例：`code/backend/internal/module/ops/application/queries/`
- 对外事件和错误：`code/backend/internal/module/ops/contracts/`
- 持久化与 Redis 适配：`code/backend/internal/module/ops/infrastructure/`
- 装配入口：`code/backend/internal/module/ops/runtime/`、`code/backend/internal/app/composition/ops_module.go`

## 当前设计

- `ops/api/http`
  - 负责：审计、dashboard、risk、notification HTTP 入口和 notification WebSocket。
  - 不负责：直接访问 GORM、Redis、auth token concrete 或 runtime module concrete。
- `ops/application/commands`
  - 负责：审计记录、通知创建/标记、竞赛 realtime dispatcher / service。
  - 不负责：决定 auth 登录语义、contest 比赛状态或 practice 提交判分。
- `ops/application/queries`
  - 负责：审计查询、dashboard 查询、通知查询和风险查询。
  - 不负责：教师复盘聚合或业务 owner 状态修改。
- `ops/infrastructure`
  - 负责：audit / notification / risk repository、dashboard state store、contest realtime stream。
  - 不负责：拥有 runtime node 源数据、token 源数据或 contest realtime event 源事实。
- `OpsModule`
  - 负责：向其他模块提供 audit recorder、WebSocket manager、notification handler 和 realtime relay 能力。
  - 不负责：把业务模块的通知触发逻辑写进 HTTP handler。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `GET /api/v1/admin/audit-logs` | `ops/api/http.AuditHandler.ListAuditLogs` | `queries.AuditService` |
| `GET /api/v1/admin/dashboard` | `DashboardHandler.GetDashboard` | `queries.DashboardService` |
| `GET /api/v1/admin/cheat-detection` | `RiskHandler.GetCheatDetection` | `queries.RiskService` |
| `POST /api/v1/admin/notifications` | `NotificationHandler.PublishAdminNotification` | `commands.NotificationService` |
| `GET /api/v1/notifications` | `NotificationHandler.ListNotifications` | `queries.NotificationService` |
| `PUT /api/v1/notifications/:id/read` | `NotificationHandler.MarkAsRead` | `commands.NotificationService` |
| `GET /ws/notifications` | `NotificationHandler.ServeWS` | notification WebSocket manager + token service |
| `GET /ws/contests/:id/announcements`、`/scoreboard`、`/awd-preview` | `contest/api/http.RealtimeHandler` consuming ops manager | contest realtime relay |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Audit command/query | `ops/application/commands/audit_service.go`、`queries/audit_service.go` | 审计写入和查询 | 决定业务动作是否允许 |
| Notification command/query | `notification_service.go` | 创建通知、标记已读、列表查询、源事件幂等 | 业务事件生成 |
| Contest realtime service / dispatcher | `contest_realtime_service.go`、`contest_realtime_dispatcher.go` | 竞赛公告、榜单、AWD preview 事件 relay | contest 状态计算 |
| Dashboard query service | `queries/dashboard_service.go` | runtime stats、在线会话、平均值和告警聚合 | runtime node 源数据维护 |
| Risk query service | `queries/risk_service.go` | 风险/作弊检测查询 | 自动处罚 |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `audit_logs` | `ops/entity.AuditLog` | 用户操作审计事实 | ops audit command / injected recorder |
| `notifications` | `ops/entity.Notification` | 站内通知、已读状态、`source_event_key` 幂等 | notification command / outbox handler |
| `notification_batches` | `NotificationBatch` | 批量通知记录 | notification command |
| Redis dashboard cache / sessions | `dashboard_state_store.go` | dashboard 缓存和在线会话扫描 | dashboard query |
| Redis contest realtime stream | `contest_realtime_stream.go` | 每个 API 副本本地 WebSocket relay fanout | contest realtime dispatcher |
| Platform outbox / Redis Stream | `internal/platform/events` | `notification.created/read` 跨副本可恢复 fanout | root background jobs |

## 边界

- `ops` 拥有审计日志、通知记录、通知 WebSocket fanout、竞赛 realtime relay 和运营 dashboard 查询。
- `auth`、`challenge`、`practice`、`contest` 产生日志或事件，`ops` 负责记录与推送。
- `container_runtime` 提供 runtime stats；ops 只聚合和展示。
- `auth` 提供 token service；ops notification WebSocket 只消费认证能力。

## 主要用例

- 记录和查询审计日志。
- 创建、查询、标记站内通知。
- 通知 WebSocket 推送 `notification.created` / `notification.read`。
- 竞赛公告、榜单和 AWD preview realtime relay。
- 管理端 dashboard 聚合在线会话、运行时统计和风险信息。

## 数据与副作用

- PostgreSQL：审计日志、通知和通知批次。
- Redis：dashboard cache、在线会话扫描、contest realtime stream、notification stream fanout。
- WebSocket：本地 API 副本持有连接，Redis Stream 保证每个副本能消费本地推送事件。
- Outbox：业务模块写 platform outbox，ops 注册 handler 创建通知或 relay。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `auth` | notification WebSocket token service | `OpsModule.BuildNotificationHandler` |
| `container_runtime` | dashboard runtime stats | `composition.BuildOpsModule` |
| `challenge` | 发布检查通知事件 | outbox handler |
| `practice` | Flag accepted 通知事件 | outbox handler |
| `contest` | realtime relay events | contest realtime wiring |

## Guardrail

- `code/backend/internal/module/ops/architecture_test.go`：约束 API / commands / queries / ports 分层和 runtime typed deps。
- `code/backend/internal/module/ops/ports/dashboard_state_context_contract_test.go`：约束 dashboard state store context。
- `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`：覆盖通知 HTTP 和 WebSocket 集成。
- `code/backend/internal/app/full_router_access_integration_test.go`：覆盖路由访问层行为。

## 已知限制

- WebSocket fanout 是 API 副本本地连接模型；跨副本靠 Redis Stream fanout，不是独立消息服务。
