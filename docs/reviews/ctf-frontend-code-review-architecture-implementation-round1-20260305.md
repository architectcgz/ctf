# CTF 前端代码 Review（架构文档一致性专项 第 1 轮）

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | 前端架构文档一致性（`docs/architecture/frontend` vs `code/frontend/src`） |
| 轮次 | 第 1 轮 |
| 审查范围 | `docs/architecture/frontend/01~08`，`code/frontend/src` 路由/守卫/API/Store/Composables/主要视图 |
| 审查基准 | `docs/architecture/frontend/*.md` + `docs/contracts/openapi-v1.yaml` |
| 审查日期 | 2026-03-05 |
| 审查方式 | 低层实现审查（权限、安全、接口契约、状态流、可构建性、回归风险） |

## 问题清单

### 🔴 高优先级

- [H1] 路由鉴权与角色校验被注释，受保护页面可被未登录/越权访问
  - 文件：`code/frontend/src/router/guards.ts:68-85`
  - 问题描述：`requiresAuth` 校验与 `meta.roles` 校验被整段注释，仅保留 profile 加载逻辑，绝大多数受保护页面会直接放行。
  - 影响范围/风险：存在前端越权访问风险；若后端也有疏漏将升级为实际权限绕过。
  - 修正建议：立即恢复鉴权守卫与角色守卫（未登录跳 `/login?redirect=...`，越权跳 `/403` 或 `/dashboard`）。

- [H2] 竞赛加入队伍接口路径与契约不一致
  - 文件：`code/frontend/src/api/contest.ts:53-58`
  - 对照：`docs/contracts/openapi-v1.yaml:1588-1608`、`docs/architecture/frontend/04-api-layer.md:149-151`
  - 问题描述：前端调用 `POST /contests/{id}/join`，契约定义为 `POST /contests/{id}/teams/{tid}/join`。
  - 影响范围/风险：联调阶段大概率 404/405，竞赛组队流程不可用。
  - 修正建议：统一 API 契约与实现（前端改 URL+参数，或同步更新 OpenAPI 与架构文档）。

- [H3] 当前分支 TypeScript 不可通过，构建链路阻塞
  - 文件：
    - `code/frontend/src/views/admin/ChallengeManage.vue:191`
    - `code/frontend/src/views/challenges/ChallengeDetail.vue:28`
    - `code/frontend/src/views/challenges/ChallengeList.vue:76`
    - `code/frontend/src/views/contests/ContestList.vue:90`
    - `code/frontend/src/views/profile/SkillProfile.vue:239`
  - 问题描述：存在缺失类型导入、字段与契约不匹配、错误枚举值、ECharts formatter 类型不兼容等问题。
  - 影响范围/风险：`vue-tsc` 失败，CI/发布阻塞；同时反映出契约漂移。
  - 修正建议：先清零 typecheck 错误，再继续功能迭代。

### 🟡 中优先级

- [M1] WebSocket 实现与架构文档约束不一致
  - 文件：`code/frontend/src/composables/useWebSocket.ts`
  - 对照：`docs/architecture/frontend/05-websocket-composables.md:16-47`
  - 问题描述：文档要求 `?ticket=` 连接、60s pong 超时与鉴权 close code 特殊处理；实现为连接后发 `auth` 消息，且未实现 pong 超时/close code 停止重连逻辑。
  - 影响范围/风险：网关协议不一致时连接失败；鉴权失败可能触发无意义重连。
  - 修正建议：以单一协议为准（建议优先以 OpenAPI/网关实际实现），同步修正文档与前端。

- [M2] 核心页面与数据流未按文档落地，功能完成度与文档描述不一致
  - 文件：
    - `code/frontend/src/views/contests/ContestDetail.vue`
    - `code/frontend/src/views/scoreboard/ScoreboardView.vue`
    - `code/frontend/src/views/notifications/NotificationList.vue`
    - `code/frontend/src/views/dashboard/DashboardView.vue`
    - `code/frontend/src/views/profile/UserProfile.vue`
    - `code/frontend/src/views/teacher/*`
    - `code/frontend/src/views/admin/AdminDashboard.vue`
    - `code/frontend/src/views/admin/AuditLog.vue`
    - `code/frontend/src/views/admin/CheatDetection.vue`
  - 问题描述：多个页面仍是占位或 mock，未接入文档声明的 API/WS 流程。
  - 影响范围/风险：验收预期与实际交付偏差较大，联调与演示风险高。
  - 修正建议：补充“已实现/未实现矩阵”并分批落地优先级页面。

- [M3] 通知下拉只改本地 Store，不回写后端已读状态
  - 文件：`code/frontend/src/components/layout/NotificationDropdown.vue:71-77`
  - 问题描述：`markAsRead/markAllRead` 仅调用 store，未调用 `api/notification.ts`。
  - 影响范围/风险：刷新后状态回滚，前后端未读数不一致。
  - 修正建议：接入 `markAsRead` API，批量已读可按后端能力补接口或逐条调用。

- [M4] 实例页空状态跳转到不存在路由
  - 文件：`code/frontend/src/views/instances/InstanceList.vue:78`
  - 问题描述：空状态跳转 `/student/challenges`，实际路由为 `/challenges`。
  - 影响范围/风险：用户无法通过空状态入口回到挑战列表。
  - 修正建议：改为 `/challenges`。

- [M5] 错误处理兜底文案不准确
  - 文件：`code/frontend/src/api/request.ts:173-181`
  - 对照：`docs/architecture/frontend/04-api-layer.md:221`
  - 问题描述：未命中 `errorMap` 的 4xx/5xx 统一提示“网络连接失败”。
  - 影响范围/风险：服务端业务错误被误报为网络问题，影响排障与用户理解。
  - 修正建议：区分网络错误与业务错误，HTTP 错误走“通用失败文案 + request_id”。

### 🟢 低优先级

- [L1] 设计系统文档路径引用不一致
  - 文件：`docs/architecture/frontend/01-architecture-overview.md:6`
  - 问题描述：写的是 `design-system/MASTER.md`，仓库实际路径为 `design-system/ctf-platform/MASTER.md`。
  - 影响范围/风险：新成员按文档定位会失败。
  - 修正建议：修正文档链接。

- [L2] `useToast()` 在模块顶层调用导致测试注入告警
  - 文件：`code/frontend/src/api/request.ts:30`
  - 问题描述：非组件上下文调用 `inject` 在测试日志中出现警告。
  - 影响范围/风险：日志噪声，影响问题定位效率。
  - 修正建议：延迟获取 toast（在拦截器执行时获取）或改为独立消息通道。

## 运行/测试结果

- 命令：`npm run typecheck`
  - 结果：失败（5 个 TS 错误）
- 命令：`npm run test:run`
  - 结果：失败（2 个失败用例：`ContestDetail.test.ts`、`InstanceList.test.ts`）
- 命令：`npm run lint`
  - 结果：失败（ESLint 10 未找到 `eslint.config.*`）

## 风险点与回滚点

- 风险点
  - 鉴权校验失效引发越权访问风险（最高优先级）
  - 接口契约漂移导致关键流程（组队）不可用
  - `typecheck/test/lint` 失败导致交付与回归验证链路不完整
  - 文档宣称功能与实际实现偏差，带来验收与联调风险

- 回滚点
  - 本轮仅新增 review 文档，无业务代码改动，若需回滚仅删除本文件即可。

## 变更摘要（按文件）

- 新增：`docs/reviews/ctf-frontend-code-review-architecture-implementation-round1-20260305.md`
