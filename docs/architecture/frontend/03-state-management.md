# 前端状态管理设计

> 状态：Current
> 事实源：`code/frontend/src/stores/`、`code/frontend/src/features/**/model/`、`code/frontend/src/composables/use*.ts`
> 替代：无

## 定位

本文档只说明前端状态应该放在哪一层，以及当前跨页面共享状态、页面级状态和实时状态的 owner。

- 覆盖：Pinia store、页面级 feature model、通用 composable、登录态恢复、通知和竞赛共享状态。
- 不覆盖：具体页面的视觉结构、组件样式和构建配置。

## 当前设计

- `code/frontend/src/stores/auth.ts`
  - 负责：维护当前用户快照、`sessionRestored` 标记、登录态派生值和基于 session cookie 的恢复流程
  - 不负责：持久化 access token、refresh token，或在浏览器里长期保存认证状态

- `code/frontend/src/stores/notification.ts`
  - 负责：维护通知列表、未读数、去重排序和读取状态更新
  - 不负责：决定何时拉首屏通知或何时建立实时连接；这些由 `useNotificationRealtime()` 驱动

- `code/frontend/src/stores/contest.ts`
  - 负责：承接竞赛级共享状态，如当前竞赛摘要、排行榜、公告和冻结态
  - 不负责：承接所有竞赛页面的局部表单、筛选条件或页面 loading 状态

- `code/frontend/src/features/**/model`、`code/frontend/src/composables/use*.ts`
  - 负责：页面级异步加载、路由 query 同步、导出流、分页、重试、回调桥接和一次性派生状态，例如 `useChallengeDetailPage.ts`、`useScoreboardDetailPage.ts`、`useTeacherStudentAnalysisPage.ts`
  - 不负责：把真正跨页面共享的会话状态重新复制回某个局部 composable

## 1. 状态归属规则

当前前端按下面的 owner 规则分层：

| 状态类型 | 当前 owner | 说明 |
| --- | --- | --- |
| 登录用户、角色、是否已恢复 session | `stores/auth.ts` | 跨页共享，且守卫依赖 |
| 通知列表、未读数 | `stores/notification.ts` | 顶栏、列表页和详情页共享 |
| 竞赛摘要、公告、排行榜冻结态 | `stores/contest.ts` | 竞赛相关页面共享 |
| 列表分页、筛选、表单草稿、局部 loading | `features/**/model` 或 `composables/use*.ts` | 跟随页面生命周期销毁 |
| 路由 query tab、导出任务轮询、删除确认流 | `composables/useRouteQueryTabs.ts`、`useReportStatusPolling.ts`、`useDestructiveConfirm.ts` 等 | 以“页面能力域”切分 |

判断原则：

- 只有跨页面共享、且多个组件需要同时读写的状态，才进入 Pinia。
- 页面内一次性流程，不进 store，直接由 feature model 持有。
- route view 只组合状态，不直接变成“大型页面控制器”。

## 2. 认证状态

`code/frontend/src/stores/auth.ts` 当前模型：

- 核心状态：
  - `user`
  - `sessionRestored`
- 核心派生：
  - `isLoggedIn`
  - `isAdmin`
  - `isTeacher`
  - `isStudent`

主流程：

1. 守卫或页面需要确认登录态时调用 `restore()`
2. `restore()` 先清理历史 `localStorage` token 残留
3. 若当前还未恢复过，则通过 `getProfile()` 读取 `/auth/profile`
4. 成功后写入 `user`，失败则保持 `user = null`
5. 无论成功失败，都把 `sessionRestored` 置为 `true`

不变量：

- 认证依赖 HttpOnly session cookie，前端只保留用户快照的内存态。
- `restorePromise` 保证同一时刻只跑一条恢复链路，避免重复请求。
- 登出时只清理本地快照和历史 token，不尝试在 store 层做导航判断。

## 3. 通知状态

`code/frontend/src/stores/notification.ts` 当前模型：

- `notifications`
- `unreadCount`
- `setNotifications()`
- `upsertNotification()`
- `markAsRead()`
- `markAllRead()`

当前行为：

1. `useNotificationRealtime()` 先调用 `getNotifications({ page: 1, page_size: 20 })`
2. 首屏通知通过 `setNotifications()` 去重并按 `created_at` 倒序写入
3. 实时事件 `notification.created` 进入 `upsertNotification()`
4. 实时事件 `notification.read` 或用户操作进入 `markAsRead()`

不变量：

- `unreadCount` 是 computed，不单独维护第二份计数状态。
- store 内部统一把通知 `id` 归一成字符串。
- `upsertNotification()` 只保留最近 20 条，避免顶栏实时列表无限增长。

## 4. 竞赛共享状态

`code/frontend/src/stores/contest.ts` 当前只保留轻量共享状态：

| 字段 | 用途 |
| --- | --- |
| `currentContest` | 当前竞赛摘要 |
| `scoreboard` | 当前共享排行榜数据 |
| `announcements` | 当前共享公告数据 |
| `isFrozen` | 排行榜是否冻结 |
| `myTeam` | 当前队伍信息 |

边界：

- store 只做共享数据容器，不直接编排页面请求。
- `ScoreboardDetail` 页面的完整数据流仍由 `useScoreboardDetailPage.ts` 持有。
- 竞赛公告实时连接、排行榜实时连接由 feature realtime composable 触发，不在 store 自己开 socket。

## 5. 页面级状态与通用 composable

当前项目已经把大量页面控制逻辑下沉到 `features/**/model`：

- 题目详情：`code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- 排行榜详情：`code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- 学生仪表盘：`code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts`
- 教师学生分析：`code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`

通用页面能力则放在 `code/frontend/src/composables/use*.ts`，例如：

- `usePagination.ts`：分页请求、取消旧请求、页码切换
- `useRouteQueryTabs.ts`、`useUrlSyncedTabs.ts`：query 同步和 tab 切换
- `useReportStatusPolling.ts`：导出/生成任务轮询
- `useToast.ts`：全局反馈

这样做的直接目的是：

- route view 只保留页面壳和组合入口
- API 调用链集中在 feature model，而不是分散到 `.vue` 模板里
- 页面销毁后局部状态跟着释放，不污染全局 store

## 6. 数据流与副作用

当前主链路：

1. route view 进入页面
2. view 调用 feature model
3. feature model 调 API 模块
4. 需要跨页共享时写入 store
5. 需要局部渲染时保留在当前 composable 的 `ref` / `computed`

实时链路：

1. `useWebSocket()` 建连
2. feature realtime composable 解析事件
3. 写 store 或触发页面回调刷新

边界约束：

- store 不直接持有路由对象
- view 不直接 import 非 contract API 模块
- 低层 UI 不直接依赖 store、router 和业务 API

## 7. 兼容与历史例外

- `auth` store 仍会清理历史 `ctf_access_token` / `ctf_refresh_token` localStorage 键，属于从旧 token 模式迁移到 session cookie 的兼容收口。
- 当前没有通用“把所有页面状态都持久化到 store”的机制；这是刻意保留的边界，不是缺失实现。

## 8. Guardrail

- 前端分层与 import 边界：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- route view 不直接拥有业务 API 和路由 query hook：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 通知实时状态同步：`code/frontend/src/features/notifications/model/useNotificationRealtime.test.ts`
