# CTF Backend 代码 Review（instance-start 第 1 轮）：靶机实例启动功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | instance-start |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit b99727e，6 个文件，285 行增加 / 18 行删除 |
| 变更概述 | 实现靶机实例启动功能（B18），包括并发限制、动态 Flag 生成、容器创建流程 |
| 审查基准 | `docs/tasks/backend-task-breakdown.md` B18 任务定义 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | N/A（首次审查） |

## 问题清单

### 🔴 高优先级

#### [H1] 容器创建逻辑未实际实现，仅为模拟代码

- **文件**：`code/backend/internal/module/practice/service.go:108-127`
- **问题描述**：`createContainer` 方法中使用 `TODO` 注释和模拟实现，未调用实际的 Docker SDK 或 container 模块
- **影响范围/风险**：
  - 容器实际未创建，实例启动功能完全不可用
  - 生成的 ContainerID、NetworkID、AccessURL 都是假数据
  - 动态 Flag 无法注入到容器环境变量中
  - 违反任务 B18 的核心交付要求
- **修正建议**：
  ```go
  func (s *Service) createContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, flag string) error {
      // 1. 构建容器配置
      containerConfig := &container.ContainerConfig{
          Image: chal.ImageName, // 需要从 Challenge 获取镜像名
          Env: map[string]string{
              "FLAG": flag, // 注入 Flag
          },
          Resources: container.ResourceLimits{
              CPUQuota: s.config.Container.DefaultCPUQuota,
              Memory:   s.config.Container.DefaultMemory,
              PidsLimit: s.config.Container.DefaultPidsLimit,
          },
      }

      // 2. 调用 container.Service 创建容器
      containerID, networkID, port, err := s.containerService.CreateAndStartInstance(ctx, containerConfig)
      if err != nil {
          return errcode.ErrContainerCreateFailed.WithCause(err)
      }

      // 3. 更新实例信息
      instance.ContainerID = containerID
      instance.NetworkID = networkID
      instance.AccessURL = fmt.Sprintf("http://%s:%d", s.config.Container.PublicHost, port)

      return nil
  }
  ```

#### [H2] 缺少 container.Service 依赖注入

- **文件**：`code/backend/internal/module/practice/service.go:18-37`
- **问题描述**：Service 结构体中没有 `containerService` 字段，无法调用容器模块的功能
- **影响范围/风险**：
  - 无法创建、启动、停止、删除容器
  - 无法创建网络和端口映射
  - 违反分层架构原则（应通过 container 模块操作容器）
- **修正建议**：
  ```go
  type Service struct {
      challengeRepo    *challenge.Repository
      instanceRepo     *container.Repository
      containerService *container.Service  // 新增
      config           *config.Config
      logger           *zap.Logger
  }

  func NewService(
      challengeRepo *challenge.Repository,
      instanceRepo *container.Repository,
      containerService *container.Service,  // 新增参数
      config *config.Config,
      logger *zap.Logger,
  ) *Service {
      return &Service{
          challengeRepo:    challengeRepo,
          instanceRepo:     instanceRepo,
          containerService: containerService,  // 注入
          config:           config,
          logger:           logger,
      }
  }
  ```

#### [H3] router.go 中未初始化 container.Service

- **文件**：`code/backend/internal/app/router.go:112-115`
- **问题描述**：创建 practiceService 时缺少 containerService 参数
- **影响范围/风险**：编译错误或运行时 panic（如果 H2 修复后）
- **修正建议**：
  ```go
  // 容器模块
  containerService := containerModule.NewService(cfg, log.Named("container_service"))

  // 实践模块（学员）
  instanceRepo := containerModule.NewRepository(db)
  practiceService := practiceModule.NewService(
      challengeRepo,
      instanceRepo,
      containerService,  // 新增
      cfg,
      log.Named("practice_service"),
  )
  ```

#### [H4] 动态 Flag 生成使用错误的 salt 参数

- **文件**：`code/backend/internal/module/practice/service.go:66`
- **问题描述**：调用 `crypto.GenerateDynamicFlag` 时传入 `chal.FlagSalt`，但根据 `pkg/crypto/flag.go:14` 的签名，第三个参数应该是 `globalSecret`，第四个才是 `nonce`
- **影响范围/风险**：
  - Flag 生成算法错误，无法验证
  - 每个靶场使用不同的 salt 作为密钥，降低安全性
  - 与 Flag 验证逻辑不一致
