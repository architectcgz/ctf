# CTF 平台代码 Review（audit-log 第 1 轮）：审计日志功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | audit-log |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit ebed353，4 个核心文件，249 行新增 |
| 变更概述 | 实现审计日志功能，记录用户登录、镜像管理、靶场管理、Flag 提交等关键操作 |
| 审查基准 | CLAUDE.md（CTF 平台开发规范） |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] Model 和 DTO 未分离，违反架构规范

- **文件**：
  - `code/backend/internal/module/audit/service.go:12`
  - `code/backend/internal/module/audit/handler.go:42`
- **问题描述**：
  - Service 接口 `ListLogs` 直接返回 `[]*model.AuditLog`，违反了"Service 层必须返回 DTO"的规范
  - Handler 直接将 `model.AuditLog` 通过 `response.Page()` 返回给前端，导致数据库模型直接暴露到 API 层
- **影响范围/风险**：
  - 违反分层架构原则，Model 和 API 耦合
  - 未来如果 Model 增加敏感字段（如操作详情、请求体等），会直接泄漏到 API
  - 无法灵活控制 API 响应格式（如字段重命名、格式转换、关联数据加载）
- **修正建议**：
  1. 在 `internal/dto/` 下创建 `audit.go`，定义 DTO：
  ```go
  package dto

  import "time"

  type AuditLogResp struct {
      ID        int64     `json:"id"`
      UserID    int64     `json:"user_id"`
      Username  string    `json:"username"`  // 关联用户名
      Action    string    `json:"action"`
      Resource  string    `json:"resource"`
      IP        string    `json:"ip"`
      CreatedAt time.Time `json:"created_at"`
  }
  ```

  2. 修改 Service 接口返回 DTO：
  ```go
  type Service interface {
      Log(userID int64, action, resource, ip string) error
      ListLogs(ctx context.Context, filter ListFilter) ([]*dto.AuditLogResp, int64, error)
  }
  ```

  3. 在 Service 实现中添加 Model → DTO 转换函数：
  ```go
  func (s *service) ListLogs(ctx context.Context, filter ListFilter) ([]*dto.AuditLogResp, int64, error) {
      logs, total, err := s.repo.List(ctx, filter)
      if err != nil {
          return nil, 0, err
      }

      // 转换为 DTO
      resp := make([]*dto.AuditLogResp, len(logs))
      for i, log := range logs {
          resp[i] = toAuditLogResp(log)
      }
      return resp, total, nil
  }

  func toAuditLogResp(log *model.AuditLog) *dto.AuditLogResp {
      return &dto.AuditLogResp{
          ID:        log.ID,
          UserID:    log.UserID,
          Action:    log.Action,
          Resource:  log.Resource,
          IP:        log.IP,
          CreatedAt: log.CreatedAt,
      }
  }
  ```

#### [H2] 审计中间件在登录接口上的位置错误

- **文件**：`code/backend/internal/app/router.go:80`
- **问题描述**：
  ```go
  authGroup.POST("/login", middleware.Audit(auditService, model.AuditActionLogin, "auth/login"), authHandler.Login)
  ```
  审计中间件放在 handler 之前，但中间件通过 `c.Next()` 后才记录，此时 `authctx.MustCurrentUser(c)` 获取的是登录前的用户（UserID=0），导致登录审计无法记录正确的用户 ID
- **影响范围/风险**：
  - 登录操作的审计日志无法记录用户 ID（因为登录前 context 中没有用户信息）
  - 登录审计功能完全失效
- **修正建议**：
  登录审计需要特殊处理，有两种方案：

  **方案 1（推荐）**：在 `authHandler.Login` 内部调用审计服务
  ```go
  // 在 auth/handler.go 的 Login 方法中
  func (h *Handler) Login(c *gin.Context) {
      // ... 登录逻辑 ...

      // 登录成功后记录审计日志
      if err := h.auditService.Log(user.ID, model.AuditActionLogin, "auth/login", c.ClientIP()); err != nil {
          h.log.Error("failed to log audit", zap.Error(err))
      }

      response.Success(c, loginResp)
  }
  ```

  **方案 2**：创建专门的登录审计中间件
  ```go
  func AuditLogin(logger AuditLogger) gin.HandlerFunc {
      return func(c *gin.Context) {
          c.Next()

          if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
              // 从响应中提取用户 ID（需要在 Login handler 中设置）
              if userID, exists := c.Get("login_user_id"); exists {
                  _ = logger.Log(userID.(int64), model.AuditActionLogin, "auth/login", c.ClientIP())
              }
          }
      }
  }
  ```

