# CTF 靶场平台 API 设计规范

> 版本：v1.0 | 更新日期：2026-03-01

---

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
- `DELETE` 返回 `204 No Content`（无响应体）或 `200`（带确认信息）

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
| `204 No Content` | 删除成功（无响应体） |
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
| `11002` | Access Token 已过期 | 401 |
| `11003` | Refresh Token 已过期 | 401 |
| `11004` | Token 格式无效 | 401 |
| `11005` | Token 已被吊销（黑名单） | 401 |
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

### 4.1 JWT 结构

采用标准 JWT（JSON Web Token）三段式结构：

```
Header.Payload.Signature
```

**Header：**

```json
{
  "alg": "RS256",
  "typ": "JWT"
}
```

> 签名算法说明：
> - 生产环境强制使用 `RS256`（RSA + SHA-256 非对称签名），私钥签发、公钥验证，便于多服务场景下安全分发验证能力
> - 开发环境可降级为 `HS256`（对称签名），通过配置项 `jwt.algorithm` 切换
> - 服务端硬编码允许的 `alg` 白名单 `["RS256", "ES256"]`，拒绝 `none` 和 `HS256`（生产），防止算法混淆攻击

**Payload（Access Token）：**

```json
{
  "sub": "10042",
  "username": "zhangsan",
  "role": "student",
  "iat": 1740787200,
  "exp": 1740794400,
  "jti": "tok_a1b2c3d4"
}
```

| 字段 | 说明 |
|------|------|
| `sub` | 用户 ID（字符串） |
| `username` | 用户名 |
| `role` | 角色：`student` / `teacher` / `admin` |
| `iat` | 签发时间（Unix 时间戳） |
| `exp` | 过期时间（Unix 时间戳） |
| `jti` | Token 唯一标识，用于黑名单吊销 |

**Signature：**

```
RSA-SHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), private_key)
```

> 生产环境使用 RSA 私钥签名，公钥验证。密钥对通过配置文件路径注入：`jwt.private_key_path` / `jwt.public_key_path`。

### 4.2 双 Token 机制

| Token 类型 | 有效期 | 用途 |
|------------|--------|------|
| Access Token | 默认 15 分钟（可配置） | 携带于请求头，用于接口鉴权 |
| Refresh Token | 7 天 | 仅用于刷新 Access Token，存储于 HttpOnly Cookie 或客户端安全存储 |

请求头携带方式：

```
Authorization: Bearer <access_token>
```

### 4.3 Token 刷新流程

```
客户端                          服务端
  │                               │
  │  请求接口（携带 Access Token）  │
  │──────────────────────────────>│
  │  返回 401（Token 过期）        │
  │<──────────────────────────────│
  │                               │
  │  POST /api/v1/auth/refresh    │
  │  Body: { refresh_token }（或从 HttpOnly Cookie 读取） │
  │──────────────────────────────>│
  │                               │── 校验 Refresh Token 有效性
  │                               │── 检查是否在黑名单中
  │                               │── 签发新 Access Token
  │                               │── 轮换 Refresh Token（旧的失效）
  │  返回新 Access + Refresh Token │
  │<──────────────────────────────│
  │                               │
  │  使用新 Access Token 重试请求  │
  │──────────────────────────────>│
```

- Refresh Token 采用**轮换策略**：每次刷新后旧 Refresh Token 立即失效，防止重放攻击
- 若检测到已失效的 Refresh Token 被再次使用，视为 Token 泄露，吊销该用户所有 Token

### 4.4 Token 黑名单（Redis）

用于支持主动登出和强制下线场景：

```
Redis Key:   token:blacklist:{jti}
Value:       1
TTL:         与 Token 剩余有效期一致（避免无限膨胀）
```

**触发黑名单的场景：**

| 场景 | 操作 |
|------|------|
| 用户主动登出 | 将当前 Access Token 的 `jti` 加入黑名单 |
| 修改密码 / 管理员强制下线 / 账户被禁用 | 设置用户维度吊销时间点 `ctf:token:revoked_after:{user_id}=now`（可选：同时将当前 `jti` 加入黑名单） |

**鉴权中间件校验流程：**

1. 解析 JWT，校验签名和过期时间
2. 从 Payload 提取 `jti`，查询 Redis 黑名单
3. 若在黑名单中，返回 `11005 Token 已被吊销`
4. 校验用户维度吊销时间点：若 `iat < revoked_after`，返回 `11005 Token 已被吊销`
5. 校验通过，将用户信息注入请求上下文

---

## 5. 核心 API 列表

> 角色说明：`*` = 所有人（含未登录）、`@` = 已登录用户、`S` = 学员、`T` = 教师、`A` = 管理员

### 5.1 认证模块

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `POST` | `/api/v1/auth/register` | * | 用户注册 |
| `POST` | `/api/v1/auth/login` | * | 用户登录，返回双 Token |
| `POST` | `/api/v1/auth/refresh` | @ | 刷新 Access Token |
| `POST` | `/api/v1/auth/logout` | @ | 登出，Token 加入黑名单 |
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
| `POST` | `/api/v1/challenges/:id/submissions` | S | 提交 Flag |
| `POST` | `/api/v1/challenges/:id/hints/:level/unlock` | S | 解锁指定等级提示 |
| `GET` | `/api/v1/users/me/progress` | S | 我的解题进度 |
| `GET` | `/api/v1/users/me/timeline` | S | 我的解题时间线 |

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
| `POST` | `/api/v1/contests/:id/challenges/:cid/instances` | S | 竞赛中启动靶机实例（校验竞赛状态=running/frozen、用户已报名） |
| `GET` | `/api/v1/contests/:id/challenges` | S | 竞赛题目列表（学员视图，仅 running/frozen 状态可见） |
| `GET` | `/api/v1/contests/:id/my-progress` | S | 我在该竞赛的解题进度 |

### 5.5 技能评估

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/users/me/skill-profile` | S | 我的能力画像（雷达图数据） |
| `GET` | `/api/v1/users/me/recommendations` | S | 个性化推荐靶场 |
| `GET` | `/api/v1/users/:id/skill-profile` | T | 教师查看指定学员能力画像 |
| `GET` | `/api/v1/classes/:id/statistics` | T | 班级整体统计数据 |
| `POST` | `/api/v1/reports/personal` | S,T | 导出个人能力报告（PDF） |
| `POST` | `/api/v1/reports/class` | T | 导出班级报告（PDF） |

### 5.6 系统管理

| 方法 | 路径 | 角色 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/admin/dashboard` | A | 管理仪表盘（概览数据） |
| `GET` | `/api/v1/admin/audit-logs` | A | 审计日志查询 |
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

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 7200,
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
Authorization: Bearer <access_token>
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
Authorization: Bearer <access_token>
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
Authorization: Bearer <access_token>
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
Authorization: Bearer <access_token>
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
| `/ws/scoreboard/:contest_id` | 竞赛排行榜实时推送 | 需要 Token |
| `/ws/notifications` | 用户通知实时推送 | 需要 Token |
| `/ws/contest/:id/announcements` | 竞赛公告实时推送 | 需要 Token |

连接时通过**短期一次性 ticket** 认证（避免 Token 泄露到 URL 日志/代理/Referer）：

```
1. 客户端先调用 POST /api/v1/auth/ws-ticket 获取一次性 ticket（有效期 30s，使用后立即失效）
2. 使用 ticket 建立 WebSocket 连接：ws://host/ws/scoreboard/5?ticket=<one_time_ticket>
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
| Token 刷新 | `rate:refresh:{user_id}` | 每分钟 10 次 | `POST /api/v1/auth/refresh` |

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