- **修正建议**：
  ```go
  // 从环境变量或配置获取全局密钥
  globalSecret := s.config.Container.FlagGlobalSecret
  if globalSecret == "" {
      return nil, errcode.ErrInternal.WithCause(errors.New("Flag 全局密钥未配置"))
  }
  flag = crypto.GenerateDynamicFlag(userID, challengeID, globalSecret, nonce)
  ```

#### [H5] 缺少容器创建失败后的资源清理

- **文件**：`code/backend/internal/module/practice/service.go:87-91`
- **问题描述**：容器创建失败时只更新了实例状态为 `failed`，但未清理可能已创建的部分资源（网络、端口占用、数据库记录）
- **影响范围/风险**：
  - 资源泄漏（网络、端口）
  - 数据库中留下大量 failed 状态的脏数据
  - 端口可能被占用但容器未运行
- **修正建议**：
  ```go
  if err := s.createContainer(ctx, instance, chal, flag); err != nil {
      s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))

      // 清理资源
      if instance.NetworkID != "" {
          s.containerService.RemoveNetwork(instance.NetworkID)
      }
      if instance.ContainerID != "" {
          s.containerService.RemoveContainer(instance.ContainerID)
      }

      // 删除数据库记录（或标记为 failed）
      s.instanceRepo.UpdateStatus(instance.ID, model.InstanceStatusFailed)

      return nil, err
  }
  ```

#### [H6] 并发限制检查存在竞态条件

- **文件**：`code/backend/internal/module/practice/service.go:41-47`
- **问题描述**：先查询实例数量，再创建新实例，两个操作之间没有加锁，存在 TOCTOU（Time-of-check to time-of-use）竞态
- **影响范围/风险**：
  - 并发请求可能绕过限制，创建超过上限的实例
  - 例如：用户同时发起 3 个请求，都通过了检查，最终创建了 3 个实例（假设上限是 3）
- **修正建议**：
  ```go
  // 方案 1：使用 Redis 分布式锁
  lockKey := fmt.Sprintf("instance:create:user:%d", userID)
  lock, err := s.redisClient.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
  if err != nil || !lock {
      return nil, errcode.ErrTooManyRequests
  }
  defer s.redisClient.Del(ctx, lockKey)

  // 方案 2：使用数据库乐观锁 + 唯一约束
  // 在 instances 表添加唯一索引：user_id + status (where status in ('creating', 'running'))
  // 依赖数据库约束保证并发安全
  ```

### 🟡 中优先级

#### [M1] Challenge 模型缺少 ImageName 字段

- **文件**：`code/backend/internal/module/practice/service.go:120`
- **问题描述**：代码注释中提到 `chal.ImageName`，但根据之前的 Challenge 模型定义，该字段不存在（只有 ImageID）
- **影响范围/风险**：编译错误或需要额外查询 Image 表
- **修正建议**：
  ```go
  // 方案 1：通过 ImageID 查询镜像信息
  image, err := s.imageRepo.FindByID(chal.ImageID)
  if err != nil {
      return errcode.ErrInternal.WithCause(err)
  }
  containerConfig.Image = fmt.Sprintf("%s:%s", image.Name, image.Tag)

  // 方案 2：在 Challenge 模型中添加冗余字段 ImageFullName
  ```

#### [M2] 缺少静态 Flag 的处理逻辑

- **文件**：`code/backend/internal/module/practice/service.go:60-67`
- **问题描述**：只处理了动态 Flag 类型，静态 Flag 类型未处理，flag 变量为空字符串
- **影响范围/风险**：
  - 静态 Flag 类型的靶场无法启动（容器内没有 Flag）
  - 违反任务 B18 要求（需支持静态和动态两种类型）
- **修正建议**：
  ```go
  var flag string
  var nonce string

  if chal.FlagType == model.FlagTypeDynamic {
      nonce, err = crypto.GenerateNonce()
      if err != nil {
          return nil, errcode.ErrInternal.WithCause(err)
      }
      globalSecret := s.config.Container.FlagGlobalSecret
      if globalSecret == "" {
          return nil, errcode.ErrInternal.WithCause(errors.New("Flag 全局密钥未配置"))
      }
      flag = crypto.GenerateDynamicFlag(userID, challengeID, globalSecret, nonce)
  } else if chal.FlagType == model.FlagTypeStatic {
      // 静态 Flag 从数据库读取（已加密存储）
      // 注意：不要直接返回明文 Flag，应该在容器内生成或预置
      flag = chal.StaticFlag // 假设 Challenge 模型有此字段
  }
  ```

