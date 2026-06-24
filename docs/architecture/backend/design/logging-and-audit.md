# 日志与审计策略

> 状态：Current
> 事实源：`code/backend/internal/platform/logctx/`、各模块 service 代码
> 替代：无

## 本文档范围

| 负责 | 不负责 |
|------|--------|
| 定义日志级别策略（info/warn/error） | 日志采集基础设施（Filebeat、Logstash）→ `docs/operations/` |
| 规定结构化日志字段约定与命名规范 | 日志存储和查询系统（Elasticsearch）→ `docs/operations/` |
| 说明审计日志范围与必填字段 | 日志告警和监控规则 → `docs/operations/` |
| 维护日志轮转和归档策略 | 日志脱敏和合规要求 → `docs/security/` |

## 当前设计

### 组件边界

**本设计负责**：
- 定义日志级别策略（info/warn/error 使用场景）
- 说明结构化日志字段约定和命名规则
- 规定审计日志范围和必填字段
- 记录日志轮转和归档策略

**本设计不负责**：
- 日志采集和分析工具配置（见 `docs/operations/`）
- 具体业务操作的审计内容（见各模块文档）
- 日志存储和备份方案（见 `docs/operations/`）

### 日志框架

**使用库**：`go.uber.org/zap`

**代码位置**：`code/backend/internal/platform/logctx/logctx.go`

#### 核心接口

```go
func Info(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field)
func Warn(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field)
func Error(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field)
```

#### 特性

- **上下文感知**：通过 `context.Context` 传递请求级元数据（如 request_id、user_id）
- **结构化日志**：使用 `zap.Field` 而非字符串拼接
- **零分配优化**：zap 提供高性能日志输出
- **自动字段注入**：logctx 包自动注入 ctx 中的关键字段

### 日志级别策略

#### 1. Info 级别

**定位**：正常业务流程中的关键节点和状态变更。

**适用场景**：
- 用户操作完成（登录成功、提交 Flag、创建竞赛）
- 状态机迁移（竞赛开始、封榜、结束）
- 定时任务执行（状态更新、榜单重建）
- 外部服务调用（容器创建、Flag 分发）

**示例**：
```go
logctx.Info(ctx, s.log, "auth_register_attempt", 
    zap.String("username", req.Username))

logctx.Info(ctx, s.log, "contest_status_transitioned",
    zap.Int64("contest_id", contestID),
    zap.String("from_status", oldStatus),
    zap.String("to_status", newStatus))
```

**原则**：
- 日志消息使用蛇形命名法（`snake_case`）
- 动词使用过去时态（如 `created`、`updated`、`failed`）
- 每个关键操作只记录一次 Info 日志（避免日志洪流）

#### 2. Warn 级别

**定位**：非预期但不影响业务流程继续的异常情况。

**适用场景**：
- 业务规则校验失败（用户名已存在、Flag 错误、权限不足）
- 可重试的外部依赖失败（缓存未命中、弱事件发布失败）
- 配置缺失或降级运行（EventBus 未配置、缓存禁用）
- 数据不一致但有兜底机制（缓存与数据库不一致）

**示例**：
```go
logctx.Warn(ctx, s.log, "auth_register_failed_username_exists", 
    zap.String("username", req.Username))

s.log.Warn("publish_contest_event_failed", 
    zap.String("event", evt.Name), 
    zap.Error(err))
```

**原则**：
- Warn 日志应包含失败原因（通过 message 或 error 字段）
- 不应频繁出现 Warn 日志（如每次正常校验失败都记录）
- Warn 不阻塞业务流程，调用方应继续执行或降级

#### 3. Error 级别

**定位**：影响业务流程的错误，需要人工介入或告警。

**适用场景**：
- 数据库操作失败（插入 / 更新 / 删除失败）
- 不可重试的外部依赖失败（容器创建失败、Docker 连接断开）
- 系统级错误（文件读写失败、内存不足、panic 恢复）
- 关键业务逻辑错误（状态机非法迁移、数据完整性破坏）

**示例**：
```go
logctx.Error(ctx, s.log, "auth_register_failed", 
    zap.String("username", req.Username), 
    zap.Error(err))

logctx.Error(ctx, s.log, "contest_status_update_db_failed",
    zap.Int64("contest_id", contestID),
    zap.String("status", status),
    zap.Error(err))
```