#### [H3] 中间件忽略审计日志记录失败

- **文件**：`code/backend/internal/middleware/audit.go:24`
- **问题描述**：
  ```go
  _ = logger.Log(user.UserID, action, resource, ip)
  ```
  使用 `_` 忽略审计日志记录失败的错误，导致审计功能静默失败
- **影响范围/风险**：
  - 数据库连接失败、磁盘满等情况下，审计日志丢失且无告警
  - 违反审计日志的完整性要求（审计日志应该是可靠的）
  - 无法追溯审计功能是否正常工作
- **修正建议**：
  ```go
  func Audit(logger AuditLogger, action, resource string, log *zap.Logger) gin.HandlerFunc {
      return func(c *gin.Context) {
          c.Next()

          if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
              user := authctx.MustCurrentUser(c)
              if user.UserID == 0 {
                  return
              }

              ip := c.ClientIP()
              if err := logger.Log(user.UserID, action, resource, ip); err != nil {
                  // 记录错误但不影响业务流程
                  log.Error("failed to log audit",
                      zap.Int64("user_id", user.UserID),
                      zap.String("action", action),
                      zap.String("resource", resource),
                      zap.Error(err),
                  )
              }
          }
      }
  }
  ```

### 🟡 中优先级

#### [M1] 缺少分页参数校验

- **文件**：`code/backend/internal/module/audit/handler.go:22-23`
- **问题描述**：
  ```go
  filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
  filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
  ```
  忽略 `strconv.Atoi` 的错误，且未校验分页参数的合法性（如 page < 1、pageSize > 1000）
- **影响范围/风险**：
  - 恶意用户可以传入超大 pageSize（如 999999）导致数据库查询性能问题
  - 负数或零值的 page 会导致 SQL offset 计算错误
- **修正建议**：
  ```go
  func (h *Handler) ListAuditLogs(c *gin.Context) {
      var filter ListFilter

      page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
      if err != nil || page < 1 {
          page = 1
      }
      filter.Page = page

      pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
      if err != nil || pageSize < 1 || pageSize > 100 {
          pageSize = 20
      }
      filter.PageSize = pageSize

      // ... 其余逻辑
  }
  ```

#### [M2] Repository 查询存在 SQL 注入风险

- **文件**：`code/backend/internal/module/audit/repository.go:46`
- **问题描述**：
  ```go
  query = query.Where("resource LIKE ?", "%"+filter.Resource+"%")
  ```
  虽然使用了参数化查询，但 LIKE 模糊查询未对特殊字符（`%`、`_`）进行转义，用户输入 `%` 会匹配所有记录
- **影响范围/风险**：
  - 用户可以通过输入 `%` 绕过过滤条件
  - 性能问题：`%xxx%` 的 LIKE 查询无法使用索引
- **修正建议**：
  ```go
  if filter.Resource != "" {
      // 转义 LIKE 特殊字符
      escaped := strings.ReplaceAll(filter.Resource, "%", "\\%")
      escaped = strings.ReplaceAll(escaped, "_", "\\_")
      query = query.Where("resource LIKE ?", "%"+escaped+"%")
  }
  ```

  或者使用精确匹配：
  ```go
  if filter.Resource != "" {
      query = query.Where("resource = ?", filter.Resource)
  }
  ```

#### [M3] 缺少审计日志查询的权限校验

- **文件**：`code/backend/internal/app/router.go:95`
- **问题描述**：
  ```go
  adminOnly.GET("/audit-logs", auditHandler.ListAuditLogs)
  ```
  虽然限制了管理员访问，但未校验管理员是否可以查询其他管理员的操作日志（可能涉及敏感操作）
- **影响范围/风险**：
  - 普通管理员可以查看超级管理员的操作记录
  - 缺少审计日志的访问审计（谁查看了审计日志）
- **修正建议**：
  1. 如果需要限制普通管理员只能查看自己的日志：
  ```go
  func (h *Handler) ListAuditLogs(c *gin.Context) {
      user := authctx.MustCurrentUser(c)

      var filter ListFilter
      // ... 解析参数 ...

      // 非超级管理员只能查看自己的日志
      if user.Role != model.RoleSuperAdmin {
          filter.UserID = &user.UserID
      }

      // ... 其余逻辑
  }
  ```

  2. 添加"查看审计日志"操作的审计记录：
  ```go
  adminOnly.GET("/audit-logs",
      middleware.Audit(auditService, "view", "audit-logs"),
      auditHandler.ListAuditLogs)
  ```

