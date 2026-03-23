# Task Plan

## Goal

完成 `notification` 的 `system -> ops` Phase 2 硬迁移，并删除后端中的 `system` 模块实现。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点通知调用链 | completed | 已确认涉及 HTTP、websocket、ws ticket、practice 事件消费 |
| 2. 迁移通知实现到 `ops` | completed | handler / service / repository 已迁入 `ops` 分层目录 |
| 3. 调整 composition 与 router | completed | 通知路由与 websocket 已改由 `ops` handler 装配 |
| 4. 删除 `system` 模块实现 | completed | `internal/module/system` 已无后端实现文件 |
| 5. 定向验证 | completed | `ops/...`、`app`、`auth` 相关测试已通过 |

## Acceptance Checks

- `/api/v1/notifications`、`/api/v1/notifications/:id/read`、`/ws/notifications` 路径保持不变
- 通知 handler 来自 `internal/module/ops`
- practice 事件消费仍由通知服务注册
- websocket ticket 握手与通知推送行为保持稳定
- 后端代码中不再保留 `internal/module/system` 运行时实现

## Constraints

- 不保留 `system facade -> ops` 兼容层
- 不改变外部通知 API 协议
- 不碰前端脏改动