**原则**：
- Error 日志必须包含 `zap.Error(err)`，保留完整错误链
- 应在错误发生的最外层记录（避免同一错误被多次记录）
- Error 日志应触发监控告警或进入错误追踪系统

#### 4. Debug 和 Fatal 级别

**Debug**：
- 当前后端代码中不使用 Debug 级别
- 开发调试信息通过单元测试或临时 Info 日志输出

**Fatal**：
- 仅用于应用启动失败场景（数据库连接失败、配置加载失败）
- Fatal 会触发 `os.Exit(1)`，不应在业务逻辑中使用

### 结构化日志字段约定

#### 1. 标准字段命名

| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| `user_id` | `int64` | 操作者用户 ID | `zap.Int64("user_id", userID)` |
| `username` | `string` | 用户名（登录 / 注册场景） | `zap.String("username", "alice")` |
| `contest_id` | `int64` | 竞赛 ID | `zap.Int64("contest_id", 123)` |
| `challenge_id` | `int64` | 题目 ID | `zap.Int64("challenge_id", 456)` |
| `submission_id` | `int64` | 提交 ID | `zap.Int64("submission_id", 789)` |
| `team_id` | `int64` | 队伍 ID | `zap.Int64("team_id", 101)` |
| `instance_id` | `string` | 容器实例 ID | `zap.String("instance_id", "cnt_abc123")` |
| `event` | `string` | 事件名 | `zap.String("event", "contest.submission.created")` |
| `status` | `string` | 状态值 | `zap.String("status", "running")` |
| `error` | `error` | 错误对象 | `zap.Error(err)` |
| `duration_ms` | `int64` | 耗时（毫秒） | `zap.Int64("duration_ms", 123)` |

#### 2. 命名规则

- 使用蛇形命名法（`snake_case`）
- ID 字段使用 `<entity>_id` 格式
- 布尔字段使用 `is_` 或 `has_` 前缀（如 `is_correct`、`has_permission`）
- 时间字段使用 `<action>_at` 格式（如 `created_at`、`submitted_at`）

#### 3. 禁止模式

- 禁止使用字符串拼接：`logger.Info(fmt.Sprintf("user %s login", username))`
- 禁止使用任意 key：`zap.Any("data", obj)` 应改为具体字段
- 禁止记录敏感信息：密码、Token、完整 Flag、邮箱（应脱敏）

### 审计日志范围

#### 1. 必须记录的操作

**用户身份相关**：
- 用户注册（成功 / 失败）
- 用户登录（成功 / 失败）
- 密码修改
- 角色变更

**代码位置**：
- `code/backend/internal/module/auth/application/commands/service.go`

**示例**：
```go
logctx.Info(ctx, s.log, "auth_register_attempt", 
    zap.String("username", req.Username))

logctx.Warn(ctx, s.log, "auth_login_failed", 
    zap.String("username", req.Username),
    zap.String("reason", "invalid_password"))
```

**权限变更相关**：
- 权限授予 / 撤销
- 角色分配 / 移除
- 管理员操作（创建竞赛、发布题目、修改配置）

**敏感操作相关**：
- 竞赛状态变更（开始 / 封榜 / 结束）
- Flag 分发和验证
- 容器实例创建 / 销毁
- 批量数据导入 / 导出

#### 2. 审计日志必填字段

| 字段 | 说明 |
|------|------|
| `msg` | 操作描述（如 `auth_login_success`） |
| `user_id` | 操作者 ID（系统操作可省略） |
| `target_id` | 目标实体 ID（如 `contest_id`、`user_id`） |
| `timestamp` | 操作时间（由 zap 自动注入） |
| `result` | 操作结果（通过 log level 和 `error` 字段体现） |

#### 3. 审计日志示例

```go
// 成功操作
logctx.Info(ctx, s.log, "contest_created",
    zap.Int64("user_id", userID),
    zap.Int64("contest_id", contestID),
    zap.String("contest_name", name))

// 失败操作
logctx.Warn(ctx, s.log, "contest_update_permission_denied",
    zap.Int64("user_id", userID),
    zap.Int64("contest_id", contestID),
    zap.String("required_role", "admin"))

// 系统操作
logctx.Info(ctx, s.log, "contest_status_auto_transitioned",
    zap.Int64("contest_id", contestID),
    zap.String("from_status", oldStatus),
    zap.String("to_status", newStatus),
    zap.String("trigger", "time_window"))
```