#### [M4] 缺少时间范围查询

- **文件**：`code/backend/internal/module/audit/repository.go:36-47`
- **问题描述**：
  ListFilter 缺少时间范围过滤（start_time、end_time），审计日志查询通常需要按时间范围筛选
- **影响范围/风险**：
  - 无法按时间范围查询历史审计记录
  - 随着数据量增长，查询性能下降
- **修正建议**：
  ```go
  type ListFilter struct {
      UserID    *int64
      Action    string
      Resource  string
      StartTime *time.Time
      EndTime   *time.Time
      Page      int
      PageSize  int
  }

  func (r *repository) List(ctx context.Context, filter ListFilter) ([]*model.AuditLog, int64, error) {
      query := r.db.WithContext(ctx).Model(&model.AuditLog{})

      // ... 其他过滤条件 ...

      if filter.StartTime != nil {
          query = query.Where("created_at >= ?", *filter.StartTime)
      }
      if filter.EndTime != nil {
          query = query.Where("created_at <= ?", *filter.EndTime)
      }

      // ... 其余逻辑
  }
  ```

#### [M5] Service.Log 方法使用 context.Background() 不合理

- **文件**：`code/backend/internal/module/audit/service.go:31`
- **问题描述**：
  ```go
  return s.repo.Create(context.Background(), log)
  ```
  使用 `context.Background()` 而不是传入的 context，导致无法传播超时、取消信号和 trace 信息
- **影响范围/风险**：
  - 无法通过 context 控制审计日志写入的超时
  - 分布式追踪链路断裂
- **修正建议**：
  ```go
  func (s *service) Log(ctx context.Context, userID int64, action, resource, ip string) error {
      log := &model.AuditLog{
          UserID:    userID,
          Action:    action,
          Resource:  resource,
          IP:        ip,
          CreatedAt: time.Now(),
      }
      return s.repo.Create(ctx, log)
  }
  ```

  同时修改接口定义和中间件调用：
  ```go
  type AuditLogger interface {
      Log(ctx context.Context, userID int64, action, resource, ip string) error
  }

  // middleware/audit.go
  _ = logger.Log(c.Request.Context(), user.UserID, action, resource, ip)
  ```

### 🟢 低优先级

#### [L1] 缺少注册操作的审计

- **文件**：`code/backend/internal/app/router.go:79`
- **问题描述**：
  ```go
  authGroup.POST("/register", authHandler.Register)
  ```
  注册操作未添加审计日志，无法追溯新用户的注册来源
- **影响范围/风险**：
  - 无法追踪恶意注册行为
  - 缺少用户注册的 IP 记录
- **修正建议**：
  在 `authHandler.Register` 内部添加审计日志：
  ```go
  func (h *Handler) Register(c *gin.Context) {
      // ... 注册逻辑 ...

      // 注册成功后记录审计日志
      if err := h.auditService.Log(c.Request.Context(), user.ID, "register", "auth/register", c.ClientIP()); err != nil {
          h.log.Error("failed to log audit", zap.Error(err))
      }

      response.Success(c, userResp)
  }
  ```

#### [L2] 缺少登出操作的审计

- **文件**：`code/backend/internal/app/router.go:85`
- **问题描述**：
  ```go
  protected.POST("/auth/logout", authHandler.Logout)
  ```
  登出操作未添加审计日志
- **影响范围/风险**：
  - 无法追踪用户会话的完整生命周期（登录-操作-登出）
- **修正建议**：
  ```go
  protected.POST("/auth/logout",
      middleware.Audit(auditService, "logout", "auth/logout"),
      authHandler.Logout)
  ```

#### [L3] 数据库迁移文件缺少外键约束

- **文件**：`code/backend/migrations/000010_create_audit_logs_table.up.sql:3`
- **问题描述**：
  ```sql
  user_id BIGINT NOT NULL,
  ```
  `user_id` 字段未添加外键约束到 `users` 表，可能导致数据不一致（用户删除后审计日志仍存在孤立记录）
