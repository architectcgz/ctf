# CTF 靶场平台 API 设计规范

> 状态：Current
> 事实源：`docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`、`code/backend/internal/app/router_routes.go`
> 替代：无

## 定位

本文档说明当前后端 HTTP / WebSocket 契约的事实入口、路由装配边界和统一响应口径。

- 负责：描述当前接口命名空间、Envelope、鉴权票据、核心模块路由以及契约事实源的优先级。
- 不负责：替代 OpenAPI、handler 实现或前端页面文案成为唯一真相；也不把旧示例中的过时字段自动视为当前契约。

## 当前设计

- `docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`
  - 负责：定义当前 HTTP / WebSocket 契约、Envelope、分页结构、AWD service/readiness/checker preview、challenge import/self-check、teacher review 等字段事实；OpenAPI 是机器可读主契约，说明文档补充示例与非 OpenAPI 细节
  - 不负责：承载 handler 内部实现、权限判断细节或运行时补偿逻辑；这些仍以后端代码为准

- `code/backend/internal/app/router.go`、`code/backend/internal/app/router_routes.go`
  - 负责：装配 `/api/v1/*` 路由、模块 handler、中间件、审计入口和运行时代理路径，把 `auth`、`challenge`、`practice`、`contest`、`assessment`、`ops` 的对外命名空间固定下来；当前 `/instances*`、AWD target proxy 和 defense SSH 相关路径统一通过 `InstanceModule.Handler` 挂载，而 `challenge`、`contest`、`ops` 依赖的容器能力则通过 `ContainerRuntimeModule` 装配，但外部 URL 与响应契约保持不变
  - 不负责：在 composition 层实现业务状态机或直接拼装 DTO 字段

- `code/backend/internal/middleware/request_id.go`、`auth.go`、`audit.go`、`recovery.go`
  - 负责：补齐 `request_id`、认证上下文、审计记录和统一错误恢复，使 `code / message / data / request_id` 的响应口径能稳定落到 handler 链路
  - 不负责：替代模块 application service 的业务错误语义，或绕开模块 owner 直接写协议分支

- `code/backend/internal/module/*/api/http/*.go`、`code/backend/internal/module/auth/infrastructure/token_service.go`、`code/backend/internal/app/composition/ops_module.go`
  - 负责：在各模块内完成参数绑定、DTO 映射、Session / WebSocket ticket、实例访问 proxy ticket 和通知推送接入
  - 不负责：让前端直接依赖底层 Redis / Docker / SQL 结构，或回退到无 Envelope、无 request id 的散装接口风格

## 接口或数据影响

- 当前统一响应结构以 `ApiEnvelopeBase` 为根，核心字段是 `code`、`message`、`data`、`request_id`；见 `docs/contracts/openapi-v1.yaml`。
- 当前认证仍以服务端 Session + Cookie 为主，并通过 `GET /api/v1/ws-ticket`、实例访问 proxy ticket、AWD 攻击 / 防守 ticket 承接 WebSocket 与运行时访问。
- 当前重要 API 面包括 challenge import / commit / self-check、附件下载、实例访问、竞赛状态 / 榜单、AWD service 管理 / preview / readiness / round / attack log，以及教师复盘 / 报告导出链路；如果示例与 `openapi-v1.yaml` 冲突，以契约文档优先。
- 2026-05-11 的 runtime-instance phase 2 / slice 2 没有新增、删除或重命名外部 HTTP 路径；变化仅在内部装配层：实例访问、教师实例列表、AWD target proxy 与 defense SSH 路由由 `InstanceModule` 承接，`challenge`、`contest`、`ops` 依赖的容器能力由 `ContainerRuntimeModule` 承接；`RuntimeModule` 仅保留过渡兼容别名。
- 2026-05-12 的 challenge-ops boundary phase 3 / slice 2 没有新增、删除或重命名外部 HTTP / WebSocket 契约；题目发布自检完成后的教师通知仍通过既有 `/api/v1/notifications` 与 `/ws/notifications` 暴露，内部改为由 `challenge.publish_check_finished` 事件从 `challenge` 交给 `ops` 消费。
- 2026-05-12 的 contest-ops boundary phase 3 / slice 3 同样没有新增、删除或重命名外部 HTTP / WebSocket 契约；`/ws/contests/:id/announcements`、`/ws/contests/:id/scoreboard`、`/ws/contests/:id/awd-preview` 保持现有消息类型与 payload，内部改为由 `contest` 发布 realtime 事件、`ops` relay 到既有 WebSocket manager。
- 2026-05-12 的 teacher overview aggregate 新增 `GET /api/v1/teacher/overview`；该接口由 `teaching_query` 输出 summary-scope 聚合结果，服务 `/academy/overview`，不再让前端通过默认班级去拼 `classes/:name/*` 的 detail DTO。
- 2026-05-12 的 practice 查询收口 phase 4 / slice 1 没有新增、删除或重命名外部 HTTP 契约；`GET /api/v1/users/me/progress` 与 `GET /api/v1/users/me/timeline` 保持原路径与响应结构，内部 owner 已收口回 `practice/api/http` 与 `practice/application/queries`。
- 2026-05-12 的 teaching query overview query surface phase 4 / slice 2 同样没有新增、删除或重命名外部 HTTP 契约；`GET /api/v1/teacher/overview` 保持原路径与响应结构，内部 owner 从宽 `teaching_query/application/queries.Service` 收口到独立 `OverviewService`，handler 也改为分别依赖 overview 与其余教师查询。
- 2026-05-12 的 teaching query class insight query surface phase 4 / slice 3 同样没有新增、删除或重命名外部 HTTP 契约；`GET /api/v1/teacher/classes/:name/{summary,trend,review}` 保持原路径与响应结构，内部 owner 从剩余宽 `teaching_query/application/queries.Service` 收口到独立 `ClassInsightService`，handler 也改为分别依赖目录/学生查询、overview 和 class insight。
- 2026-05-13 的 teaching query student review query surface phase 4 / slice 4 同样没有新增、删除或重命名外部 HTTP 契约；`GET /api/v1/teacher/students/:id/{progress,recommendations,timeline,evidence,attack-sessions}` 保持原路径、响应结构和权限口径不变，内部 owner 从剩余宽 `teaching_query/application/queries.Service` 收口到独立 `StudentReviewService`，handler 改为分别依赖目录查询、overview、class insight 和 student review 四类 query owner。
- 2026-05-13 的 auth CAS validator phase 5 / slice 16 同样没有新增、删除或重命名外部 HTTP 契约；`GET /api/v1/auth/cas/status`、`GET /api/v1/auth/cas/login`、`GET /api/v1/auth/cas/callback` 保持原路径、响应结构和错误码口径不变，内部仅把 CAS ticket 校验从 auth command service 直接持有的 HTTP client 收口到 `auth/ports.CASTicketValidator` 与 `auth/infrastructure/cas_ticket_validator.go`。
- 当前 AWD 学生侧运行时 HTTP 面只保留 `POST /api/v1/contests/:id/awd/services/:sid/defense/ssh`；不存在 `defense/files`、`defense/directories`、`defense/commands` 路由，runtime HTTP facade 也不再为这组已下线路由保留 service interface。