### 日志轮转和归档策略

#### 1. 日志输出

**开发环境**：
- 输出到 stdout（便于容器日志采集）
- 格式：JSON（便于日志系统解析）

**生产环境**：
- 输出到 stdout + 文件（双重保障）
- 文件路径：`/var/log/ctf-api/app.log`
- 格式：JSON

#### 2. 轮转策略

**文件轮转**：
- 按日期轮转：每天生成新文件（`app-2026-06-24.log`）
- 按大小轮转：单文件超过 100MB 时分割
- 保留策略：本地保留最近 7 天日志

**实现方式**：
- 使用 `lumberjack` 库（`gopkg.in/natefinch/lumberjack.v2`）
- 配置在应用启动时初始化

**示例配置**：
```go
logger := &lumberjack.Logger{
    Filename:   "/var/log/ctf-api/app.log",
    MaxSize:    100, // MB
    MaxBackups: 7,   // 保留最近 7 个备份
    MaxAge:     7,   // 保留最近 7 天
    Compress:   true, // 压缩旧日志
}
```

#### 3. 归档策略

**短期归档（7-30 天）**：
- 存储位置：本地磁盘或对象存储
- 用途：问题排查、审计回溯

**长期归档（30 天以上）**：
- 存储位置：冷存储（S3 Glacier、阿里云 OSS 归档）
- 用途：合规审计、历史分析

**归档范围**：
- 审计日志：永久保留
- 业务日志：保留 30 天
- 调试日志：保留 7 天

## 边界

### 本文档负责

- 定义日志级别策略和使用场景
- 规定结构化日志字段命名约定
- 说明审计日志范围和必填字段
- 维护日志轮转和归档策略

### 本文档不负责

- 日志采集基础设施（Filebeat、Logstash）→ `docs/operations/`
- 日志存储和查询系统（Elasticsearch、Kibana）→ `docs/operations/`
- 日志告警和监控规则（Prometheus、Grafana）→ `docs/operations/`
- 日志脱敏和合规要求（GDPR、等保）→ `docs/security/`

## Guardrail

### 日志级别检查

- Review 时需确认：
  - Info 日志是否为关键业务节点
  - Warn 日志是否真正非预期但不阻塞流程
  - Error 日志是否包含 `zap.Error(err)`

### 结构化日志检查

- 禁止字符串拼接：使用 `zap.String` 而非 `fmt.Sprintf`
- 禁止记录敏感信息：密码、Token、完整 Flag
- 字段命名是否符合蛇形命名法

### 审计日志完整性

- 用户身份操作是否记录审计日志
- 权限变更操作是否记录操作者和目标
- 敏感操作是否包含必填字段（user_id、target_id、result）

### 代码位置检查

- 是否使用 `logctx.Info/Warn/Error` 而非直接调用 `logger.Info`
- 是否传递 `context.Context` 以注入请求级元数据

## 已知限制

### 1. 日志级别过于粗粒度

- 当前只有 Info/Warn/Error 三级
- 缺乏 Debug 级别（开发调试不便）
- 建议未来引入可配置的日志级别（开发环境启用 Debug）

### 2. 审计日志缺乏统一入口

- 当前审计日志散落在各模块 service 中
- 缺乏统一的审计日志接口和格式
- 建议未来引入 `platform/audit` 包，提供 `audit.Log(ctx, action, target)` 接口

### 3. 敏感信息脱敏不完整

- 当前只有人工 review 保证，缺乏自动脱敏
- 邮箱、IP 地址等 PII 可能泄漏到日志
- 建议引入 `zap.Field` 包装器，自动脱敏敏感字段

### 4. 日志轮转配置分散

- 日志轮转策略在应用启动代码中硬编码
- 缺乏统一配置文件或环境变量
- 建议迁移到 `config` 包统一管理

### 5. 缺乏日志采样

- 高频操作（如 Flag 验证）可能产生日志洪流
- 当前无采样或限流机制
- 建议引入日志采样（如每 1000 次记录一次）

### 6. 日志查询能力不足

- 当前只能通过文件 grep 或 Elasticsearch 手动查询
- 缺乏日志关联分析（如同一请求的所有日志）
- 建议引入 request_id 或 trace_id，串联请求全链路日志