- **影响范围/风险**：
  - 用户删除后，审计日志中的 user_id 指向不存在的用户
  - 查询时需要额外处理用户不存在的情况
- **修正建议**：
  审计日志通常不应该因为用户删除而删除（需要保留历史记录），因此不建议添加 `ON DELETE CASCADE`。但可以考虑：

  1. 不添加外键约束（当前做法可接受）
  2. 或者在查询时 LEFT JOIN users 表，显示已删除用户为 "已删除用户"

#### [L4] 缺少审计日志的数据保留策略

- **文件**：无
- **问题描述**：
  审计日志会无限增长，缺少数据归档或清理策略
- **影响范围/风险**：
  - 长期运行后表数据量过大，影响查询性能
  - 存储成本增加
- **修正建议**：
  1. 在文档中说明审计日志的保留策略（如保留 1 年）
  2. 添加定时任务清理过期数据：
  ```go
  // 定期归档或删除 1 年前的审计日志
  DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '1 year';
  ```
  3. 或者使用数据库分区表按月分区

#### [L5] IP 字段长度可能不足

- **文件**：`code/backend/migrations/000010_create_audit_logs_table.up.sql:6`
- **问题描述**：
  ```sql
  ip VARCHAR(45) NOT NULL,
  ```
  IPv6 地址最长为 45 字符（如 `ffff:ffff:ffff:ffff:ffff:ffff:255.255.255.255`），但如果经过代理，可能包含多个 IP（X-Forwarded-For），长度不足
- **影响范围/风险**：
  - 代理环境下 IP 地址可能被截断
- **修正建议**：
  ```sql
  ip VARCHAR(255) NOT NULL,
  ```

#### [L6] 缺少操作结果字段

- **文件**：`code/backend/internal/model/audit_log.go:13-20`
- **问题描述**：
  AuditLog 模型缺少操作结果字段（成功/失败），当前只记录成功的操作（`c.Writer.Status() >= 200 && c.Writer.Status() < 300`）
- **影响范围/风险**：
  - 无法追踪失败的操作尝试（如登录失败、权限不足等）
  - 无法分析攻击行为（如暴力破解）
- **修正建议**：
  1. 在 Model 中添加字段：
  ```go
  type AuditLog struct {
      // ... 现有字段 ...
      Status    int    `gorm:"column:status"`        // HTTP 状态码
      Success   bool   `gorm:"column:success;index"` // 是否成功
      ErrorMsg  string `gorm:"column:error_msg"`     // 错误信息（可选）
  }
  ```

  2. 修改中间件记录所有操作（包括失败）：
  ```go
  func Audit(logger AuditLogger, action, resource string) gin.HandlerFunc {
      return func(c *gin.Context) {
          c.Next()

          user := authctx.MustCurrentUser(c)
          if user.UserID == 0 {
              return
          }

          status := c.Writer.Status()
          success := status >= 200 && status < 300

          _ = logger.LogWithStatus(user.UserID, action, resource, c.ClientIP(), status, success)
      }
  }
  ```

#### [L7] 缺少单元测试

- **文件**：无
- **问题描述**：
  审计模块缺少单元测试，无法验证功能正确性
- **影响范围/风险**：
  - 重构时可能引入 bug
  - 无法验证边界条件（如分页、过滤、权限）
- **修正建议**：
  添加测试文件：
  - `code/backend/internal/module/audit/repository_test.go`
  - `code/backend/internal/module/audit/service_test.go`
  - `code/backend/internal/module/audit/handler_test.go`
  - `code/backend/internal/middleware/audit_test.go`

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 3 |
| 🟡 中 | 5 |
| 🟢 低 | 7 |
| 合计 | 15 |

## 总体评价

审计日志功能的核心逻辑已实现，包括中间件、Repository、Service、Handler 分层清晰，数据库索引设计合理。但存在以下主要问题：

1. **架构一致性问题**：最严重的是 Model 和 DTO 未分离（H1），违反了项目的核心架构规范，必须修复
2. **功能缺陷**：登录审计的实现方式错误（H2），导致功能完全失效
3. **可靠性问题**：审计日志记录失败被忽略（H3），违反审计系统的可靠性要求
4. **安全性问题**：缺少分页校验（M1）、LIKE 查询未转义（M2）、权限校验不足（M3）

建议优先修复所有高优先级和中优先级问题后再合并到主分支。低优先级问题可以在后续迭代中逐步完善。
