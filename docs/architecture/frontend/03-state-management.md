# 前端状态管理设计

> 对应：01-architecture-overview.md

---

## 1. Store 划分原则

- 仅全局共享状态放入 Pinia Store，页面局部状态用组件 `ref/reactive`
- Store 不直接调用 API，由组件或 composable 调用 API 后写入 Store
- 异常情况：WebSocket 推送的数据可直接写入 Store（通知、排行榜）

---

## 2. Store 定义

### 2.1 auth Store

```ts
// stores/auth.ts
export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)          // { id, username, role, avatar, class_name }
  const accessToken = ref('')

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeacher = computed(() => user.value?.role === 'teacher')
  const isStudent = computed(() => user.value?.role === 'student')

  // 登录成功后调用
  function setAuth(data) { ... }
  // 刷新 Token 后更新
  function updateTokens(access, refresh) { ... }
  // 登出清除
  function logout() { ... }
  // 从 localStorage 恢复
  function restore() { ... }

  return { user, accessToken, isLoggedIn, isAdmin, isTeacher, isStudent, setAuth, updateTokens, logout, restore }
})
```

持久化策略：
- `accessToken` 可存 `localStorage`（便于刷新页面保持登录态）。
- **Refresh Token 必须由后端写入 HttpOnly Cookie（前端不落盘）**。
- `user` 信息每次刷新页面从 `/api/v1/auth/profile` 重新拉取（避免前端落盘敏感信息）。

### 2.2 notification Store

```js
// stores/notification.ts
export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref([])   // 最近通知列表（下拉面板用）
  const unreadCount = ref(0)

  function addNotification(item) { ... }   // WebSocket 推送时调用
  function markAsRead(id) { ... }
  function markAllRead() { ... }

  return { notifications, unreadCount, addNotification, markAsRead, markAllRead }
})
```

### 2.3 contest Store

```js
// stores/contest.ts — 当前正在参与的竞赛状态
export const useContestStore = defineStore('contest', () => {
  const currentContest = ref(null)  // 竞赛基本信息
  const scoreboard = ref([])        // 排行榜数据（WebSocket 驱动）
  const announcements = ref([])     // 公告列表（WebSocket 驱动）
  const isFrozen = ref(false)       // 排行榜是否冻结
  const myTeam = ref(null)          // 我的队伍信息

  function updateScoreboard(data) { ... }
  function addAnnouncement(item) { ... }
  function setFrozen(val) { ... }

  return { currentContest, scoreboard, announcements, isFrozen, myTeam, updateScoreboard, addAnnouncement, setFrozen }
})
```

---

## 3. 数据流向

```
API Response / WebSocket Message
        │
        ▼
  Composable / Component
        │
        ├─ 全局状态 → Pinia Store → 多组件响应式消费
        │
        └─ 局部状态 → 组件内 ref/reactive → 仅当前组件使用
```

不需要 Store 的场景（直接用组件状态）：
- 靶场列表筛选条件、分页状态
- 表单输入值
- Dialog/Drawer 开关状态
- 单个页面的加载状态
