# 前端 WebSocket 与 Composables 设计

> 对应：../backend/04-api-design.md §6 WebSocket 接口

---

## 1. WebSocket 连接管理

### 1.1 连接流程

```
组件挂载 (onMounted)
  │
  ├─ 调用 POST /api/v1/auth/ws-ticket 获取一次性 ticket
  │
  ├─ 建立连接: ws://host/ws/{endpoint}?ticket={ticket}
  │
  ├─ 连接成功 → 启动心跳定时器（30s 间隔）
  │
  ├─ 收到消息 → 按 type 分发到对应处理函数
  │
  ├─ 连接断开 → 指数退避重连（1s → 2s → 4s → ... → 30s 上限）
  │   └─ 每次重连都必须重新获取 ws-ticket（ticket 单次使用且短 TTL，不可复用）
  │
  └─ 组件卸载 (onUnmounted) → 关闭连接 + 清除定时器
```

### 1.2 useWebSocket composable

```ts
// composables/useWebSocket.ts
export function useWebSocket(endpoint, handlers) {
  // 参数:
  //   endpoint: 'scoreboard/5' | 'notifications' | 'contest/5/announcements'
  //   handlers: { 'scoreboard.update': (payload) => {}, ... }
  //
  // 返回:
  //   { status, connect, disconnect, send }
  //
  // 内部逻辑:
  //   - 自动获取 ws-ticket
  //   - 使用 import.meta.env.VITE_WS_BASE_URL 作为基础地址（推荐同源：/ws）
  //   - 心跳: 每 30s 发送 { type: 'ping' }
  //   - 超时: 60s 未收到 pong → 视为断开
  //   - 重连: 指数退避，最多 20 次
  //   - onUnmounted 自动清理
}
```

### 1.3 使用场景

| 端点 | 使用页面 | 消息类型 | 写入目标 |
|------|----------|----------|----------|
| `/ws/scoreboard/:id` | ContestDetail (排行榜 Tab) | `scoreboard.update`, `scoreboard.frozen` | contestStore |
| `/ws/notifications` | AppLayout (全局) | `notification.new` | notificationStore |
| `/ws/contest/:id/announcements` | ContestDetail (公告 Tab) | `announcement.new` | contestStore |

---

## 2. 核心 Composables

### 2.1 useAuth

```js
// 封装认证相关逻辑
// - login(username, password) → 调用 API + 写入 store + 跳转
// - register(data) → 调用 API + 自动登录
// - logout() → 调用 API + 清除 store + 跳转 /login
// - restoreSession() → 从 localStorage 恢复 access token（如有） + 拉取 profile；需要时触发 refresh
```

### 2.2 useToast

```js
// 全局 Toast 通知
// - success(message) → 绿色 Toast，3s 消失
// - error(message) → 红色 Toast，5s 消失
// - warning(message) → 黄色 Toast，4s 消失
// - info(message) → 蓝色 Toast，3s 消失
//
// 实现: provide/inject 模式，AppToast 组件挂载在 App.vue
// 位置: 右上角固定，多条 Toast 垂直堆叠
```

### 2.3 usePagination

```js
// 通用分页逻辑
// 参数: fetchFn(params) → API 调用函数
// 返回: { list, total, page, pageSize, loading, changePage, changePageSize, refresh }
//
// 自动处理:
//   - 页码变化时重新请求
//   - loading 状态管理
//   - 空列表判断
```

### 2.4 useCountdown

```js
// 倒计时（靶机实例剩余时间 / 竞赛剩余时间）
// 参数: expiresAt (ISO 时间字符串)
// 返回: { remaining, formatted, isExpired, isUrgent }
//
// isUrgent: 剩余 < 5 分钟时为 true（用于 text-warning 样式）
// formatted: "01:23:45" 格式
// 每秒更新，onUnmounted 自动清理
```

### 2.5 useClipboard

```js
// 复制到剪贴板
// - copy(text) → navigator.clipboard.writeText + Toast 提示
// - 降级: 不支持 Clipboard API 时使用 execCommand
```

---

## 3. 安全与运维约束（必须）

- ticket 通过 query 参数传递时，**服务端/网关必须避免在 access log 中记录 querystring**（至少对 `/ws` 单独配置 log_format，或关闭该 location 的 access log）。
- 若服务端返回明确的鉴权失败关闭码（例如 4401/4001 等自定义 close code），前端应停止自动重连并触发登出/重新登录流程（避免无意义重连风暴）。