## Guardrail

- 路由装配与全链路 HTTP 验证：`code/backend/internal/app/router_test.go`、`code/backend/internal/app/full_router_integration_test.go`
- runtime access facade 与 retired defense workbench 约束：`code/backend/internal/app/composition/architecture_test.go`、`code/backend/internal/module/runtime/architecture_test.go`
- 状态矩阵与实例访问链路：`code/backend/internal/app/full_router_state_matrix_integration_test.go`
- Auth / Session 契约：`code/backend/internal/module/auth/api/http/http_integration_test.go`
- 通知与 WebSocket 契约：`code/backend/internal/module/ops/api/http/notification_http_integration_test.go`
- 题目导入与附件 / 自检契约：`code/backend/internal/app/challenge_import_integration_test.go`

## 历史迁移

- 当前接口事实已经收口为“契约文档 + router 装配 + 模块 handler”三层结构，不再依赖单篇 API 说明文档独自维护所有字段。
- 下文保留的示例、错误码段落和列表页说明仍可作为详细参考；若与 `docs/contracts/openapi-v1.yaml` 或当前 handler 行为冲突，以契约和代码为准。

## 1. API 设计规范

### 1.1 URL 命名规则

- 采用 RESTful 风格，资源名使用**复数名词**、**kebab-case**
- 所有接口统一前缀 `/api/v1/`
- 路径层级不超过 4 层，避免过深嵌套

```
✅ GET  /api/v1/challenges
✅ GET  /api/v1/challenges/:id/hints
✅ POST /api/v1/contests/:id/teams/:tid/join

❌ GET  /api/v1/getChallenge        （动词命名）
❌ GET  /api/v1/challenge            （单数）
❌ GET  /api/v1/Challenge_List       （大写 + 下划线）
```

### 1.2 HTTP 方法语义

| 方法 | 语义 | 幂等 | 示例 |
|------|------|------|------|
| `GET` | 查询资源，不产生副作用 | 是 | `GET /api/v1/challenges` |
| `POST` | 创建资源 / 执行操作 | 否 | `POST /api/v1/challenges/:id/submissions` |
| `PUT` | 全量更新资源 | 是 | `PUT /api/v1/admin/challenges/:id` |
| `PATCH` | 部分更新资源 | 是 | `PATCH /api/v1/admin/challenges/:id` |
| `DELETE` | 删除资源 | 是 | `DELETE /api/v1/instances/:id` |

- `POST` 用于非幂等操作（创建、提交 Flag、启动实例等）
- `PUT` 与 `PATCH` 的区别：`PUT` 需要传完整字段，`PATCH` 只传需要修改的字段
- `DELETE` 统一返回 Envelope（含 `data=null`），不使用 `204 No Content`（与前端统一解析逻辑一致）

### 1.3 版本策略

- 采用 **URL 路径版本化**：`/api/v1/...`
- 当发生不兼容变更时递增版本号（v2），旧版本保留至少 **2 个迭代周期**
- 小版本兼容性变更（新增字段、新增可选参数）不递增版本号

### 1.4 分页规范

请求参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码，从 1 开始 |
| `page_size` | int | 20 | 每页条数，最大 100 |

示例：`GET /api/v1/challenges?page=2&page_size=10`

### 1.5 排序规范

- 使用 `sort` 查询参数，格式：`字段名:排序方向`
- 多字段排序用逗号分隔
- 排序方向：`asc`（升序）、`desc`（降序）

```
GET /api/v1/challenges?sort=created_at:desc
GET /api/v1/challenges?sort=difficulty:asc,created_at:desc
```

### 1.6 筛选规范

- 筛选条件通过查询参数传递，参数名与资源字段名一致
- 支持精确匹配和范围查询

```
GET /api/v1/challenges?category=web&difficulty=medium
GET /api/v1/challenges?tag=sql-injection&status=active
GET /api/v1/admin/audit-logs?start_time=2026-01-01T00:00:00Z&end_time=2026-02-01T00:00:00Z
```

