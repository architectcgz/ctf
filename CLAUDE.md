# CTF 平台开发规范

> 本文件对所有 agent（Kiro、Codex 等）生效

## 后端代码规范

### 1. 模型分层（强制）

**Model 和 DTO 必须分离，禁止共用**

```go
// ✅ 正确：Model 用于数据库
// internal/model/user.go
type User struct {
    ID           int64  `gorm:"primaryKey"`
    Username     string `gorm:"uniqueIndex"`
    PasswordHash string `gorm:"column:password_hash"` // 敏感字段
    Email        string
    Role         string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// ✅ 正确：DTO 用于 API
// internal/dto/auth.go
type UserResp struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    // 注意：不包含 PasswordHash
}

// ❌ 错误：直接返回 Model
func (h *Handler) GetUser(c *gin.Context) {
    user, _ := h.repo.FindByID(id)
    c.JSON(200, user) // 会泄漏 PasswordHash
}
```

### 2. 分层职责（强制）

**Repository → Service → Handler 严格分层**

```go
// Repository 层：操作 Model，返回 Model
type UserRepository interface {
    Create(user *model.User) error
    FindByID(id int64) (*model.User, error)
}

// Service 层：接收 DTO，返回 DTO，内部转换 Model
type AuthService interface {
    Register(req *dto.RegisterReq) (*dto.UserResp, error)
    GetUser(id int64) (*dto.UserResp, error)
}

func (s *AuthService) GetUser(id int64) (*dto.UserResp, error) {
    // 1. Repository 返回 Model
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // 2. Model → DTO 转换
    return toUserResp(user), nil
}

// 转换函数：放在 Service 层
func toUserResp(user *model.User) *dto.UserResp {
    return &dto.UserResp{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        Role:     user.Role,
    }
}

// Handler 层：只处理 DTO
func (h *AuthHandler) GetProfile(c *gin.Context) {
    userID := c.GetInt64("user_id")

    // 调用 Service，接收 DTO
    userResp, err := h.authService.GetUser(userID)
    if err != nil {
        response.FromError(c, err)
        return
    }

    // 返回 DTO
    response.Success(c, userResp)
}
```

### 3. 安全规范（强制）

- **密码**：必须 bcrypt 加密，只存储哈希值
- **敏感字段**：PasswordHash、Token、Salt 等禁止出现在 DTO
- **JWT**：使用 RS256 签名，禁止 HS256（生产环境）
- **输入校验**：所有外部输入必须使用 binding 标签校验
- **SQL 注入**：使用 GORM 参数化查询，禁止字符串拼接

### 4. 错误处理（强制）

```go
// ✅ 正确：使用统一错误码
if user == nil {
    return nil, errcode.ErrNotFound("用户")
}

// ✅ 正确：Service 返回业务错误
func (s *Service) DoSomething() error {
    if condition {
        return errcode.ErrForbidden()
    }
    return nil
}

// ✅ 正确：Handler 统一处理错误
func (h *Handler) Handle(c *gin.Context) {
    err := h.service.DoSomething()
    if err != nil {
        response.FromError(c, err) // 自动识别错误类型
        return
    }
    response.Success(c, data)
}

// ❌ 错误：直接返回 error 字符串
return errors.New("用户不存在")

// ❌ 错误：Handler 直接返回 HTTP 状态码
c.JSON(404, gin.H{"error": "not found"})
```

### 5. 响应格式（强制）

**所有接口必须使用统一响应封装**

```go
// ✅ 正确：使用 response 包
response.Success(c, data)
response.Error(c, errcode.ErrNotFound("资源"))
response.Page(c, list, total, page, pageSize)

// ❌ 错误：直接使用 c.JSON
c.JSON(200, gin.H{"data": data})
```

### 6. 数据库操作（强制）

```go
// ✅ 正确：Repository 封装 GORM
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
    var user model.User
    err := r.db.Where("id = ?", id).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// ❌ 错误：Service 直接操作 GORM
func (s *Service) GetUser(id int64) (*dto.UserResp, error) {
    var user model.User
    s.db.First(&user, id) // 不应该在 Service 层直接用 db
    return toUserResp(&user), nil
}
```

### 7. 运行时容器操作（强制）

**所有运行时容器操作必须通过 runtime 模块，禁止业务模块直接调用 Docker SDK 或 `runtimeinfra`**

```go
// ✅ 正确：通过 runtime.RuntimeFacade
containerID, networkID, hostPort, servicePort, err := s.runtimeService.CreateContainer(ctx, imageRef, env, reservedHostPort)

// ❌ 错误：业务模块直接使用 Docker SDK
container, err := dockerClient.ContainerCreate(...)
```

### 8. 命名规范

- **文件名**：snake_case（如 `user_service.go`）
- **包名**：小写单词（如 `auth`, `challenge`）
- **接口名**：大写开头 + 名词（如 `UserRepository`, `AuthService`）
- **方法名**：大写开头 + 动词（如 `CreateUser`, `FindByID`）
- **变量名**：驼峰命名（如 `userID`, `challengeList`）

### 9. 代码检查清单

编写代码时必须检查：

- [ ] Model 和 DTO 是否分离？
- [ ] Repository 是否返回 Model？
- [ ] Service 是否返回 DTO？
- [ ] Handler 是否只处理 DTO？
- [ ] 敏感字段是否泄漏到 API？
- [ ] 是否使用统一响应格式？
- [ ] 是否使用统一错误码？
- [ ] 输入是否校验？
- [ ] 是否有 SQL 注入风险？
- [ ] 运行时容器操作是否通过 runtime 模块？

## 前端代码规范

（待补充）

## Review 要点

Code reviewer 重点检查：

1. **架构一致性**：是否遵循分层架构
2. **安全性**：敏感字段、SQL 注入、XSS
3. **资源泄漏**：容器、网络、文件句柄是否清理
4. **错误处理**：是否完整、是否使用统一错误码
5. **性能**：是否有 N+1 查询、是否需要缓存

---

> 所有 agent 在编写代码前必须阅读本规范
> 违反规范的代码将在 code review 中被拒绝