#### [M3] 端口分配算法过于简单，存在冲突风险

- **文件**：`code/backend/internal/module/practice/service.go:123`
- **问题描述**：使用 `instance.ID % 端口范围` 计算端口，可能导致端口冲突（多个实例映射到同一端口）
- **影响范围/风险**：
  - 端口冲突导致容器启动失败
  - 端口分配不均匀
- **修正建议**：
  ```go
  // 应该调用 container.Service 的端口分配方法
  port, err := s.containerService.AllocatePort()
  if err != nil {
      return errcode.ErrContainerCreateFailed.WithCause(err)
  }
  instance.AccessURL = fmt.Sprintf("http://%s:%d", s.config.Container.PublicHost, port)
  ```

#### [M4] 缺少实例重复创建检测

- **文件**：`code/backend/internal/module/practice/service.go:40-57`
- **问题描述**：未检查用户是否已有该靶场的运行中实例，可能重复创建
- **影响范围/风险**：
  - 用户可以为同一靶场创建多个实例，浪费资源
  - 违反常见 CTF 平台的业务规则（一个用户同时只能有一个靶场实例）
- **修正建议**：
  ```go
  // 检查是否已有该靶场的运行中实例
  existingInstance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
  if err == nil && existingInstance != nil {
      // 返回已有实例
      return toInstanceResp(existingInstance), nil
  }
  ```

#### [M5] 超时时间硬编码

- **文件**：`code/backend/internal/module/practice/service.go:85`
- **问题描述**：容器创建超时时间硬编码为 30 秒，应该从配置读取
- **影响范围/风险**：违反配置外部化原则
- **修正建议**：
  ```go
  // 在 config.ContainerConfig 中添加字段
  type ContainerConfig struct {
      // ...
      CreateTimeout time.Duration `mapstructure:"create_timeout"`
  }

  // 使用配置
  ctx, cancel := context.WithTimeout(context.Background(), s.config.Container.CreateTimeout)
  ```

#### [M6] InstanceInfo 缺少 ChallengeName 字段填充

- **文件**：`code/backend/internal/module/practice/service.go:167-181`
- **问题描述**：`toInstanceInfo` 函数中 `ChallengeName` 字段未填充，但 DTO 定义中有此字段
- **影响范围/风险**：前端无法显示靶场名称，用户体验差
- **修正建议**：
  ```go
  func (s *Service) ListUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
      instances, err := s.instanceRepo.FindByUserID(userID)
      if err != nil {
          return nil, errcode.ErrInternal.WithCause(err)
      }

      // 批量查询靶场信息
      challengeIDs := make([]int64, len(instances))
      for i, inst := range instances {
          challengeIDs[i] = inst.ChallengeID
      }
      challenges, _ := s.challengeRepo.FindByIDs(challengeIDs) // 需要实现此方法
      challengeMap := make(map[int64]*model.Challenge)
      for _, c := range challenges {
          challengeMap[c.ID] = c
      }

      result := make([]*dto.InstanceInfo, len(instances))
      for i, inst := range instances {
          result[i] = toInstanceInfo(inst)
          if chal, ok := challengeMap[inst.ChallengeID]; ok {
              result[i].ChallengeName = chal.Title
          }
      }
      return result, nil
  }
  ```

### 🟢 低优先级

#### [L1] 日志记录不完整

- **文件**：`code/backend/internal/module/practice/service.go:88, 100-103`
- **问题描述**：只在失败和成功时记录日志，缺少关键步骤的日志（如并发限制触发、Flag 生成）
- **影响范围/风险**：问题排查困难，缺少审计信息
- **修正建议**：
  ```go
  // 并发限制触发
  if len(instances) >= s.config.Container.MaxConcurrentPerUser {
      s.logger.Warn("用户实例数量超限",
          zap.Int64("user_id", userID),
          zap.Int("current", len(instances)),
          zap.Int("limit", s.config.Container.MaxConcurrentPerUser))
      return nil, errcode.ErrInstanceLimitExceeded
  }

  // Flag 生成
  s.logger.Debug("生成动态 Flag",
      zap.Int64("user_id", userID),
      zap.Int64("challenge_id", challengeID),
      zap.String("nonce", nonce))
  ```