- 模糊搜索使用 `keyword` 参数：`GET /api/v1/challenges?keyword=注入`

---

## 2. 统一响应格式

> **接口契约声明（强制）**：除本设计规范的通用约束外，所有与前端联调相关的“接口返回字段/类型/分页结构/WS 消息 payload”以 `ctf/docs/contracts/api-contract-v1.md` 为准；若本文示例与契约文档不一致，以契约文档优先。

### 2.1 基础结构

所有接口响应均遵循以下 JSON 结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "request_id": "req_a1b2c3d4e5f6"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 业务状态码，`0` 表示成功，非零表示错误 |
| `message` | string | 状态描述，成功时为 `"success"`，失败时为错误摘要 |
| `data` | object/array/null | 业务数据，错误时为 `null` |
| `request_id` | string | 请求唯一标识，用于日志追踪和问题排查 |

### 2.2 成功响应

**单个资源：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 42,
    "title": "SQL 注入基础",
    "category": "web",
    "difficulty": "easy"
  },
  "request_id": "req_a1b2c3d4e5f6"
}
```

### 2.3 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      { "id": 1, "title": "SQL 注入基础" },
      { "id": 2, "title": "XSS 反射型攻击" }
    ],
    "total": 56,
    "page": 1,
    "page_size": 20
  },
  "request_id": "req_b2c3d4e5f6a1"
}
```

### 2.4 错误响应

**通用错误：**

```json
{
  "code": 10001,
  "message": "请求参数错误",
  "data": null,
  "request_id": "req_c3d4e5f6a1b2"
}
```

**字段级校验错误（含 `errors` 数组）：**

```json
{
  "code": 10001,
  "message": "请求参数校验失败",
  "data": null,
  "errors": [
    { "field": "username", "message": "用户名长度需在 3-20 个字符之间" },
    { "field": "password", "message": "密码必须包含大小写字母和数字" }
  ],
  "request_id": "req_d4e5f6a1b2c3"
}
```

### 2.5 HTTP 状态码使用

| HTTP 状态码 | 使用场景 |
|-------------|----------|
| `200 OK` | 查询成功、更新成功 |
| `201 Created` | 资源创建成功 |
| `204 No Content` | 不使用（平台统一返回 Envelope，含 `request_id`；删除成功返回 `200` + `data=null`） |
| `400 Bad Request` | 参数校验失败 |
| `401 Unauthorized` | 未认证 / Token 无效 |
| `403 Forbidden` | 无权限访问 |
| `404 Not Found` | 资源不存在 |
| `409 Conflict` | 资源冲突（如重复报名） |
| `429 Too Many Requests` | 请求频率超限 |
| `500 Internal Server Error` | 服务器内部错误 |

---

## 3. 错误码体系

错误码采用 **5 位整数**，按模块划分区间，便于快速定位问题来源。

### 3.1 通用错误码（10000-10999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `10000` | 未知错误 | 500 |
| `10001` | 请求参数错误 | 400 |
| `10002` | 请求参数校验失败（含字段级错误） | 400 |
| `10003` | 未认证，请先登录 | 401 |
| `10004` | 无权限访问该资源 | 403 |
| `10005` | 请求的资源不存在 | 404 |
| `10006` | 请求方法不允许 | 405 |
| `10007` | 资源冲突 | 409 |
| `10008` | 请求频率超限，请稍后重试 | 429 |
| `10009` | 服务器内部错误 | 500 |
| `10010` | 服务暂时不可用 | 503 |

### 3.2 认证模块（11000-11999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `11001` | 用户名或密码错误 | 401 |
| `11002` | 会话已失效或不存在 | 401 |
| `11003` | 会话已过期 | 401 |
| `11004` | Session Cookie 格式无效 | 401 |
| `11005` | 会话已被吊销 | 401 |
| `11006` | 账户已被锁定 | 403 |
| `11007` | 账户已被禁用 | 403 |
| `11008` | 用户名已存在 | 409 |
| `11009` | 邮箱已被注册 | 409 |
| `11010` | 登录失败次数过多，账户临时锁定 | 429 |

### 3.3 靶场模块（12000-12999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `12001` | 镜像构建失败 | 500 |
| `12002` | 镜像拉取超时 | 504 |
| `12003` | 靶机未配置 Flag | 400 |
| `12004` | 靶机已下线，无法操作 | 410 |
| `12005` | 靶场配置不完整 | 400 |
| `12006` | 拓扑结构校验失败 | 400 |
| `12007` | 靶场分类不存在 | 404 |
| `12008` | 靶场标签数量超限 | 400 |

### 3.4 演练模块（13000-13999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `13001` | 实例启动失败 | 500 |
| `13002` | 实例数量已达上限 | 429 |
| `13003` | Flag 错误 | 400 |
| `13004` | Flag 提交已被锁定（频率限制） | 429 |
| `13005` | 实例已过期，请重新启动 | 410 |
| `13006` | 实例延时次数已达上限 | 400 |
| `13007` | 该靶场已解出，无法重复提交 | 409 |
| `13008` | 提示已全部解锁 | 400 |
| `13009` | 实例正在启动中，请稍候 | 409 |

### 3.5 竞赛模块（14000-14999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `14001` | 竞赛尚未开始 | 403 |
| `14002` | 竞赛已结束 | 403 |
| `14003` | 队伍人数已满 | 400 |
| `14004` | 已报名该竞赛，不可重复报名 | 409 |
| `14005` | 未报名该竞赛 | 403 |
| `14006` | 竞赛报名已截止 | 403 |
| `14007` | 已加入其他队伍，不可重复加入 | 409 |
| `14008` | 队伍邀请码无效 | 400 |
| `14009` | 竞赛题目尚未发布 | 403 |

