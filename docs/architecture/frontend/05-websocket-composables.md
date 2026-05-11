# 前端 WebSocket 与实时 Composable 设计

> 状态：Current
> 事实源：`code/frontend/src/composables/useWebSocket.ts`、`code/frontend/src/features/*/model/*Realtime*.ts`、`code/backend/internal/app/router.go`
> 替代：无

## 定位

本文档只说明前端实时连接层的 owner、连接生命周期和当前落地的实时能力。

- 覆盖：`useWebSocket()`、通知实时、竞赛公告实时、排行榜实时、AWD 预览实时。
- 不覆盖：所有通用 composable 的总览。分页、登录、tab 同步等通用能力继续在各自 owning 文档中说明。

## 当前设计

- `code/frontend/src/composables/useWebSocket.ts`
  - 负责：获取一次性 ticket、拼接 WS 地址、发送心跳、pong 超时关闭、指数退避重连、鉴权失败登出和组件卸载清理
  - 不负责：理解具体业务事件 payload，也不直接写业务 store

- `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
  - 负责：先同步首屏通知，再建立 `notifications` 实时连接，并把 `notification.created / notification.read` 写回 `notification` store
  - 不负责：在 store 内自己建连，或把通知详情页的页面交互塞回实时层

- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts`、`code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.ts`、`code/frontend/src/features/awd-inspector/model/useContestAwdPreviewRealtime.ts`
  - 负责：把不同业务端点包装成页面可调用的 `start / stop / status`
  - 不负责：在实时通道里传输整页状态；当前公告和排行榜更新仍以“事件触发后 HTTP 刷新”为主

## 1. 连接入口

前端所有业务实时连接都以 `useWebSocket(endpoint, handlers)` 为统一入口。

当前固定规则：

| 项目 | 当前值 |
| --- | --- |
| WS 基地址 | `import.meta.env.VITE_WS_BASE_URL || '/ws'` |
| 认证方式 | 先调用 `POST /auth/ws-ticket`，再把 ticket 放进 query |
| 连接状态 | `idle` / `connecting` / `open` / `closed` / `error` |
| 鉴权失败 close code | `4001`、`4401` |

对应代码：

- ticket 获取：`code/frontend/src/api/auth.ts`
- 环境变量与重连常量：`code/frontend/src/utils/constants.ts`

## 2. 连接生命周期

`useWebSocket()` 当前链路：

1. `connect()` 时先调用 `getWsTicket()`
2. 以 `${wsBase}/${endpoint}?ticket=...` 建立连接
3. `open` 后把状态置为 `open`，并启动心跳
4. 每 `30s` 发送一次 `ping`
5. 若 `60s` 内没有收到 `pong`，主动用 `4000` 关闭连接
6. 非手动关闭时按指数退避重连
7. 组件卸载或手动断开时执行 `disconnect()`，并清理所有 timer

当前重连参数：

- 基础延迟：`1000ms`
- 最大延迟：`30000ms`
- 最大重连次数：`20`

说明：

- 每次重连都会重新申请 ticket，不复用旧 ticket。
- `manualClose = true` 时不会继续自动重连。

## 3. 当前实时端点

前后端当前已经对齐的端点如下：

| 端点 | 前端 owner | 后端入口 | 当前行为 |
| --- | --- | --- | --- |
| `/ws/notifications` | `useNotificationRealtime()` | `code/backend/internal/app/router.go` -> `opsModule.NotificationHandler.ServeWS` | 首屏拉通知后，实时写入通知 store |
| `/ws/contests/:id/announcements` | `useContestAnnouncementRealtime()` | `contestRealtimeHandler.ServeAnnouncementWS` | 收到事件后触发页面重新拉公告 |
| `/ws/contests/:id/scoreboard` | `useContestScoreboardRealtime()` | `contestRealtimeHandler.ServeScoreboardWS` | 收到 `scoreboard.updated` 后触发 HTTP 刷榜 |
| `/ws/contests/:id/awd-preview` | `useContestAwdPreviewRealtime()` | `contestRealtimeHandler.ServeAWDPreviewWS` | 推送 AWD 预览/校验进度 |

## 4. 业务 owner 切分

### 4.1 通知

`useNotificationRealtime()` 当前流程：

1. 先 `getNotifications({ page: 1, page_size: 20 })`
2. 用 `store.setNotifications()` 初始化通知列表
3. 再连接 `notifications` socket
4. `notification.created` -> `store.upsertNotification()`
5. `notification.read` -> `store.markAsRead()`

这样做的边界是：

- 首屏完整数据仍来自 HTTP
- WS 只承担增量同步
- store 只做共享状态容器，不承担建连职责

### 4.2 竞赛公告与排行榜

当前公告和排行榜都采用“轻事件 + HTTP 刷新”的模型：

- 公告实时：`contest.announcement.created`、`contest.announcement.deleted`
- 排行榜实时：`scoreboard.updated`

相关 owner：

- 公告：`code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts`
- 排行榜桥接：`code/frontend/src/components/scoreboard/ScoreboardRealtimeBridge.vue`
- 排行榜详情页：`code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`

说明：

- 排行榜通道当前不下发整榜数据，只发更新事件，再由页面重新请求 `getScoreboard()`
- `ScoreboardDetail` 只有在竞赛状态为 `running` 或 `frozen` 时才启用实时桥接
- `ContestAWDWorkspacePanel.vue` 也会复用同一个 `ScoreboardRealtimeBridge`

### 4.3 AWD 预览

`useContestAwdPreviewRealtime()` 当前服务于平台侧 AWD 配置/预览流程：

- 端点：`contests/${contestId}/awd-preview`
- 事件：`awd.preview.progress`
- payload：当前阶段、尝试次数、明细、错误信息等进度字段

这条链路只负责把进度事件交给调用方，不在 composable 内持有额外业务状态。

## 5. 错误、回退与安全边界

当前回退策略：

- close code 为 `4001` 或 `4401` 时：
  - `authStore.logout()`
  - 跳转 `/401`
- 普通断开时：
  - 进入指数退避重连
- 首次连接失败时：
  - 由各业务 composable 自己给出 warning，例如通知页和排行榜页的“已切换为手动查看/手动刷新”

安全边界：

- ticket 通过 query 传递，因此服务端和网关不能把 `/ws` querystring 原样记进 access log
- 业务 payload 解析放在具体 realtime composable，不在 `useWebSocket()` 里混入业务判断

## 6. Guardrail

- 通知实时同步行为测试：`code/frontend/src/features/notifications/model/useNotificationRealtime.test.ts`
- 页面层不能直接持有复杂路由/业务 owner：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 后端实时端点注册：`code/backend/internal/app/router.go`