#### [L2] 错误处理可以更精细

- **文件**：`code/backend/internal/module/practice/service.go:50-56`
- **问题描述**：查询靶场失败时统一返回 `ErrNotFound`，未区分数据库错误和记录不存在
- **影响范围/风险**：数据库故障时返回 404 而非 500，误导用户
- **修正建议**：
  ```go
  chal, err := s.challengeRepo.FindByID(challengeID)
  if err != nil {
      if err == gorm.ErrRecordNotFound {
          return nil, errcode.ErrNotFound
      }
      return nil, errcode.ErrInternal.WithCause(err)
  }
  ```

#### [L3] GetInstance 权限校验可以提取为中间件

- **文件**：`code/backend/internal/module/practice/service.go:130-137`
- **问题描述**：在 Service 层校验 `instance.UserID != userID`，这类权限校验更适合在中间件或 Handler 层处理
- **影响范围/风险**：职责混乱，Service 层应该专注业务逻辑
- **修正建议**：
  ```go
  // 方案 1：在 Handler 层校验
  func (h *Handler) GetInstance(c *gin.Context) {
      userID := c.GetInt64("user_id")
      instanceID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

      instance, err := h.service.GetInstance(instanceID)
      if err != nil {
          response.FromError(c, err)
          return
      }

      // 权限校验
      if instance.UserID != userID {
          response.Error(c, errcode.ErrForbidden)
          return
      }

      response.Success(c, instance)
  }

  // 方案 2：使用 RBAC 中间件 + 资源所有权校验
  ```

#### [L4] 模拟延迟代码应该删除

- **文件**：`code/backend/internal/module/practice/service.go:113-116`
- **问题描述**：`time.After(100 * time.Millisecond)` 是测试代码，不应该出现在生产代码中
- **影响范围/风险**：无意义的延迟，降低性能
- **修正建议**：删除此段代码

#### [L5] flag_handler.go 和 flag_service.go 的错误处理改动不一致

- **文件**：`code/backend/internal/module/challenge/flag_handler.go:26,30,42,55,61`
- **问题描述**：将 `response.Error(c, err)` 改为 `response.Error(c, errcode.ErrInvalidParams)` 或 `response.ValidationError(c, err)`，但这些改动与本次 B18 任务无关
- **影响范围/风险**：混入无关改动，违反单一职责原则
- **修正建议**：将这些改动拆分到独立的 commit 或 PR

#### [L6] test_helper.go 注释掉的代码应该删除

- **文件**：`code/backend/internal/module/challenge/test_helper.go:20-28`
- **问题描述**：注释掉的 `setupTagTestDB` 函数应该删除，而不是保留注释
- **影响范围/风险**：代码冗余，降低可读性
- **修正建议**：删除注释代码，或者在 Tag 模型实现后再添加

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 6 |
| 🟡 中 | 6 |
| 🟢 低 | 6 |
| 合计 | 18 |

## 总体评价

本次实现完成了 B18 任务的基本框架，包括：
- ✅ 并发限制检查（存在竞态风险）
- ✅ 动态 Flag 生成逻辑（参数错误）
- ✅ 实例记录创建
- ✅ 超时控制
- ✅ Handler/Service/Repository 分层正确
- ✅ DTO 和 Model 分离

但存在以下严重问题：
- ❌ **容器创建逻辑完全未实现**（H1，阻塞性问题）
- ❌ **缺少 container.Service 依赖**（H2-H3，架构缺陷）
- ❌ **动态 Flag 生成参数错误**（H4，功能缺陷）
- ❌ **资源清理不完整**（H5，资源泄漏风险）
- ❌ **并发限制存在竞态条件**（H6，安全风险）

**建议**：
1. 优先修复 H1-H6 高优先级问题，这些问题会导致功能完全不可用或存在严重风险
2. 补充 container.Service 的集成，调用实际的 Docker SDK
3. 添加完整的资源清理逻辑和错误处理
4. 补充单元测试和集成测试
5. 修复后进行第 2 轮 code review

**不可合并**：当前代码存在阻塞性问题（H1），不建议合并到主分支。