### 3.6 评估模块（15000-15999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `15001` | 无解题数据，无法生成能力画像 | 400 |
| `15002` | 报告生成失败 | 500 |
| `15003` | 报告正在生成中，请稍候 | 409 |
| `15004` | 班级无学员数据 | 400 |
| `15005` | 评估维度配置缺失 | 500 |

### 3.7 容器模块（16000-16999）

| 错误码 | 含义 | HTTP 状态码 |
|--------|------|-------------|
| `16001` | 容器启动超时 | 504 |
| `16002` | 集群资源不足，无法分配容器 | 503 |
| `16003` | 容器网络创建失败 | 500 |
| `16004` | 容器镜像不存在 | 404 |
| `16005` | 容器端口分配冲突 | 500 |
| `16006` | 容器健康检查未通过 | 500 |
| `16007` | 容器已被销毁，无法操作 | 410 |

---

## 4. 认证方案

### 4.1 服务端 Session 结构

认证采用 Redis 持久化的 opaque session。浏览器只保留 `HttpOnly + Secure + SameSite` session cookie，不暴露 access token，也不提供 refresh 接口。

**服务端 Session 记录：**

```json
{
  "id": "sess_b1c2d3e4",
  "user_id": 10042,
  "username": "zhangsan",
  "role": "student",
  "expires_at": "2026-04-30T12:00:00Z"
}
```

| 字段 | 说明 |
|------|------|
| `id` | Session ID，随机 opaque 标识 |
| `user_id` | 用户 ID |
| `username` | 用户名快照 |
| `role` | 角色：`student` / `teacher` / `admin` |
| `expires_at` | 会话过期时间 |

### 4.2 会话 Cookie 机制

| 项目 | 默认值 | 用途 |
|------|--------|------|
| Session TTL | 7 天 | 控制 Redis session 记录与 cookie 的有效期 |
| Session Cookie | `ctf_session` | 浏览器自动随请求携带，用于接口鉴权 |

受保护请求的携带方式：

```
Cookie: ctf_session=<session_id>
```

### 4.3 会话恢复与失效流程

```
客户端                          服务端
  │                               │
  │  请求接口（浏览器自动携带 session cookie） │
  │──────────────────────────────>│
  │                               │── 从 cookie 读取 session_id
  │                               │── 查询 Redis session
  │                               │── 校验会话是否存在/未过期
  │  返回业务数据或 401            │
  │<──────────────────────────────│
  │                               │
  │  POST /api/v1/auth/logout     │
  │──────────────────────────────>│
  │                               │── 删除 Redis session
  │                               │── Clear-Cookie: ctf_session
  │  返回 200                     │
  │<──────────────────────────────│
```

- 登录成功或注册成功后，服务端通过 `Set-Cookie` 写入 session cookie。
- 页面刷新后的登录恢复依赖 `/api/v1/auth/profile`，不再走 `/api/v1/auth/refresh`。

### 4.4 Session 存储（Redis）

用于支持主动登出和强制下线场景：

```
Redis Key:   ctf:auth:session:{session_id}
Value:       session json
TTL:         与 session_ttl 对齐
```

**触发会话失效的场景：**

| 场景 | 操作 |
|------|------|
| 用户主动登出 | 删除当前 session，并清除 cookie |
| 修改密码 / 管理员强制下线 / 账户被禁用 | 删除对应 session 记录，使后续请求统一返回 401 |

**鉴权中间件校验流程：**

1. 从 Cookie 读取 session ID
2. 查询 Redis session 记录并校验 TTL
3. 若会话不存在或已过期，返回 `11002/11003`
4. 校验通过，将用户信息注入请求上下文

---

## 5. 核心 API 列表

> 角色说明：`*` = 所有人（含未登录）、`@` = 已登录用户、`S` = 学员、`T` = 教师、`A` = 管理员

### 5.1 认证模块

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `POST` | `/api/v1/auth/register` | * | 用户注册 |
| `POST` | `/api/v1/auth/login` | * | 用户登录：返回用户信息，并通过 HttpOnly Cookie 写入 session |
| `POST` | `/api/v1/auth/logout` | @ | 登出，删除当前 session 并清除 cookie |
| `GET` | `/api/v1/auth/profile` | @ | 获取当前用户信息 |
| `PUT` | `/api/v1/auth/password` | @ | 修改密码 |
| `POST` | `/api/v1/auth/ws-ticket` | @ | 获取 WebSocket 一次性 ticket（TTL 30s） |

### 5.1.1 用户管理（管理员）

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/admin/users` | A | 用户列表（分页、按状态/班级/角色筛选） |
| `POST` | `/api/v1/admin/users` | A | 创建用户（管理员手动添加） |
| `GET` | `/api/v1/admin/users/:id` | A | 用户详情（含角色、解题统计） |
| `PUT` | `/api/v1/admin/users/:id` | A | 更新用户信息（状态、班级、角色等） |
| `DELETE` | `/api/v1/admin/users/:id` | A | 删除用户（软删除） |
| `POST` | `/api/v1/admin/users/import` | A | 批量导入用户（CSV/Excel，含学号、姓名、班级） |
| `PUT` | `/api/v1/admin/users/:id/status` | A | 变更用户状态（启用/禁用/解锁） |
| `PUT` | `/api/v1/admin/users/:id/roles` | A | 分配用户角色 |
| `PUT` | `/api/v1/admin/users/:id/reset-password` | A | 重置用户密码 |
| `GET` | `/api/v1/admin/classes` | A | 班级列表（去重聚合） |

### 5.1.2 教师功能

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/teacher/classes` | T | 教师管辖的班级列表 |
| `GET` | `/api/v1/teacher/classes/:name/students` | T | 班级学员列表（含解题进度） |
| `GET` | `/api/v1/teacher/students/:id/progress` | T | 查看指定学员解题详情 |
| `GET` | `/api/v1/teacher/challenges` | T | 靶场列表（教师视图，可查看 Writeup） |
| `POST` | `/api/v1/teacher/challenges` | T | 创建靶场（教师出题，需管理员审核） |
| `PUT` | `/api/v1/teacher/challenges/:id` | T | 更新自己创建的靶场 |
| `GET` | `/api/v1/teacher/contests` | T | 教师管理的竞赛列表 |
| `POST` | `/api/v1/teacher/contests` | T | 创建竞赛（教师发起，需管理员审核或自动通过） |
| `PUT` | `/api/v1/teacher/contests/:id` | T | 更新自己创建的竞赛 |
| `GET` | `/api/v1/teacher/contests/:id/submissions` | T | 查看竞赛提交记录 |
| `GET` | `/api/v1/teacher/contests/:id/scoreboard` | T | 查看竞赛排行榜（含冻结期间真实排名） |

### 5.2 靶场管理（管理员）

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/admin/images` | A | 镜像列表 |
| `POST` | `/api/v1/admin/images` | A | 创建/上传镜像 |
| `PUT` | `/api/v1/admin/images/:id` | A | 更新镜像信息 |
| `DELETE` | `/api/v1/admin/images/:id` | A | 删除镜像 |
| `GET` | `/api/v1/admin/challenges` | A | 靶场列表（管理视图） |
| `POST` | `/api/v1/admin/challenges` | A | 创建靶场 |
| `GET` | `/api/v1/admin/challenges/:id` | A | 靶场详情（含配置） |
| `PUT` | `/api/v1/admin/challenges/:id` | A | 更新靶场 |
| `DELETE` | `/api/v1/admin/challenges/:id` | A | 删除靶场 |
| `GET` | `/api/v1/admin/challenges/:id/hints` | A | 靶场提示列表 |
| `POST` | `/api/v1/admin/challenges/:id/hints` | A | 添加提示 |
| `PUT` | `/api/v1/admin/challenges/:id/hints/:hid` | A | 更新提示 |
| `DELETE` | `/api/v1/admin/challenges/:id/hints/:hid` | A | 删除提示 |
| `GET` | `/api/v1/admin/challenges/:id/writeup` | A | 获取 Writeup |
| `PUT` | `/api/v1/admin/challenges/:id/writeup` | A | 创建/更新 Writeup |
| `DELETE` | `/api/v1/admin/challenges/:id/writeup` | A | 删除 Writeup |
| `POST` | `/api/v1/challenges/:id/writeup-submissions` | U | 学员保存/提交自己的 writeup |
| `GET` | `/api/v1/challenges/:id/writeup-submissions/me` | U | 获取当前学员自己的 writeup |
| `GET` | `/api/v1/teacher/writeup-submissions` | T | 教师按学生/班级/题目筛选 writeup |
| `GET` | `/api/v1/teacher/writeup-submissions/:id` | T | 教师查看单条 writeup 详情 |
| `PUT` | `/api/v1/teacher/writeup-submissions/:id/review` | T | 教师评阅 writeup |
| `GET` | `/api/v1/admin/tags` | A | 标签列表 |
| `POST` | `/api/v1/admin/tags` | A | 创建标签 |
| `PUT` | `/api/v1/admin/tags/:id` | A | 更新标签 |
| `DELETE` | `/api/v1/admin/tags/:id` | A | 删除标签 |
| `GET` | `/api/v1/admin/topologies` | A | 拓扑列表 |
| `POST` | `/api/v1/admin/topologies` | A | 创建拓扑 |
| `PUT` | `/api/v1/admin/topologies/:id` | A | 更新拓扑 |
| `DELETE` | `/api/v1/admin/topologies/:id` | A | 删除拓扑 |
| `POST` | `/api/v1/admin/challenges/import` | A | 批量导入靶场（ZIP 包，含 Dockerfile + 元数据 JSON） |
| `PUT` | `/api/v1/admin/challenges/:id/resource-limits` | A | 配置靶机资源限制（CPU/内存/磁盘/PID/带宽） |

### 5.3 攻防演练（学员）

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/challenges` | S | 靶场列表（支持分页、筛选、排序） |
| `GET` | `/api/v1/challenges/:id` | S | 靶场详情（含已解锁提示） |
| `POST` | `/api/v1/challenges/:id/instances` | S | 启动靶机实例 |
| `DELETE` | `/api/v1/instances/:id` | S | 销毁靶机实例 |
| `POST` | `/api/v1/instances/:id/extend` | S | 延长实例有效期 |
| `GET` | `/api/v1/instances` | S | 我的实例列表 |
| `GET` | `/api/v1/instances/:id/access` | S | 获取实例访问入口并签发 proxy ticket |
| `POST` | `/api/v1/challenges/:id/submissions` | S | 提交 Flag |
| `POST` | `/api/v1/challenges/:id/hints/:level/unlock` | S | 解锁指定等级提示 |
| `GET` | `/api/v1/users/me/progress` | S | 我的解题进度 |
| `GET` | `/api/v1/users/me/timeline` | S | 我的解题时间线 |

> 说明：
> - `POST /api/v1/challenges/:id/instances` 负责启动或复用实例，不签发 `proxy ticket`。
> - `GET /api/v1/instances/:id/access` 在访问前校验权限，并签发短时 `proxy ticket`。
> - `proxy ticket` 仅用于平台代理访问和共享实例访问上下文传递，不是最终提交凭证。
> - 所有题目统一通过提交流程提交 `flag`。

### 5.4 竞赛管理

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/admin/contests` | A | 竞赛列表（管理视图） |
| `POST` | `/api/v1/admin/contests` | A | 创建竞赛 |
| `GET` | `/api/v1/admin/contests/:id` | A | 竞赛详情（含配置） |
| `PUT` | `/api/v1/admin/contests/:id` | A | 更新竞赛 |
| `DELETE` | `/api/v1/admin/contests/:id` | A | 删除竞赛 |
| `GET` | `/api/v1/admin/contests/:id/challenges` | A | 竞赛题目列表（管理视图，含分值配置） |
| `POST` | `/api/v1/admin/contests/:id/challenges` | A | 添加题目到竞赛（关联已有靶场，可设竞赛专属分值） |
| `PUT` | `/api/v1/admin/contests/:id/challenges/:cid` | A | 更新竞赛题目配置（分值、排序、是否启用） |
| `DELETE` | `/api/v1/admin/contests/:id/challenges/:cid` | A | 从竞赛移除题目 |
| `POST` | `/api/v1/admin/contests/:id/challenges/batch` | A | 批量添加题目到竞赛 |
| `PUT` | `/api/v1/admin/contests/:id/status` | A | 管理员手动变更竞赛状态（发布/取消/紧急结束） |
| `POST` | `/api/v1/admin/contests/:id/announcements` | A | 发布竞赛公告 |
| `GET` | `/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/summary` | A | 获取 AWD 轮次代理流量摘要（趋势、热点路径、攻击方/受害方排行） |
| `GET` | `/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/events` | A | 获取 AWD 轮次代理流量明细（支持按队伍、题目、状态组筛选） |
| `GET` | `/api/v1/contests` | @ | 竞赛列表（公开视图） |
| `GET` | `/api/v1/contests/:id` | @ | 竞赛详情 |
| `POST` | `/api/v1/contests/:id/register` | S | 报名竞赛 |
| `GET` | `/api/v1/contests/:id/teams` | S | 队伍列表 |
| `POST` | `/api/v1/contests/:id/teams` | S | 创建队伍 |
| `PUT` | `/api/v1/contests/:id/teams/:tid` | S | 更新队伍信息（队长） |
| `DELETE` | `/api/v1/contests/:id/teams/:tid` | S | 解散队伍（队长） |
| `POST` | `/api/v1/contests/:id/teams/:tid/join` | S | 加入队伍 |
| `GET` | `/api/v1/contests/:id/scoreboard` | @ | 竞赛排行榜 |
| `GET` | `/api/v1/contests/:id/announcements` | @ | 竞赛公告列表 |
| `POST` | `/api/v1/contests/:id/challenges/:cid/submissions` | S | 竞赛中提交 Flag |
| `POST` | `/api/v1/contests/:id/challenges/:cid/instances` | S | 非 AWD 竞赛中启动靶机实例（校验竞赛状态=running/frozen、用户已报名） |
| `POST` | `/api/v1/contests/:id/awd/services/:sid/instances` | S | AWD 竞赛中按 `service_id` 启动队伍共享实例（校验竞赛状态=running/frozen、用户已报名且已入队） |
| `GET` | `/api/v1/contests/:id/challenges` | S | 竞赛题目列表（学员视图，仅 running/frozen 状态可见） |
| `GET` | `/api/v1/contests/:id/my-progress` | S | 我在该竞赛的解题进度 |

> AWD 流量监控说明：
> - 上述 `traffic/*` 接口基于 `awd_traffic_events` 事实表返回平台代理链路下的共享实例访问摘要。
> - 该能力用于展示“攻击相关流量态势”，不直接等价于攻击成功判定；成功攻破仍以 `awd_attack_logs` 为准。

### 5.5 技能评估

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/users/me/skill-profile` | S | 我的能力画像（雷达图数据） |
| `GET` | `/api/v1/users/me/recommendations` | S | 个性化推荐靶场 |
| `GET` | `/api/v1/users/:id/skill-profile` | T | 教师查看指定学员能力画像 |
| `GET` | `/api/v1/teacher/students/:id/evidence` | T,A | 教师/管理员查看学员攻防证据链与复盘摘要 |
| `GET` | `/api/v1/classes/:id/statistics` | T | 班级整体统计数据 |
| `POST` | `/api/v1/reports/personal` | S,T | 导出个人能力报告（PDF） |
| `POST` | `/api/v1/reports/class` | T | 导出班级报告（PDF） |
| `POST` | `/api/v1/teacher/students/:id/review-archive/export` | T,A | 导出指定学员复盘归档（JSON） |

### 5.6 系统管理

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/admin/dashboard` | A | 管理仪表盘（概览数据） |
| `GET` | `/api/v1/admin/audit-logs` | A | 审计日志查询 |
| `POST` | `/api/v1/admin/contests/:id/export` | A | 导出赛事结果归档（JSON） |
| `GET` | `/api/v1/notifications` | @ | 我的通知列表 |
| `PUT` | `/api/v1/notifications/:id/read` | @ | 标记通知为已读 |

### 5.7 核心接口详细示例

#### 5.7.1 用户登录

**请求：**

```
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "username": "zhangsan",
  "password": "P@ssw0rd123"
}
```

**成功响应（200）：**

```
Set-Cookie: ctf_session=<session_id>; HttpOnly; Secure; SameSite=Lax; Path=/api/v1; Max-Age=604800
```

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user": {
      "id": 10042,
      "username": "zhangsan",
      "role": "student",
      "avatar": "https://cdn.example.com/avatars/10042.png"
    }
  },
  "request_id": "req_f1a2b3c4d5e6"
}
```

**失败响应（401）：**

```json
{
  "code": 11001,
  "message": "用户名或密码错误",
  "data": null,
  "request_id": "req_a2b3c4d5e6f1"
}
```

#### 5.7.2 获取靶场列表（分页 + 筛选 + 排序）

**请求：**

```
GET /api/v1/challenges?page=1&page_size=10&category=web&difficulty=easy&sort=created_at:desc
Cookie: ctf_session=<session_id>
```

**成功响应（200）：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 101,
        "title": "SQL 注入基础",
        "category": "web",
        "difficulty": "easy",
        "tags": ["sql-injection", "owasp-top10"],
        "solved_count": 128,
        "total_attempts": 356,
        "is_solved": false,
        "points": 100,
        "created_at": "2026-02-15T10:30:00Z"
      },
      {
        "id": 102,
        "title": "XSS 反射型攻击",
        "category": "web",
        "difficulty": "easy",
        "tags": ["xss", "owasp-top10"],
        "solved_count": 95,
        "total_attempts": 210,
        "is_solved": true,
        "points": 100,
        "created_at": "2026-02-10T08:00:00Z"
      }
    ],
    "total": 23,
    "page": 1,
    "page_size": 10
  },
  "request_id": "req_b3c4d5e6f1a2"
}
```

#### 5.7.3 启动靶机实例

**请求：**

```
POST /api/v1/challenges/101/instances
Cookie: ctf_session=<session_id>
```

**成功响应（201）：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "inst_x7y8z9",
    "challenge_id": 101,
    "status": "running",
    "access_url": "http://10.10.1.42:8080",
    "ssh_info": {
      "host": "10.10.1.42",
      "port": 22,
      "username": "ctf"
    },
    "flag_type": "dynamic",
    "expires_at": "2026-03-01T04:44:58Z",
    "remaining_extends": 2,
    "created_at": "2026-03-01T02:44:58Z"
  },
  "request_id": "req_c4d5e6f1a2b3"
}
```

**失败响应（429 实例数超限）：**

```json
{
  "code": 13002,
  "message": "实例数量已达上限，请先销毁已有实例",
  "data": null,
  "request_id": "req_d5e6f1a2b3c4"
}
```

#### 5.7.4 提交 Flag

**请求：**

```
POST /api/v1/challenges/101/submissions
Cookie: ctf_session=<session_id>
Content-Type: application/json
```

```json
{
  "flag": "flag{sql_1nj3ct10n_b4s1c_2026}"
}
```

**成功响应（200 正确）：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "correct": true,
    "points_earned": 100,
    "first_blood": false,
    "solved_at": "2026-03-01T03:15:22Z",
    "challenge_progress": {
      "total_challenges": 56,
      "solved_challenges": 12
    }
  },
  "request_id": "req_e6f1a2b3c4d5"
}
```

**失败响应（400 Flag 错误）：**

```json
{
  "code": 13003,
  "message": "Flag 错误，请重试",
  "data": {
    "correct": false,
    "remaining_attempts": 17
  },
  "request_id": "req_f1a2b3c4d5e6"
}
```

#### 5.7.5 竞赛排行榜

**请求：**

```
GET /api/v1/contests/5/scoreboard?page=1&page_size=20
Cookie: ctf_session=<session_id>
```

**成功响应（200）：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "contest": {
      "id": 5,
      "title": "2026 春季校园 CTF 挑战赛",
      "status": "running",
      "started_at": "2026-03-01T09:00:00Z",
      "ends_at": "2026-03-01T21:00:00Z"
    },
    "scoreboard": {
      "list": [
        {
          "rank": 1,
          "team_id": 12,
          "team_name": "Binary Wizards",
          "score": 2450,
          "solved_count": 8,
          "last_submission_at": "2026-03-01T11:32:15Z"
        },
        {
          "rank": 2,
          "team_id": 7,
          "team_name": "Null Pointers",
          "score": 2100,
          "solved_count": 7,
          "last_submission_at": "2026-03-01T11:45:03Z"
        }
      ],
      "total": 32,
      "page": 1,
      "page_size": 20
    },
    "frozen": false
  },
  "request_id": "req_g2h3i4j5k6l7"
}
```

---

## 6. WebSocket 接口设计

### 6.1 连接端点

| 端点 | 用途 | 认证 |
|------|------|------|
| `/ws/notifications` | 用户通知实时推送 | 需要 Token |
| `/ws/contests/:id/announcements` | 竞赛公告实时推送 | 需要一次性 ticket |
| `/ws/contests/:id/scoreboard` | 竞赛排行榜实时推送 | 需要一次性 ticket |

连接时通过**短期一次性 ticket** 认证（避免 Token 泄露到 URL 日志/代理/Referer）：

```
1. 客户端先调用 POST /api/v1/auth/ws-ticket 获取一次性 ticket（有效期 30s，使用后立即失效）
2. 使用 ticket 建立 WebSocket 连接：ws://host/ws/contests/5/scoreboard?ticket=<one_time_ticket>
3. 服务端验证 ticket 有效性后建立连接，立即删除 ticket
```

> 安全说明：
> - ticket 存储于 Redis，TTL 30s，使用后立即 DEL，防止重放
> - ticket 与用户 ID 绑定，不可跨用户使用
> - 备选方案：连接建立后通过首条消息发送 Access Token 认证，但 ticket 方案实现更简洁

### 6.2 消息格式

所有 WebSocket 消息采用 JSON 格式，统一结构：

```json
{
  "type": "消息类型",
  "payload": {},
  "timestamp": "2026-03-01T11:32:15Z"
}
```

**消息类型定义：**

| `type` | 端点 | payload | 说明 |
|--------|------|---------|------|
| `system.connected` | 全部 | `{ user_id, heartbeat_interval_seconds, retry }` | 连接建立成功后的系统握手消息 |
| `contest.announcement.created` | `/ws/contests/:id/announcements` | `{ contest_id, announcement }` | 新公告创建后广播，前端可直接插入或重新拉取 |
| `contest.announcement.deleted` | `/ws/contests/:id/announcements` | `{ contest_id, announcement_id }` | 公告删除后广播 |
| `scoreboard.updated` | `/ws/contests/:id/scoreboard` | `{ contest_id }` | 榜单发生变化，前端收到后调用现有 HTTP 排行榜接口刷新 |
| `notification.created` | `/ws/notifications` | `NotificationItem` | 用户私有通知新增 |
| `notification.read` | `/ws/notifications` | `NotificationItem` | 用户私有通知已读状态变更 |

| type | 方向 | 说明 |
|------|------|------|
| `ping` | 客户端 → 服务端 | 心跳探测 |
| `pong` | 服务端 → 客户端 | 心跳响应 |
| `scoreboard.update` | 服务端 → 客户端 | 排行榜数据更新 |
| `scoreboard.frozen` | 服务端 → 客户端 | 排行榜冻结通知 |
| `notification.new` | 服务端 → 客户端 | 新通知推送 |
| `announcement.new` | 服务端 → 客户端 | 新竞赛公告推送 |
| `instance.status` | 服务端 → 客户端 | 靶机实例状态变更 |

**排行榜更新消息示例：**

```json
{
  "type": "scoreboard.update",
  "payload": {
    "rank": 1,
    "team_id": 12,
    "team_name": "Binary Wizards",
    "score": 2550,
    "solved_count": 9,
    "last_submission_at": "2026-03-01T12:05:33Z"
  },
  "timestamp": "2026-03-01T12:05:34Z"
}
```

**通知推送消息示例：**

```json
{
  "type": "notification.new",
  "payload": {
    "id": 1024,
    "title": "靶机实例即将过期",
    "content": "您的靶机「SQL 注入基础」将在 10 分钟后过期，请及时延时或提交 Flag",
    "level": "warning",
    "created_at": "2026-03-01T02:34:58Z"
  },
  "timestamp": "2026-03-01T02:34:58Z"
}
```

### 6.3 心跳机制

- 客户端每 **30 秒** 发送一次 `ping` 消息
- 服务端收到 `ping` 后立即回复 `pong`
- 若服务端 **90 秒** 内未收到客户端心跳，主动断开连接
- 若客户端 **60 秒** 内未收到 `pong` 响应，视为连接断开，触发重连

```json
// 客户端发送
{ "type": "ping", "payload": {}, "timestamp": "2026-03-01T12:00:30Z" }

// 服务端响应
{ "type": "pong", "payload": {}, "timestamp": "2026-03-01T12:00:30Z" }
```

### 6.4 重连策略

采用**指数退避**（Exponential Backoff）策略：

| 重连次数 | 等待时间 | 说明 |
|----------|----------|------|
| 第 1 次 | 1 秒 | 立即重试 |
| 第 2 次 | 2 秒 | |
| 第 3 次 | 4 秒 | |
| 第 4 次 | 8 秒 | |
| 第 5 次 | 16 秒 | |
| 第 6 次及以后 | 30 秒 | 最大退避上限 |

- 最大重连次数：**20 次**
- 超过最大次数后停止重连，提示用户手动刷新页面
- 每次成功连接后重置重连计数器
- 重连时携带 `last_event_id` 参数，服务端补发断线期间的消息（若支持）

---

## 7. 限流策略

基于 Redis 实现滑动窗口限流，通过 Gin 中间件统一拦截。

### 7.1 限流维度

| 维度 | 限流键 | 限制 | 适用范围 |
|------|--------|------|----------|
| 全局（IP） | `rate:ip:{ip}` | 每分钟 120 次 | 所有接口 |
| 登录 | `rate:login:{ip}` | 每分钟 10 次 | `POST /api/v1/auth/login` |
| 注册 | `rate:register:{ip}` | 每分钟 5 次 | `POST /api/v1/auth/register` |
| Flag 提交 | `rate:flag:{user_id}:{challenge_id}` | 每分钟 20 次 | Flag 提交接口 |
| 实例启动 | `rate:instance:{user_id}` | 每分钟 5 次 | 实例创建接口 |

### 7.2 限流响应头

每个请求的响应中携带限流相关 Header，便于客户端感知：

| Header | 说明 |
|--------|------|
| `X-RateLimit-Limit` | 当前窗口允许的最大请求数 |
| `X-RateLimit-Remaining` | 当前窗口剩余可用请求数 |
| `X-RateLimit-Reset` | 当前窗口重置时间（Unix 时间戳） |
| `Retry-After` | 触发限流时，建议客户端等待的秒数 |

### 7.3 限流触发响应

当请求触发限流时，返回 HTTP `429 Too Many Requests`：

```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 20
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1740798360
Retry-After: 45
Content-Type: application/json
```

```json
{
  "code": 10008,
  "message": "请求频率超限，请稍后重试",
  "data": null,
  "request_id": "req_h3i4j5k6l7m8"
}
```
